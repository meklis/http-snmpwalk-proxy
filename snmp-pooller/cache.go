package pooller

import (
	"time"
	"fmt"
)
//TODO: Realize new cache over channels
//Current cache have concurency problem and has slow mutex locking


//Switch counter managment
func (w *Worker) getCountRequestForSwitch(swIp string) int {
	count, exist := w.limitCountForSwitch.Get(swIp)
	if !exist {
		w.Logger.DebugF("(AsyncSwitchCountWorkersCache) Data for count request of switch not found for ip %v, returned count = 0",swIp)
		return 0
	} else {
		//w.Logger.DebugF("(AsyncSwitchCountWorkersCache) Count request for switch %v = %v",swIp, count.(int))
		return  count.(int)
	}
}
func (w *Worker) addRequestForSwitch(swIp string) int {
	count, exist := w.limitCountForSwitch.Get(swIp)
	if !exist {
		w.Logger.DebugF("(AsyncSwitchCountWorkersCache) Create data for count request for one switch, ip=%v",swIp)
		w.limitCountForSwitch.Set(swIp, 1, time.Duration(w.Config.LimitRequestResetTimeout + 60) * time.Second)
		return 1
	} else {
		c := count.(int)
		c++
		if c <= 0 {
			c = 1
		}
		w.Logger.DebugF("(AsyncSwitchCountWorkersCache) Increment count request for switch %v from %v to %v",swIp, count.(int), c)
		if err := w.limitCountForSwitch.Replace(swIp, c, time.Duration(w.Config.LimitRequestResetTimeout+60)*time.Second); err != nil {
			w.Logger.WarningF("(AsyncSwitchCountWorkersCache) Error increment count limit for switch %v with err:%v",swIp, err.Error())
		}
		return  c
	}
}
func (w *Worker) deleteRequestForSwitch(swIp string) int {
	count, exist := w.limitCountForSwitch.Get(swIp)
	if exist {
		c := count.(int)
		c--
		if c <= 0 {
			w.Logger.DebugF("(AsyncSwitchCountWorkersCache) Delete decrement cache for ip=%v, because count = 0", swIp)
			w.limitCountForSwitch.Delete(swIp)
			return  0
		} else {
			if err := w.limitCountForSwitch.Replace(swIp, c, time.Duration(w.Config.LimitRequestResetTimeout+60)*time.Second); err != nil {
				w.Logger.WarningF("(AsyncSwitchCountWorkersCache) Error decrementing count limit for switch %v with err:%v",swIp, err.Error())
			}
			return  c
		}
	} else {
		w.Logger.WarningF("(AsyncSwitchCountWorkersCache) Err decrement count request for ip=%v, limit not found in cache",swIp)
	}
	return  0
}


//Request counter managment
func (w *Worker) addCountFromRequest(requestId string) int {
	count, exist := w.limitCountForRequest.Get(requestId)
	if !exist {
		w.Logger.DebugF("(AsyncRequestCountWorkersCache) %v request not found, creating new cache and return count=1", requestId)
		w.limitCountForRequest.Set(requestId, 1, time.Duration(w.Config.LimitRequestResetTimeout + 60) * time.Second)
		return 1
	} else {
		c:= count.(int)
		c++
		if c <= 0 {
			c = 1
		}
		if err := w.limitCountForRequest.Replace(requestId, c, time.Duration(w.Config.LimitRequestResetTimeout + 60) * time.Second); err != nil {
			w.Logger.WarningF("(AsyncRequestCountWorkersCache) Err increment count request for request %v, limit not found in cache",requestId)
		}
		return  c
	}
}
func (w *Worker) deleteCountFromRequest(requestId string) int {
	count, exist := w.limitCountForRequest.Get(requestId)
	if exist {
		c:=count.(int)
		c--
		if c <= 0 {
			w.limitCountForRequest.Delete(requestId)
			return  0
		} else {
			if err := w.limitCountForRequest.Replace(requestId, c, time.Duration(w.Config.LimitRequestResetTimeout+60)*time.Second); err != nil {
				w.Logger.WarningF("(AsyncRequestCountWorkersCache) Err decrement request limit for request: %v, err: %v",requestId,err)
			}
			return c
		}
	} else {
		w.Logger.WarningF("(AsyncRequestCountWorkersCache) Err decrement request limit for request %v - not exist",requestId)
	}
	return 0
}
func (w *Worker) getCountFromRequest(requestId string) int {
	count, exist := w.limitCountForRequest.Get(requestId)
	if exist {
	//	w.Logger.DebugF("(AsyncRequestCountWorkersCache) Cache for request %v with count = %v", requestId, count.(int))
		return  count.(int)
	} else {
		w.Logger.DebugF("(AsyncRequestCountWorkersCache) Cache for request %v not exist when requested count", requestId)
		return 0
	}

}


//Collected request-response data with key request UUID
//@TODO Has problem add response to collector when count of collector > 1
func (w *Worker) addRequestData(requestId string, poolItem Pooller) int {
	data, exist := w.requestCollector.Get(requestId)
	if !exist {
		w.Logger.DebugF("Create response collector for requestId %v",requestId)
		poollers := make([]Pooller, 0)
		poollers = append(poollers, poolItem)
		w.Logger.DebugF("Add response from device %v with requestId %v",poolItem.RequestBody.Ip, requestId)
		w.requestCollector.Set(requestId, poollers, time.Duration(w.Config.LimitRequestResetTimeout + 600) * time.Second)
		return 1
	} else {
		poollers := data.([]Pooller)
		poollers = append(poollers, poolItem)
		w.Logger.DebugF("Add response from device %v with requestId %v. Now collector len is: %v",poolItem.RequestBody.Ip, requestId, len(poollers))
		if err := w.requestCollector.Replace(requestId, poollers, time.Duration(w.Config.LimitRequestResetTimeout + 600) * time.Second); err != nil {
			w.Logger.WarningF("(addRequestData) Err add collector data for request %v : %v",requestId, err.Error())
		}
		return  len(poollers)
	}
}
func (w *Worker) delRequestData(requestId string) {
	_, exist := w.requestCollector.Get(requestId)
	if exist {
		w.requestCollector.Delete(requestId)
	}
}
func (w *Worker) getRequestData(requestId string) []Pooller {
	data, exist := w.requestCollector.Get(requestId)
	if !exist {
		poollers := make([]Pooller, 0)
		return  poollers
	} else {
		return  data.([]Pooller)
	}
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

