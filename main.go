package main


import (
	"flag"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/gin-gonic/gin"
	"log"
	"bitbucket.org/meklis/helpprovider_snmp/snmp-pooller"
	"time"
	"encoding/json"
	"net/http"
	"bitbucket.org/meklis/helpprovider_snmp/validator"
	"strconv"
	"bitbucket.org/meklis/helpprovider_snmp/logger"
	"os"
	"fmt"
)

const VERSION  = "1.0.8"

type Configuration struct{
	Handler struct{
		Listen string 	`yaml:"listen"`
		Prefix string 	`yaml:"prefix"`
	} `yaml:"handler"`
	Cache struct{
		Default struct{
			Purge int `yaml:"purge"`
		} `yaml:"defaults"`
		RemoteExpirationSec int `yaml:"remote_expiration_sec"`
	} `yaml:"cache"`
	System struct{
		MaxAsyncRequestToHost uint `yaml:"max_async_workers_to_host"`
		MaxAsyncWorkersForRequest uint `yaml:"max_async_workers_for_request"`
		CountWorkers uint `yaml:"count_workers"`
		RequestResetTimeoutSec uint `yaml:"request_reset_timeout_sec"`
		ResponseCollectorCount uint `yaml:"response_collector_count"`
		MaxCountInOneRequest uint `yaml:"max_count_in_one_request"`
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
	lg *logger.Logger
)

func init()  {
	flag.StringVar(&pathConfig, "c", "snmp.conf.yml", "Configuration file for proxy-auth module")
	flag.Parse()
}

func main() {
	PrintLogo()
	// Load configuration
	if err := LoadConfig(); err != nil {
		log.Panicln("ERROR LOADING CONFIGURATION FILE:", err.Error())
	}

	if Config.Logger.Console.Enabled {
		color := 0
		if Config.Logger.Console.EnabledColor {
			color = 1
		}
		lg, _ = logger.New("pooler", color, os.Stdout)
		lg.SetLogLevel(logger.LogLevel(Config.Logger.Console.LogLevel))
		if Config.Logger.Console.LogLevel < 5 {
			lg.SetFormat("#%{id} %{time} > %{level} %{message}")
		} else {
			lg.SetFormat("#%{id} %{time} (%{filename}:%{line}) > %{level} %{message}")
		}
	} else {
		lg, _ = logger.New("no_log",0, os.DevNull)
	}
	gin.SetMode("release")

	//Define gin
	gin.DefaultErrorWriter  = ioutil.Discard
	gin.DefaultWriter  = ioutil.Discard
	r := gin.Default()
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		lg.Notice(fmt.Sprintf("HTTP | %3d | %13v | %15s | %-7s  %s   %s",
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			param.ErrorMessage,
		))
		return ""
	}))

	//Initial snmp pooller
	pool := pooller.New(pooller.InitWorkerConfiguration{
		DefaultSnmpTimeout: Config.Snmp.Timeout,
		DefaultSnmpRepeats: Config.Snmp.Repeats,
		LimitRequestResetTimeout: Config.System.RequestResetTimeoutSec,
		LimitResponseCollectorCount: Config.System.ResponseCollectorCount,
		LimitCountWorkers: Config.System.CountWorkers,
		CacheRemoteResponseCacheTimeout: time.Duration(Config.Cache.RemoteExpirationSec) * time.Second,
		CachePurge: time.Second * time.Duration(Config.Cache.Default.Purge),
		CacheExpiration: time.Second * time.Duration(Config.System.RequestResetTimeoutSec + 600),
		LimitOneRequest: Config.System.MaxAsyncWorkersForRequest,
		LimitOneDevice: Config.System.MaxAsyncRequestToHost,
	})

	//logPooler.SetLogLevel(logger.WarningLevel)
	pool.Logger = lg

	r.Use(func(c *gin.Context) {
		c.Set("POOLLER", pool)
		c.Next()
	})

	go func() {
		for {
			data := pool.GetStatus()
			bytes, _ := json.Marshal(data)
			lg.DebugF("STATUS: %v",string(bytes))
			time.Sleep(time.Second * 3)
		}
	}()

	r.GET(Config.Handler.Prefix + "walk", func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_get").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "", "<h1>HelpProvider</h1><br><h3>SnmpPooller</h3>")
	})

	r.GET(Config.Handler.Prefix + "get_status", func(c *gin.Context) {
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.GetStatus())
	})

	r.GET(Config.Handler.Prefix + "bulk_walk",func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_get").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})
	r.GET(Config.Handler.Prefix + "get",func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_get").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})
	r.GET(Config.Handler.Prefix + "set", func(c *gin.Context) {
		request := formatGetRequest(c)
		if err := validator.GetValidator("snmp_set").Struct(&request[0]); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		P := c.MustGet("POOLLER").(*pooller.Worker)
		c.JSON(200, P.Walk(request))
	})

	r.POST(Config.Handler.Prefix + "get", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		if uint(len(data)) > Config.System.MaxCountInOneRequest {
			AbortWithStatus(c, http.StatusBadRequest, fmt.Sprintf("Reached max count devices in one request"))
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
	r.POST(Config.Handler.Prefix + "walk", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			log.Printf("Create unmarshall json %v", err.Error())
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		if uint(len(data)) > Config.System.MaxCountInOneRequest {
			AbortWithStatus(c, http.StatusBadRequest, fmt.Sprintf("Reached max count devices in one request"))
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

	r.POST(Config.Handler.Prefix + "bulk_walk", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			log.Printf("Create unmarshall json %v", err.Error())
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		if uint(len(data)) > Config.System.MaxCountInOneRequest {
			AbortWithStatus(c, http.StatusBadRequest, fmt.Sprintf("Reached max count devices in one request"))
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

	r.POST(Config.Handler.Prefix + "set", func(c *gin.Context) {
		var data []pooller.Request
		if err := c.BindJSON(&data); err != nil {
			log.Printf("Create unmarshall json %v", err.Error())
			AbortWithStatus(c, http.StatusBadRequest, err.Error())
			return
		}
		if uint(len(data)) > Config.System.MaxCountInOneRequest {
			AbortWithStatus(c, http.StatusBadRequest, fmt.Sprintf("Reached max count devices in one request"))
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
	lg.WarningF(msg)
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


func PrintLogo() {
	fmt.Printf(` 
 _     _         _          ______                        _      _               
| |   | |       | |        (_____ \                      (_)    | |              
| |___| | _____ | |  ____   _____) )  ____   ___   _   _  _   __| | _____   ____ 
|  ___  || ___ || | |  _ \ |  ____/  / ___) / _ \ | | | || | / _  || ___ | / ___)
| |   | || ____|| | | |_| || |      | |    | |_| | \ V / | |( (_| || ____|| |    
|_|   |_||_____) \_)|  __/ |_|      |_|     \___/   \_/  |_| \____||_____)|_|    
                    |_|
  ______                   ______            _ _              
 / _____)                 (_____ \          | | |             
( (____  ____  ____  ____  _____) )__   ___ | | | _____  ____ 
 \____ \|  _ \|    \|  _ \|  ____/ _ \ / _ \| | || ___ |/ ___)
 _____) ) | | | | | | |_| | |   | |_| | |_| | | || ____| |    
(______/|_| |_|_|_|_|  __/|_|    \___/ \___/ \_)_)_____)_|    
                    |_|                                                                       
ver %v
`, VERSION)
}