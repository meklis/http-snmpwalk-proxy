package workers

import (
	"../snmp"
	"time"
	"github.com/patrickmn/go-cache"
	"github.com/satori/go.uuid"
	"fmt"
)

type Pooller struct {
	UUid string
	RequestKey string
	RequestBody snmp.Request
	ResponseBody []snmp.Response
	Type RequestType
}
type StatusPooler struct {
	CountRequestQueue int `json:"request_queue"`
	CountResponseQueue int `json:"response_queue"`
	CountWorkersForSw map[string]int `json:"workers_in_sw"`
	CountWorkersForRequest map[string]int `json:"workers_in_request"`
}

type RequestType byte

const (
	requestGet RequestType = iota
	requestWalk
	requestBulkWalk
	requestSet
)


type InitWorkerConfiguration struct {
	Limit struct{
		OneRequest int
		OneDevice int
		CountWorkers int
	}
	Cache struct{
		Purge time.Duration
		Expiration time.Duration
		StatusIdTimeout time.Duration
		RemoteResponseTimeout time.Duration
	}
}

type Worker struct {
	Config InitWorkerConfiguration
	Cache *cache.Cache
	RequestCollector *cache.Cache
	LimitCountForSwitch *cache.Cache
	LimitCountForRequest *cache.Cache
    RequestQueue chan Pooller
    ResponseQueue chan Pooller
}


func New(config InitWorkerConfiguration) *Worker {
	worker := new(Worker)
	worker.Cache = cache.New(config.Cache.Expiration, config.Cache.Purge)
	worker.RequestCollector = cache.New(config.Cache.Expiration+ time.Second * 900, config.Cache.Purge + time.Second * 900)
	worker.LimitCountForSwitch = cache.New(config.Cache.Expiration+ time.Second * 900, config.Cache.Purge + time.Second * 900)
	worker.LimitCountForRequest = cache.New(config.Cache.Expiration + time.Second * 900, config.Cache.Purge + time.Second * 900)
	worker.RequestQueue = make(chan Pooller, config.Limit.CountWorkers)
	worker.ResponseQueue = make(chan Pooller, config.Limit.CountWorkers)
	for i := 0; i <= config.Limit.CountWorkers; i++ {
		go backgroundWorker(worker)
	}
	return  worker
}

func (w *Worker) Get(requests []snmp.Request) []snmp.Response {

	return  nil
}

