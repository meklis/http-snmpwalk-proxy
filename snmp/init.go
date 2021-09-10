package snmp

import (
	"fmt"
	"github.com/gosnmp/gosnmp"
	"strings"
	"time"
)

type SnmpVersion uint8

// SnmpVersion 1, 2c and 3 implemented
const (
	//	Version1  SnmpVersion = 0x0
	Version2c SnmpVersion = 0x1
	//	Version3  SnmpVersion = 0x3
)

var (
	types        = make(map[byte]string)
	typesInverse = make(map[string]byte)
)

type InitStruct struct {
	Version    SnmpVersion
	TimeoutSec time.Duration
	Repeats    int
	Ip         string
	Community  string
}

type Snmp struct {
	init   *InitStruct
	GoSnmp *gosnmp.GoSNMP
	Err    error
}

func init() {
	types[0x00] = "EndOfContents"
	types[0x00] = "UnknownType"
	types[0x01] = "Boolean"
	types[0x02] = "Integer"
	types[0x03] = "BitString"
	types[0x04] = "OctetString"
	types[0x05] = "Null"
	types[0x06] = "ObjectIdentifier"
	types[0x07] = "ObjectDescription"
	types[0x40] = "IPAddress"
	types[0x41] = "Counter32"
	types[0x42] = "Gauge32"
	types[0x43] = "TimeTicks"
	types[0x44] = "Opaque"
	types[0x45] = "NsapAddress"
	types[0x46] = "Counter64"
	types[0x47] = "Uinteger32"
	types[0x78] = "OpaqueFloat"
	types[0x79] = "OpaqueDouble"
	types[0x80] = "NoSuchObject"
	types[0x81] = "NoSuchInstance"
	types[0x82] = "EndOfMibView"

	for bt, str := range types {
		typesInverse[str] = bt
	}
}

func Connect(conf InitStruct) (error, *Snmp) {
	resp := new(Snmp)
	resp.init = &conf
	snmp := gosnmp.GoSNMP{
		Port:               161,
		Community:          "public",
		Version:            gosnmp.SnmpVersion(conf.Version),
		Timeout:            time.Duration(conf.TimeoutSec) * time.Second,
		Retries:            conf.Repeats,
		ExponentialTimeout: false,
		MaxOids:            120,
	}
	snmp.Version = gosnmp.SnmpVersion(conf.Version)
	snmp.Timeout = conf.TimeoutSec
	snmp.Retries = conf.Repeats
	snmp.Community = conf.Community
	snmp.MaxRepetitions = uint32(conf.Repeats)
	snmp.Target = conf.Ip
	snmp.ExponentialTimeout = false
	if err := snmp.Connect(); err != nil {
		return err, nil
	}
	resp.GoSnmp = &snmp
	return nil, resp
}
func (c *Snmp) Close() {
	c.GoSnmp.Conn.Close()
	c = nil
}

func convertValue(tp gosnmp.Asn1BER, val interface{}) (resp interface{}) {
	switch tp {
	case 0x03:
		return fmt.Sprintf("%v", string(val.([]byte)))
	case 0x04:
		return fmt.Sprintf("%v", string(val.([]byte)))
	}
	return val
}

func getType(tp byte) string {
	return types[tp]
}
func getTypeInverse(tp string) byte {
	return typesInverse[tp]
}

func stringToBytes(str string) string {
	bytes := []byte(str)
	arr := ""
	for _, b := range bytes {
		bb := fmt.Sprintf("%X:", b)
		if len(bb) == 2 {
			bb = "0" + bb
		}
		arr += bb
	}
	return strings.Trim(arr, ":")
}
