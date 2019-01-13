package snmp

import (
	"github.com/soniah/gosnmp"
	"time"
	"./structs"
)

type SnmpVersion uint8

// SnmpVersion 1, 2c and 3 implemented
const (
//	Version1  SnmpVersion = 0x0
	Version2c SnmpVersion = 0x1
//	Version3  SnmpVersion = 0x3
)


type InitStruct struct {
	Version SnmpVersion
	TimeoutSec time.Duration
	Repeats int
	Ip string
	Community string
}

type Snmp struct {
	init *InitStruct
	GoSnmp *gosnmp.GoSNMP
	Err error
}

var (
	types map[uint8]string
)
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
}

func Connect(conf InitStruct) (error, *Snmp) {
	resp := new(Snmp)
	resp.init = &conf
	snmp:=gosnmp.Default
	snmp.Version = gosnmp.SnmpVersion(conf.Version)
	snmp.Timeout = conf.TimeoutSec
	snmp.Retries = conf.Repeats
	snmp.Community = conf.Community
	snmp.Target = conf.Ip
	if err := snmp.Connect(); err != nil {
		return  err, nil
	}
	return nil,resp
}

func (s *Snmp) Get(oid string) (error, []structs.SnmpResp) {
	res, err := s.GoSnmp.Get([]string{oid})
	if err != nil {
		return  err, nil
	}
	resp := make([]structs.SnmpResp,0)
	for _, r := range  res.Variables {
		r.Type
	}
}

func getType(tp uint8) string {
 	types := make([])
}