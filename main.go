package main


import (
	"github.com/patrickmn/go-cache"
	"flag"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/gin-gonic/gin"
	"log"
	"./handlers"
	"helpprovider_snmp/snmp-pooller"
	"time"
	"encoding/json"
	"net/http"
)

type Configuration struct{
	Handler struct{
		Listen string 	`yaml:"listen"`
		Prefix string 	`yaml:"prefix"`
	} `yaml:"handler"`
	Cache struct{
		Default struct{
			Purge int `yaml:"purge"`
			Expiration int `yaml:"expiration"`
		} `yaml:"defaults"`
		RemoteExpirationSec int `yaml:"remote_expiration_sec"`
	} `yaml:"cache"`
	System struct{
		MaxAsyncRequestToHost int `yaml:"max_async_workers_to_host"`
		MaxAsyncWorkersForRequest int `yaml:"max_async_workers_for_request"`
		CountWorkers int `yaml:"count_workers"`
		RequestResetTimeoutSec int `yaml:"request_reset_timeout_sec"`
		ResponseCollectorCount int `yaml:"response_collector_count"`
	} `yaml:"system"`
	Snmp struct {
		Timeout int `yaml:"timeout"`
		Repeats int `yaml:"repeats"`
	} `yaml:"snmp"`
}

var (
	Config Configuration
	pathConfig string
	MODULES = make(map[string]string)
	CACHE *cache.Cache
)

func init()  {
	flag.StringVar(&pathConfig, "c", "snmp.conf.yml", "Configuration file for proxy-auth module")
	flag.Parse()
}

func main() {
	// Load configuration
	if err := LoadConfig(); err != nil {
		log.Panicln("ERROR LOADING CONFIGURATION FILE:", err.Error())
	}

	//Define gin
	r := gin.Default()

	//Initial snmp pooller
	pool := pooller.New(pooller.InitWorkerConfiguration{
		Cache: struct {
			Purge                 time.Duration
			Expiration            time.Duration
			RemoteResponseCacheTimeout time.Duration
		}{
			Purge: time.Duration(Config.Cache.Default.Purge) * time.Second,
			Expiration: time.Duration(Config.Cache.Default.Expiration) * time.Second,
			RemoteResponseCacheTimeout:  time.Duration(Config.Cache.RemoteExpirationSec) * time.Second,
		},
		Limit: struct {
			OneRequest             int
			OneDevice              int
			CountWorkers           int
			RequestResetTimeout    int
			ResponseCollectorCount int
		}{
			OneRequest: Config.System.MaxAsyncWorkersForRequest,
			OneDevice: Config.System.MaxAsyncRequestToHost,
			CountWorkers: Config.System.CountWorkers,
			RequestResetTimeout: Config.System.RequestResetTimeoutSec,
			ResponseCollectorCount: Config.System.ResponseCollectorCount,
		},
	})


	r.Use(func(c *gin.Context) {
		c.Set("POOLLER", pool)
		c.Next()
	})

	go func() {
		for {
			data := pool.GetStatus()
			bytes, _ := json.Marshal(data)
			log.Println(string(bytes))
			time.Sleep(time.Second * 1)
		}
	}()

	r.GET("/walk", func(c *gin.Context) {

		Ip, Community, Oid, _, _ := formatRequest(c)

		if Ip == "" {
			AbortWithStatus(c, http.StatusBadRequest, "Ip can not be empty")
			return
		}
		if Community == "" {
			AbortWithStatus(c, http.StatusBadRequest, "Community can not be empty")
			return
		}
		if Oid == "" {
			AbortWithStatus(c, http.StatusBadRequest, "Oid can not be empty")
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		request := make([]pooller.Request,0)
		request = append(request,pooller.Request{
			Ip: Ip,
			Oid: Oid,
			Community: Community,
			Timeout: 5,
			Repeats: 5,
			UseCache: true,
		})
		c.JSON(200, P.Walk(request))
	})
	r.GET("/bulk_walk",handlers.GetBulkWalk)
	r.GET("/get", handlers.GetGet)
	r.GET("/set", handlers.GetSet)
	r.POST("/get", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	r.POST("/walk", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})


	r.POST("/bulk_walk", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	r.POST("/set", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})


	//Server http server
	r.Run(Config.Handler.Listen)
}
func LoadConfig() error {
	bytes, err := ioutil.ReadFile(pathConfig)
	if err != nil {
		return  err
	}
	err  = yaml.Unmarshal(bytes, &Config)
	if err != nil {
		return  err
	}
	return  nil
}
type ModuleStat struct {
	ModuleName string `json:"module_name"`
	Routes gin.RoutesInfo`json:"routes"`
}

func AbortWithStatus(c *gin.Context, code int, msg string) {
	log.Println(msg)
	c.String(code, msg)
}



func formatRequest(c *gin.Context) (ip, community, oid, tp, value string ) {
	params := c.Request.URL.Query()
	var Ip, Community, Oid, Type, Value string
	if val, isset := params["ip"]; isset {
		Ip = val[0]
	}
	if val, isset := params["community"]; isset {
		Community = val[0]
	}
	if val, isset := params["oid"]; isset {
		Oid = val[0]
	}
	if val, isset := params["type"]; isset {
		Type = val[0]
	}
	if val, isset := params["value"]; isset {
		Value = val[0]
	}
	return  Ip, Community, Oid, Type, Value
}
