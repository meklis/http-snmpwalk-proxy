package handlers

import (
	"github.com/gin-gonic/gin"
	"helpprovider_snmp/snmp"
	"time"
	"net/http"
	"log"
	"../snmp-pooller"
)

func GetGet(c *gin.Context) {

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
	err, conn := snmp.Connect(snmp.InitStruct{
		Version: snmp.Version2c,
		TimeoutSec: c.MustGet("SnmpTimeout").(time.Duration),
		Repeats: c.MustGet("SnmpRepeats").(int),
		Community: Community,
		Ip: Ip,
	})
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer conn.Close()

	return
}

func GetWalk(c *gin.Context) {

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
	 c.JSON(200, P.Walk([]pooller.Request{
	 	pooller.Request{
	 		Ip: Ip,
	 		Oid: Oid,
	 		Community: Community,
	 		Timeout: 5,
	 		Repeats: 5,
	 		UseCache: true,
		},
	 }))
}
func GetBulkWalk(c *gin.Context) {

	Ip, Community, Oid , _, _:= formatRequest(c)

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
	err, conn := snmp.Connect(snmp.InitStruct{
		Version: snmp.Version2c,
		TimeoutSec: c.MustGet("SnmpTimeout").(time.Duration),
		Repeats: c.MustGet("SnmpRepeats").(int),
		Community: Community,
		Ip: Ip,
	})
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer conn.Close()

	return
}

func GetSet(c *gin.Context) {
	Ip, Community, _ , Type, Value:= formatRequest(c)
	if Ip == "" {
		AbortWithStatus(c, http.StatusBadRequest, "Ip can not be empty")
		return
	}
	if Community == "" {
		AbortWithStatus(c, http.StatusBadRequest, "Community can not be empty")
		return
	}
	if Type == "" {
		AbortWithStatus(c, http.StatusBadRequest, "Type can not be empty")
		return
	}
	if Value == "" {
		AbortWithStatus(c, http.StatusBadRequest, "Value can not be empty")
		return
	}
	err, conn := snmp.Connect(snmp.InitStruct{
		Version: snmp.Version2c,
		TimeoutSec: c.MustGet("SnmpTimeout").(time.Duration),
		Repeats: c.MustGet("SnmpRepeats").(int),
		Community: Community,
		Ip: Ip,
	})
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	defer conn.Close()

	return
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

func AbortWithStatus(c *gin.Context, code int, msg string) {
	log.Println(msg)
	c.String(code, msg)
}