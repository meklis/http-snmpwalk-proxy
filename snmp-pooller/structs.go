package pooller

import (
	"time"
	"github.com/patrickmn/go-cache"
	"../snmp"
	"../logger"
)

type Pooller struct {
	UUid string
	RequestBody Request
	ResponseBody  Response
	Type RequestType
}
type StatusPooler struct {
	CountRequestQueue int `json:"request_queue"`
	CountResponseQueue int `json:"response_queue"`
	CountWorkersForSw map[string]int `json:"workers_in_sw"`
	CountWorkersForRequest map[string]int `json:"workers_in_request"`
}

type Request struct {
	Ip string `json:"ip" snmp_get:"required,ip_address" snmp_set:"required,ip_address"`
	Community string `json:"community" snmp_get:"required,exclude_specials" snmp_set:"required,exclude_specials"`
	Oid string `json:"oid" snmp_get:"required,oid" snmp_set:"required,oid"`
	Repeats int `json:"repeats"`
	Timeout int `json:"timeout"`
	Type string `json:"type" snmp_set:"required,exclude_specials"`
	Value string `json:"value" snmp_set:"required,exclude_specials"`
	UseCache bool `json:"use_cache"`
}

type Response struct {
	Ip string `json:"ip"`
	Error string  `json:"error"`
	Oid string  `json:"oid"`
	Response []snmp.SnmpResp `json:"response,omitempty"`
	FromCache bool `json:"from_cache"`
}

type RequestType byte

const (
	requestGet RequestType = iota
	requestWalk
	requestBulkWalk
	requestSet
)


type InitWorkerConfiguration struct {
	LimitOneRequest int
	LimitOneDevice int
	LimitCountWorkers int
	LimitRequestResetTimeout int
	LimitResponseCollectorCount int
	CachePurge time.Duration
	CacheExpiration time.Duration
	CacheRemoteResponseCacheTimeout time.Duration
	DefaultSnmpTimeout int
	DefaultSnmpRepeats int
}

type Worker struct {
	Config InitWorkerConfiguration
	Logger *logger.Logger
	cache *cache.Cache
	requestCollector *cache.Cache
	limitCountForSwitch *cache.Cache
	limitCountForRequest *cache.Cache
	requestQueue chan Pooller
	responseQueue chan Pooller

}
