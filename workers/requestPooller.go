package workers

import (
	"../snmp"
	"time"
	"github.com/patrickmn/go-cache"
)

type PoollerRequest struct {
	UUid string
	RequestKey string
	RequestBody snmp.Request
}

type PoolerResponse struct {
	UUid string
	ResponseBody snmp.Response
	RequestKey string
}


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
	RequestCache *cache.Cache
    RequestQueueChan chan string
}


func Init(config InitWorkerConfiguration) *Worker {
	worker := new(Worker)
	worker.RequestCache = cache.New(config.Cache.Expiration, config.Cache.Purge)
	worker.RequestQueueChan = make(chan string, config.Limit.CountWorkers)
	for i := 0; i <= config.Limit.CountWorkers; i++ {
		go backgroundWorker(worker)
	}
	return  worker
}

func Get(requests []snmp.Request) []snmp.Response {
	//uu, _ :=  uuid.NewV4(); UUid := uu.String()

	//for _, r := range  requests {
		//pooler := PoollerRequest {
		//	UUid: UUid,
		//	RequestKey: fmt.Sprintf("%v:%v", r.Ip, r.Oid),
		//	RequestBody: r,
		//}

	//}
	return  nil
}
func Walk(r []snmp.Request) []snmp.Response {

	return  nil
}
func BulkWalk(r []snmp.Request) []snmp.Response {

	return  nil
}
func Set(r []snmp.Request)  []snmp.Response {

	return  nil
}

func backgroundWorker(w *Worker) {
	for {
		time.Sleep(time.Millisecond * 3)

	}
}

