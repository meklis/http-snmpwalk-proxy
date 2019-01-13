package snmp

import (
	"fmt"
	"github.com/soniah/gosnmp"
)

func (s *Snmp) Set(oid, tp string, value interface{}) (error, []SnmpResp) {
	pdus := make([]gosnmp.SnmpPDU,1)
	tpInt :=  getTypeInverse(tp)
	if tpInt == 0 {
		return  fmt.Errorf("Choosed type not exist" ), nil
	}


	pdus[0] = gosnmp.SnmpPDU{
		Type: gosnmp.Asn1BER(tpInt),
		Value: value,
		Name: oid,
	}
	_, err := s.GoSnmp.Set(pdus)
	if err != nil {
		return  err, nil
	}
	return s.Get(oid)
}