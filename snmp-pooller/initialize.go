package pooller

import (
	"time"
	"github.com/patrickmn/go-cache"
	"github.com/satori/go.uuid"
	"fmt"
	"log"
)


func GetDefaultConfiguration() InitWorkerConfiguration {
	return InitWorkerConfiguration{
		Cache: struct {
			Purge                      time.Duration
			Expiration                 time.Duration
			RemoteResponseCacheTimeout time.Duration
		}{Purge: time.Second * 60, Expiration: time.Second * 120 , RemoteResponseCacheTimeout: time.Second * 30},
		Limit: struct {
			OneRequest             int
			OneDevice              int
			CountWorkers           int
			RequestResetTimeout    int
			ResponseCollectorCount int
		}{OneRequest: 10, OneDevice: 5, CountWorkers: 100, RequestResetTimeout: 300, ResponseCollectorCount: 5},
	}
}


func New(config InitWorkerConfiguration) *Worker {
	log.Println("Initialize requestPooler")
	worker := new(Worker)

	worker.Config = config
	worker.Cache = cache.New(config.Cache.Expiration, config.Cache.Purge)
	worker.RequestCollector = cache.New(time.Second * 900, config.Cache.Purge)
	worker.LimitCountForSwitch = cache.New(time.Second * 900, config.Cache.Purge)
	worker.LimitCountForRequest = cache.New(time.Second * 900, config.Cache.Purge)
	worker.RequestQueue = make(chan Pooller, config.Limit.CountWorkers)
	worker.ResponseQueue = make(chan Pooller, config.Limit.ResponseCollectorCount)
	for i := 0; i <= config.Limit.CountWorkers; i++ {
		log.Println("Start snmp worker")
		go backgroundWorker(worker)
	}
	for i := 0; i <= config.Limit.ResponseCollectorCount; i++ {
		log.Println("Start response collector")
		go workerResponseCollector(worker)
	}
	return  worker
}

func (w *Worker) Get(requests []Request) []Response {
	return w.addToPool(requests, requestGet)
}

func (w *Worker) addToPool( requests []Request, requestType RequestType) []Response {
	//Generate requestId
	requestUUid := ""
	if  uuid, err :=  uuid.NewV4(); err == nil {
		requestUUid = uuid.String()
	}

	//Rebuild request for work with as map
	requestMapped := make(map[string]Pooller)
	for _, r := range  requests  {
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
			if count := w.addCountFromRequest(poolItem.UUid); count >= w.Config.Limit.OneRequest {
				continue
			}
			if count := w.addRequestForSwitch(poolItem.RequestBody.Ip); count >= w.Config.Limit.OneDevice {
				continue
			}
			w.RequestQueue <- poolItem
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
		if (time.Now().Unix() - startWaitingResponse) >  int64(w.Config.Limit.RequestResetTimeout) {
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
		CountRequestQueue: len(w.RequestQueue),
		CountResponseQueue: len(w.ResponseQueue),
	}
	status.CountWorkersForSw = make(map[string]int)
	status.CountWorkersForRequest = make(map[string]int)

	for name, count := range  w.LimitCountForSwitch.Items()  {
		status.CountWorkersForSw[name] = count.Object.(int)
	}
	for name, count := range  w.LimitCountForRequest.Items()  {
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
