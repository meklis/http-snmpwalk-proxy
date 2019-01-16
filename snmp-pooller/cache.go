package pooller

import (
	"time"
	"fmt"
	"strings"
)
//TODO: Realize new cache over channels
//Current cache have concurency problem and has slow mutex locking


//Switch counter managment
func (w *Worker) getCountRequestForSwitch(swIp string) uint {
	count, exist := w.limitCountForSwitch.Get(swIp)
	if !exist {
		w.Logger.DebugF("(AsyncSwitchCountWorkersCache) Data for count request of switch not found for ip %v, returned count = 0",swIp)
		return 0
	} else {
		//w.Logger.DebugF("(AsyncSwitchCountWorkersCache) Count request for switch %v = %v",swIp, count.(int))
		return  count.(uint)
	}
}
func (w *Worker) addRequestForSwitch(swIp string) uint {
	return w.limitCountForSwitch.IncrementUintWithSet(swIp, 1)
}
func (w *Worker) deleteRequestForSwitch(swIp string) uint {
	return w.limitCountForSwitch.DecrementUintWithDelete(swIp, 1)
}

//Request counter managment
func (w *Worker) addCountFromRequest(requestId string) uint {
	return   w.limitCountForRequest.IncrementUintWithSet(requestId, 1)
}

func (w *Worker) deleteCountFromRequest(requestId string) uint {
	return   w.limitCountForRequest.DecrementUintWithDelete(requestId, 1)
}

func (w *Worker) getCountFromRequest(requestId string) uint {
	count, exist := w.limitCountForRequest.Get(requestId)
	if exist {
		return  count.(uint)
	} else {
		w.Logger.DebugF("(AsyncRequestCountWorkersCache) Cache for request %v not exist when requested count", requestId)
		return 0
	}

}


//Collected request-response data with key request UUID
//@TODO Has problem add response to collector when count of collector > 1
func (w *Worker) addRequestData(requestId string, poolItem Pooller)  {
	key := fmt.Sprintf("%v:%v:%v", requestId, poolItem.RequestBody.Ip, poolItem.RequestBody.Oid)
	w.requestCollector.Set(key,poolItem,time.Duration(w.Config.LimitRequestResetTimeout + 600) * time.Second)
}
func (w *Worker) delRequestData(requestId string) {
	items := w.requestCollector.Items()
	for k, _ := range items {
		if strings.Contains(k, requestId) {
			w.requestCollector.Delete(k)
		}
	}
}
func (w *Worker) getRequestData(requestId string) []Pooller {
	pool := make([]Pooller, 0)
	items := w.requestCollector.Items()
	for k, val := range items {
		if strings.Contains(k, requestId) {
			pool = append(pool, val.Object.(Pooller))
		}
	}
	return pool
}

//Cache for switch data response
func (w *Worker) setCacheResponseFromDevice(pool Pooller)  {
	w.Logger.InfoF("Add response to cache from %v, oid %v",pool.RequestBody.Ip, pool.RequestBody.Oid)
	key := fmt.Sprintf("%v:%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid, pool.Type)
	_, exist := w.cache.Get(key)
	if !exist {
		w.cache.Set(key, pool, w.Config.CacheRemoteResponseCacheTimeout)
	}
}
func (w *Worker) delCacheResponseFromDevice(pool Pooller)  {
	w.Logger.InfoF("Delete response from cache for %v, oid %v",pool.RequestBody.Ip, pool.RequestBody.Oid)
	key := fmt.Sprintf("%v:%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid, pool.Type)
	_, exist := w.cache.Get(key)
	if exist {
		w.cache.Delete(key)
	}
}
func (w *Worker) getCacheResponseFromDevice(pool Pooller) (bool, Pooller) {
	w.Logger.InfoF("Get response from cache for %v, oid %v",pool.RequestBody.Ip, pool.RequestBody.Oid)
	key := fmt.Sprintf("%v:%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid, pool.Type)
	data, exist := w.cache.Get(key)
	if exist {
		return  true, data.(Pooller)
	} else {
		w.Logger.InfoF("Cache not exist for %v, oid %v",pool.RequestBody.Ip, pool.RequestBody.Oid)
		return  false,pool
	}
}

