package pooller

import (
	"time"
	"fmt"
	"log"
)

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
		c := count.(int)
		c++
		if c < 0 {
			c = 0
		}
		if err := w.LimitCountForSwitch.Replace(swIp, c, time.Duration(time.Second * 900)); err != nil {
			log.Println("Cache work ERR:",err.Error())
		}
		return  c
	}
}
func (w *Worker) deleteRequestForSwitch(swIp string) int {
	count, exist := w.LimitCountForSwitch.Get(swIp)

	if !exist {
		w.LimitCountForSwitch.Set(swIp, 0, time.Duration(time.Second * 900))
		return 0
	} else {
		c := count.(int)
		c--
		if c < 0 {
			c = 0
		}
		if err := w.LimitCountForSwitch.Replace(swIp, c, time.Duration(time.Second * 900)); err != nil {
			log.Println("Cache work ERR:",err.Error())
		}
		return  c
	}
}
func (w *Worker) addCountFromRequest(requestId string) int {
	count, exist := w.LimitCountForRequest.Get(requestId)
	if !exist {
		w.LimitCountForRequest.Set(requestId, 1, time.Duration(time.Second * 900))
		return 1
	} else {
		c:= count.(int)
		c++
		if c < 0 {
			c = 0
		}
		if err := w.LimitCountForRequest.Replace(requestId, c, time.Duration(time.Second * 900)); err != nil {
			log.Println("Cache work ERR:",err.Error())
		}
		return  c
	}
}
func (w *Worker) deleteCountFromRequest(requestId string) int {
	count, exist := w.LimitCountForRequest.Get(requestId)
	if !exist {
		w.LimitCountForRequest.Set(requestId, 0, time.Duration(time.Second * 900))
		return 0
	} else {
		c:=count.(int)
		c--
		if c < 0 {
			c = 0
		}
		if err := w.LimitCountForRequest.Replace(requestId, c, time.Duration(time.Second * 900)); err != nil {
			log.Println("Cache work ERR:",err.Error())
		}
		return  c
	}
}
func (w *Worker) getCountFromRequest(requestId string) int {
	count, exist := w.LimitCountForRequest.Get(requestId)
	if !exist {
		w.LimitCountForRequest.Set(requestId, 0, time.Duration(time.Second * 900))
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
		if err := w.LimitCountForSwitch.Replace(requestId, poollers, time.Duration(time.Second * 900)); err != nil {
			log.Println("Cache work ERR:",err.Error())
		}
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
	key := fmt.Sprintf("%v:%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid, pool.Type)
	_, exist := w.Cache.Get(key)
	if !exist {
		w.Cache.Set(key, pool, w.Config.Cache.RemoteResponseCacheTimeout)
	}
}
func (w *Worker) delCacheResponseFromDevice(pool Pooller)  {
	key := fmt.Sprintf("%v:%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid, pool.Type)
	_, exist := w.Cache.Get(key)
	if exist {
		w.Cache.Delete(key)
	}
}
func (w *Worker) getCacheResponseFromDevice(pool Pooller) (bool, Pooller) {
	key := fmt.Sprintf("%v:%v:%v", pool.RequestBody.Ip, pool.RequestBody.Oid, pool.Type)
	data, exist := w.Cache.Get(key)
	if exist {
		return  true, data.(Pooller)
	} else {
		return  false,pool
	}
}

