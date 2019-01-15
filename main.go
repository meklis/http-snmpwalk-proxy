package main


import (
	"flag"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/gin-gonic/gin"
	"log"
	"./snmp-pooller"
	"time"
	"encoding/json"
	"net/http"
	"./validator"
	"strconv"
	"./logger"
	"os"
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
	Logger struct {
		Console struct  {
			Enabled bool  `yaml:"enabled"`
			EnabledColor bool `yaml:"enable_color"`
			LogLevel int `yaml:"log_level"`
 		} `yaml:"console"`
	} `yaml:"logger"`
}

var (
	Config Configuration
	pathConfig string
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
	logPooler, _ := logger.New("pooler", 1, os.Stdout)

	gin.SetMode("release")


	//Define gin
	r := gin.Default()


	//Initial snmp pooller
	pool := pooller.New(pooller.InitWorkerConfiguration{
		DefaultSnmpTimeout: Config.Snmp.Timeout,
		DefaultSnmpRepeats: Config.Snmp.Repeats,
		LimitRequestResetTimeout: Config.System.RequestResetTimeoutSec,
		LimitResponseCollectorCount: Config.System.ResponseCollectorCount,
		LimitCountWorkers: Config.System.CountWorkers,
		CacheRemoteResponseCacheTimeout: time.Duration(Config.Cache.RemoteExpirationSec) * time.Second,
		CachePurge: time.Second * time.Duration(Config.Cache.Default.Purge),
		CacheExpiration: time.Second * time.Duration(Config.Cache.Default.Expiration),
		LimitOneRequest: Config.System.MaxAsyncWorkersForRequest,
		LimitOneDevice: Config.System.MaxAsyncRequestToHost,

	})

	//logPooler.SetLogLevel(logger.WarningLevel)
	pool.Logger = logPooler

	r.Use(func(c *gin.Context) {
		c.Set("POOLLER", pool)
		c.Next()
	})

	go func() {
		for {
			data := pool.GetStatus()
			bytes, _ := json.Marshal(data)
			logPooler.DebugF("STATUS: %v",string(bytes))
			time.Sleep(time.Second * 3)
		}
	}()

	r.GET("/walk", func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_get").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})

	r.GET("/get_status", func(c *gin.Context) {
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.GetStatus())
	})

	r.GET("/bulk_walk",func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_get").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})
	r.GET("/get",func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_get").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})
	r.GET("/set", func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_set").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})

	r.POST("/get", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			log.Printf("Create unmarshall json %v", err.Error())
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		for _, d := range  data {
			if err := validator.GetValidator("snmp_get").Struct(&d); err != nil {
				AbortWithStatus(c, http.StatusBadRequest, err.Error())
				return
			}
		}

		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Get(data))
	})
	r.POST("/walk", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			log.Printf("Create unmarshall json %v", err.Error())
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		for _, d := range  data {
			if err := validator.GetValidator("snmp_get").Struct(&d); err != nil {
				AbortWithStatus(c, http.StatusBadRequest, err.Error())
				return
			}
		}

		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(data))
	})

	r.POST("/bulk_walk", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			log.Printf("Create unmarshall json %v", err.Error())
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		for _, d := range  data {
			if err := validator.GetValidator("snmp_get").Struct(&d); err != nil {
				AbortWithStatus(c, http.StatusBadRequest, err.Error())
				return
			}
		}

		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.BulkWalk(data))
	})

	r.POST("/set", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			log.Printf("Create unmarshall json %v", err.Error())
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		for _, d := range  data {
			if err := validator.GetValidator("snmp_set").Struct(&d); err != nil {
				AbortWithStatus(c, http.StatusBadRequest, err.Error())
				return
			}
		}

		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Set(data))
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

func AbortWithStatus(c *gin.Context, code int, msg string) {
	log.Println(msg)
	c.String(code, msg)
}



func formatGetRequest(c *gin.Context) ([]pooller.Request) {
	params := c.Request.URL.Query()
	var Ip, Community, Oid, Type, Value  string
	var Repeats, Timeout int
	var UseCache bool
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
	if val, isset := params["repeats"]; isset {
		Repeats, _ = strconv.Atoi(val[0])
	}
	if val, isset := params["timeout"]; isset {
		Timeout, _ = strconv.Atoi(val[0])
	}
	if val, isset := params["use_cache"]; isset {
		if val[0] == "1" || val[0] == "true" {
			UseCache = true
		}
	}
	pool := make([]pooller.Request,1)
	pool[0] =  pooller.Request{
		Repeats: Repeats,
		UseCache: UseCache,
		Timeout: Timeout,
		Type: Type,
		Ip: Ip,
		Oid: Oid,
		Community: Community,
		Value: Value,
	}
	return pool
}
