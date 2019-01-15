package pooller

import (
	"time"
	"github.com/patrickmn/go-cache"
	"github.com/satori/go.uuid"
	"../logger"
	"fmt"
	"log"
	"os"
)


func GetDefaultConfiguration() InitWorkerConfiguration {
	return InitWorkerConfiguration{
	  CacheExpiration: time.Second * 60,
	  CachePurge: time.Second * 3,
	  CacheRemoteResponseCacheTimeout: time.Second * 120,
	  DefaultSnmpRepeats: 5,
	  DefaultSnmpTimeout: 3,
	  LimitCountWorkers: 100,
	  LimitOneDevice: 3,
	  LimitOneRequest: 50,
	  LimitRequestResetTimeout: 300,
	  LimitResponseCollectorCount: 10,
	}
}


func New(config InitWorkerConfiguration) *Worker {
	log.Println("Initialize requestPooler")
	worker := new(Worker)

	worker.Config = config
	worker.cache = cache.New(config.CacheRemoteResponseCacheTimeout, config.CachePurge)
	worker.requestCollector = cache.New(time.Duration(config.LimitRequestResetTimeout + 60) * time.Second, config.CachePurge)
	worker.limitCountForSwitch = cache.New(time.Duration(config.LimitRequestResetTimeout + 60) * time.Second, config.CachePurge)
	worker.limitCountForRequest = cache.New(time.Second * 900, config.CachePurge)
	worker.requestQueue = make(chan Pooller, config.LimitCountWorkers)
	worker.responseQueue = make(chan Pooller, config.LimitResponseCollectorCount)
	for i := 0; i <= config.LimitCountWorkers; i++ {
		log.Println("Start snmp worker")
		go backgroundWorker(worker, i)
	}
	for i := 0; i <= config.LimitCountWorkers; i++ {
		log.Println("Start response collector")
		go workerResponseCollector(worker, i)
	}
	if worker.Logger == nil {
		worker.Logger, _ = logger.New("pooler", 0, os.DevNull)
	}
	return  worker
}

func (w *Worker) Get(requests []Request) []Response {
	return w.addToPool(requests, requestGet)
}

func (w *Worker) addToPool( requests []Request, requestType RequestType) []Response {
	w.Logger.DebugF("New pool add request, length %v with type", len(requests))
	//Generate requestId
	requestUUid := ""
	if  uuid, err :=  uuid.NewV4(); err == nil {
		requestUUid = uuid.String()
	}

	//Rebuild request for work with as map
	requestMapped := make(map[string]Pooller)
	for _, r := range  requests  {
		if r.Timeout == 0 {
			r.Timeout = w.Config.DefaultSnmpTimeout
		}
		if r.Repeats == 0 {
			r.Repeats = w.Config.DefaultSnmpRepeats
		}
		requestMapped[fmt.Sprintf("%v:%v", r.Ip, r.Oid)] = Pooller{
			UUid:         requestUUid,
			RequestBody:  r,
			ResponseBody: Response{},
			Type: requestType,
		}
	}

	//Sending requests to pool
	for {
		if len(requestMapped) == 0 {
			break
		}
		for keyName, poolItem := range requestMapped {
			if count := w.addCountFromRequest(poolItem.UUid); count >= w.Config.LimitOneRequest {
				continue
			}
			if count := w.addRequestForSwitch(poolItem.RequestBody.Ip); count >= w.Config.LimitOneDevice {
				continue
			}
			w.requestQueue <- poolItem
			delete(requestMapped, keyName)
		}
	}
	startWaitingResponse := time.Now().Unix()
	response := make([]Response, 0)
	for {
		poollers := w.getRequestData(requestUUid)
		if len(poollers) >= len(requests) {
			for _, pool := range poollers {
				response = append(response, pool.ResponseBody)
			}
			break
		}
		if (time.Now().Unix() - startWaitingResponse) >  int64(w.Config.LimitRequestResetTimeout) {
			log.Println("Timeout waiting response for request", requestUUid)
			for _, pool := range poollers {
				response = append(response, pool.ResponseBody)
			}
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	w.delRequestData(requestUUid)
	return  response
}


func (w *Worker) GetStatus()  StatusPooler {
	status := StatusPooler{
		CountRequestQueue: len(w.requestQueue),
		CountResponseQueue: len(w.responseQueue),
	}
	status.CountWorkersForSw = make(map[string]int)
	status.CountWorkersForRequest = make(map[string]int)

	for name, count := range  w.limitCountForSwitch.Items()  {
		status.CountWorkersForSw[name] = count.Object.(int)
	}
	for name, count := range  w.limitCountForRequest.Items()  {
		status.CountWorkersForRequest[name] = count.Object.(int)
	}
	return  status
}


func (w *Worker) Walk(r []Request) []Response {
	return w.addToPool(r, requestWalk)
}
func (w *Worker) BulkWalk(r []Request) []Response {
	return w.addToPool(r, requestBulkWalk)
}
func (w *Worker) Set(r []Request)  []Response {
	return w.addToPool(r, requestGet)
}