func (w *Worker) addToPool( requests []snmp.Request, requestType RequestType) []snmp.Response {
	//Generate requestId
	requestUUid := ""
	if  uuid, err :=  uuid.NewV4(); err != nil {
		requestUUid = uuid.String()
	}

	//Rebuild request for work with as map
	requestMapped := make(map[string]Pooller)
	for _, r := range  requests  {
		requestMapped[fmt.Sprintf("%v:%v", r.Ip, r.Oid)] = Pooller{
			UUid:         requestUUid,
			RequestKey:   fmt.Sprintf("%v:%v", r.Ip, r.Oid),
			RequestBody:  r,
			ResponseBody: nil,
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
	//pollers := w.getRequestData(requestUUid)
	return  nil
}


func (w *Worker) workerResponseCollector() {
	//for {
//
//	}
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

func (w *Worker) getCountRequestForSwitch(swIp string) int {
	count, exist := w.LimitCountForSwitch.Get(swIp)
	if !exist {
		w.LimitCountForSwitch.Set(swIp, 0, time.Duration(time.Second * 900))
		return 0
	} else {
		return  count.(int)
	}
}
func (w *Worker) addRequestForSwitch(swIp string) int {
	count, exist := w.LimitCountForSwitch.Get(swIp)
	if !exist {
		w.LimitCountForSwitch.Set(swIp, 1, time.Duration(time.Second * 900))
		return 1
	} else {
		count.(int)++
		w.LimitCountForSwitch.Replace(swIp, count, time.Duration(time.Second * 900))
		return  count.(int)
	}
}
func (w *Worker) deleteRequestForSwitch(swIp string) int {
	count, exist := w.LimitCountForSwitch.Get(swIp)
	count.(int)--
	if !exist {
		w.LimitCountForSwitch.Set(swIp, 0, time.Duration(time.Second * 900))
		return 0
	} else {
		w.LimitCountForSwitch.Replace(swIp, count, time.Duration(time.Second * 900))
		return  count.(int)
	}
}
func (w *Worker) addCountFromRequest(requestId string) int {
	count, exist := w.LimitCountForSwitch.Get(requestId)
	if !exist {
		w.LimitCountForSwitch.Set(requestId, 1, time.Duration(time.Second * 900))
		return 1
	} else {
		count.(int)++
		w.LimitCountForSwitch.Replace(requestId, count, time.Duration(time.Second * 900))
		return  count.(int)
	}
}
func (w *Worker) deleteCountFromRequest(requestId string) int {
	count, exist := w.LimitCountForSwitch.Get(requestId)
	count.(int)--
	if !exist {
		w.LimitCountForSwitch.Set(requestId, 0, time.Duration(time.Second * 900))
		return 0
	} else {
		w.LimitCountForSwitch.Replace(requestId, count, time.Duration(time.Second * 900))
		return  count.(int)
	}
}
func (w *Worker) getCountFromRequest(requestId string) int {
	count, exist := w.LimitCountForSwitch.Get(requestId)
	if !exist {
		w.LimitCountForSwitch.Set(requestId, 0, time.Duration(time.Second * 900))
		return 0
	} else {
		return  count.(int)
	}
}
func (w *Worker) addRequestData(requestId string, poolItem Pooller) int {
	data, exist := w.RequestCollector.Get(requestId)
	if !exist {
		poollers := make([]Pooller, 0)
		poollers = append(poollers, poolItem)
		w.RequestCollector.Set(requestId, poollers, time.Duration(time.Second * 900))
		return 1
	} else {
		poollers := data.([]Pooller)
		poollers = append(poollers, poolItem)
		w.LimitCountForSwitch.Replace(requestId, poollers, time.Duration(time.Second * 900))
		return  len(poollers)
	}
}
func (w *Worker) delRequestData(requestId string) {
	_, exist := w.RequestCollector.Get(requestId)
	if exist {
		w.RequestCollector.Delete(requestId)
	}
}
func (w *Worker) getRequestData(requestId string) []Pooller {
	data, exist := w.RequestCollector.Get(requestId)
	if !exist {
		poollers := make([]Pooller, 0)
		return  poollers
	} else {
		return  data.([]Pooller)
	}
}
func (w *Worker) setCacheResponseFromDevice(pool Pooller)  {
	key := fmt.Sprintf("%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid)
	_, exist := w.Cache.Get(key)
	if !exist {
		w.Cache.Set(key, pool, w.Config.Cache.RemoteResponseTimeout)
	} else {
		w.Cache.Replace(key, pool, w.Config.Cache.RemoteResponseTimeout)
	}
}
func (w *Worker) delCacheResponseFromDevice(pool Pooller)  {
	key := fmt.Sprintf("%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid)
	_, exist := w.Cache.Get(key)
	if exist {
		w.Cache.Delete(key)
	}
}
func (w *Worker) getCacheResponseFromDevice(pool Pooller) (bool, Pooller) {
	key := fmt.Sprintf("%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid)
	data, exist := w.Cache.Get(key)
	if exist {
		return  true, data.(Pooller)
	} else {
		return  false,pool
	}
}

func (w *Worker) Walk(r []snmp.Request) []snmp.Response {

	return  nil
}
func (w *Worker) BulkWalk(r []snmp.Request) []snmp.Response {

	return  nil
}
func (w *Worker) Set(r []snmp.Request)  []snmp.Response {

	return  nil
}

func backgroundWorker(w *Worker) {
	for {
		time.Sleep(time.Millisecond * 3)

	}
}

