package snmp

import (
	"fmt"
	"github.com/gosnmp/gosnmp"
)

func (s *Snmp) Set(oid, tp string, value interface{}) (error, []SnmpResp) {
	pdus := make([]gosnmp.SnmpPDU, 1)
	tpInt := getTypeInverse(tp)
	if tpInt == 0 {
		return fmt.Errorf("Choosed type not exist"), nil
	}
	if val, ok := value.(int); ok {
		pdus[0] = gosnmp.SnmpPDU{
			Type:  gosnmp.Asn1BER(tpInt),
			Value: val,
			Name:  oid,
		}
	} else if val, ok := value.(string); ok {
		pdus[0] = gosnmp.SnmpPDU{
			Type:  gosnmp.Asn1BER(tpInt),
			Value: val,
			Name:  oid,
		}
	} else if val, ok := value.(float64); ok {
		pdus[0] = gosnmp.SnmpPDU{
			Type:  gosnmp.Asn1BER(tpInt),
			Value: int(val),
			Name:  oid,
		}
	} else {
		return fmt.Errorf("Choosed type not supported for set method"), nil
	}

	_, err := s.GoSnmp.Set(pdus)
	if err != nil {
		return err, nil
	}
	return s.Get(oid)
}
