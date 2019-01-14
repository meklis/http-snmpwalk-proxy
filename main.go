package main


import (
	"github.com/patrickmn/go-cache"
	"flag"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"time"
	"./handlers"
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
		StoreExpirationSec int `yaml:"store_expiration_sec"`
	} `yaml:"cache"`
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

	SnmpTimeOut := time.Second * time.Duration(Config.Snmp.Timeout)

	//Define gin
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Set("SnmpRepeats", Config.Snmp.Repeats)
		c.Next()
	})
	r.Use(func(c *gin.Context) {
		c.Set("SnmpTimeout", SnmpTimeOut)
		c.Next()
	})


	r.GET("/walk",handlers.GetWalk)
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

func PrintStarted() {
	fmt.Printf(`
Started module SNMP
ver: 1.0
date: 2019-01-11

configuration:
   handler.listen=%v
   handler.prefix=%v
   cache.default.purge=%v
   cache.default.expiration=%v
   cache.remote_expiration_sec=%v
   cache.store_expiration_sec=%v
   snmp.timeout=%v
   snmp.repeats=%v
`, Config.Handler.Listen, Config.Handler.Prefix, Config.Cache.Default.Purge, Config.Cache.Default.Expiration, Config.Cache.RemoteExpirationSec, Config.Cache.StoreExpirationSec, Config.Snmp.Timeout, Config.Snmp.Repeats)
}
