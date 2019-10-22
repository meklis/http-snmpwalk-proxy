package pooller

import (
	"fmt"
	"github.com/meklis/go-cache"
	"github.com/meklis/http-snmpwalk-proxy/logger"
	"github.com/satori/go.uuid"
	"os"
	"time"
)

func GetDefaultConfiguration() InitWorkerConfiguration {
	return InitWorkerConfiguration{
		CacheExpiration:                 time.Second * 60,
		CachePurge:                      time.Second * 3,
		CacheRemoteResponseCacheTimeout: time.Second * 120,
		DefaultSnmpRepeats:              5,
		DefaultSnmpTimeout:              3,
		LimitCountWorkers:               100,
		LimitOneDevice:                  3,
		LimitOneRequest:                 50,
		LimitRequestResetTimeout:        300,
		LimitResponseCollectorCount:     10,
	}
}

func New(config InitWorkerConfiguration) *Worker {
	worker := new(Worker)

	worker.Config = config
	worker.cache = cache.New(config.CacheRemoteResponseCacheTimeout, config.CachePurge)
	worker.requestCollector = cache.New(time.Duration(config.LimitRequestResetTimeout+60)*time.Second, config.CachePurge)
	worker.limitCountForSwitch = cache.New(time.Duration(config.LimitRequestResetTimeout+60)*time.Second, config.CachePurge)
	worker.limitCountForRequest = cache.New(time.Second*900, config.CachePurge)
	worker.requestQueue = make(chan Pooller, config.LimitCountWorkers)
	worker.responseQueue = make(chan Pooller, config.LimitResponseCollectorCount)

	for i := uint(0); i < config.LimitCountWorkers; i++ {
		go backgroundWorker(worker, i)
	}
	for i := uint(0); i < config.LimitCountWorkers; i++ {
		go workerResponseCollector(worker, i)
	}
	if worker.Logger == nil {
		worker.Logger, _ = logger.New("pooler", 0, os.DevNull)
	}
	return worker
}

func (w *Worker) Get(requests []Request) []Response {
	return w.addToPool(requests, requestGet)
}

func (w *Worker) addToPool(requests []Request, requestType RequestType) []Response {
	//Generate requestId
	requestUUid := ""
	if uu, err := uuid.NewV4(); err == nil {
		requestUUid = uu.String()
	}
	w.Logger.DebugF("Add new request to pool, length request: %v, with generated uuid: %v", len(requests), requestUUid)

	//Rebuild request for work with as map
	RequestsMappedForDebug := make(map[string]Pooller)
	requestMapped := make(map[string]Pooller)
	for _, r := range requests {
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
			Type:         requestType,
		}
		RequestsMappedForDebug[fmt.Sprintf("%v:%v", r.Ip, r.Oid)] = Pooller{
			UUid:         requestUUid,
			RequestBody:  r,
			ResponseBody: Response{},
			Type:         requestType,
		}
	}
	countOfStartRequests := len(requestMapped)
	//Sending requests to pool
	for {
		if len(requestMapped) == 0 {
			w.Logger.DebugF("Requests are no longer in the buffer, waiting for an answer for uuid %v", requestUUid)
			break
		}
		for keyName, poolItem := range requestMapped {
			if count := w.getCountFromRequest(poolItem.UUid); count >= w.Config.LimitOneRequest {
				//w.Logger.WarningF("Reached limit of async workers in request for uuid %v (count requests: %v, limit: %v)", requestUUid, count, w.Config.LimitOneRequest)
				continue
			}
			if count := w.getCountRequestForSwitch(poolItem.RequestBody.Ip); count >= w.Config.LimitOneDevice {
				//w.Logger.WarningF("Reached limit of async workers for switch %v", poolItem.RequestBody.Ip)
				continue
			}
			w.addCountFromRequest(poolItem.UUid)
			w.addRequestForSwitch(poolItem.RequestBody.Ip)
			w.Logger.DebugF("Send request ip:%v, oid:%v, retries:%v, timeout:%v, use_cache:%v to pool with requestId %v", poolItem.RequestBody.Ip,
				poolItem.RequestBody.Oid,
				poolItem.RequestBody.Repeats,
				poolItem.RequestBody.Timeout,
				poolItem.RequestBody.UseCache,
				requestUUid,
			)
			w.requestQueue <- poolItem
			delete(requestMapped, keyName)
		}
	}
	requestMapped = nil
	startWaitingResponse := time.Now().Unix()
	response := make([]Response, 0)
	for {
		poollers := w.getRequestData(requestUUid)
		if len(poollers) == countOfStartRequests {
			for _, pool := range poollers {
				w.Logger.DebugF("Received response from %v-%v with requestId %v", pool.RequestBody.Ip, pool.RequestBody.Oid, requestUUid)
				response = append(response, pool.ResponseBody)
			}
			break
		}
		if (time.Now().Unix() - startWaitingResponse) > int64(w.Config.LimitRequestResetTimeout) {
			lenNoResp := countOfStartRequests - len(poollers)
			w.Logger.ErrorF("Reached timeout for request waiter, no response from %v requests, with requestId: %v", lenNoResp, requestUUid)
			for _, pool := range poollers {
				response = append(response, pool.ResponseBody)
				delete(RequestsMappedForDebug, fmt.Sprintf("%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid))
			}
			for _, d := range RequestsMappedForDebug {
				w.Logger.ErrorF("Did not wait for an answer from %v with oid %v, requestId %v", d.RequestBody.Ip, d.RequestBody.Oid, requestUUid)
			}
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	RequestsMappedForDebug = nil
	w.delRequestData(requestUUid)
	return response
}

func (w *Worker) GetStatus() StatusPooler {
	status := StatusPooler{
		CountRequestQueue:  len(w.requestQueue),
		CountResponseQueue: len(w.responseQueue),
	}
	status.CountWorkersForSw = make(map[string]uint)
	status.CountWorkersForRequest = make(map[string]uint)

	for name, count := range w.limitCountForSwitch.Items() {
		status.CountWorkersForSw[name] = count.Object.(uint)
	}
	for name, count := range w.limitCountForRequest.Items() {
		status.CountWorkersForRequest[name] = count.Object.(uint)
	}
	return status
}

func (w *Worker) Walk(r []Request) []Response {
	return w.addToPool(r, requestWalk)
}
func (w *Worker) BulkWalk(r []Request) []Response {
	return w.addToPool(r, requestBulkWalk)
}
func (w *Worker) Set(r []Request) []Response {
	return w.addToPool(r, requestSet)
}
