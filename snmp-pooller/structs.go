package pooller

import (
	"time"
	"github.com/patrickmn/go-cache"
	"helpprovider_snmp/snmp"
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
	Ip string `json:"ip"`
	Community string `json:"community"`
	Oid string `json:"oid"`
	Repeats int `json:"repeats"`
	Timeout int `json:"timeout"`
	Type string `json:"type"`
	Value string `json:"value"`
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
	Limit struct{
		OneRequest int
		OneDevice int
		CountWorkers int
		RequestResetTimeout int
		ResponseCollectorCount int
	}
	Cache struct{
		Purge time.Duration
		Expiration time.Duration
		RemoteResponseCacheTimeout time.Duration
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
