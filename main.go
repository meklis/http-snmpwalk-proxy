package main


import (
	"github.com/patrickmn/go-cache"
	"flag"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/soniah/gosnmp"
	"time"
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

func main () {
	snmp:=gosnmp.Default

	snmp.Timeout = time.Duration(time.Second * 4)
	snmp.Retries = 3
	snmp.Community = "oblnswsnmpcomrw"
	snmp.Target = "10.50.124.132"
	if err := snmp.Connect(); err != nil {
		panic(err)
	}
	fmt.Println(snmp.Get([]string{".1.3.6.1.2.1.31.1.1.1.18.1"}))
}

/*
func main() {
	// Load configuration
	if err := LoadConfig(); err != nil {
		log.Panicln("ERROR LOADING CONFIGURATION FILE:", err.Error())
	}

	//Define gin
	r := gin.Default()

	//Define routes
	r.GET("/", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	//Define routes
	r.GET("/walk/:ip", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	//Define routes
	r.GET("/get/:ip", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	//Define routes
	r.GET("/set/:ip", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	r.POST("/get/:ip", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	r.POST("/walk/:ip", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})

	r.POST("/set/:ip", func(c *gin.Context) {
		r := r.Routes()
		c.JSON(200, ModuleStat{ModuleName: "SNMP walk module", Routes: r})
	})


	//Server http server
	r.Run(Config.Handler.Listen)
}*/
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
