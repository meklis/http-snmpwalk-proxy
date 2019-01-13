package snmp

import (
	"fmt"
)


func (s *Snmp) Get(oid string) (error, []SnmpResp) {
	res, err := s.GoSnmp.Get([]string{oid})
	if err != nil {
		return  err, nil
	}
	resp := make([]SnmpResp,0)
	for _, r := range  res.Variables {
		str := SnmpResp{
			Value: convertValue(r.Type, r.Value),
			HexValue: stringToBytes(fmt.Sprintf("%v", convertValue(r.Type, r.Value))),
			Oid: r.Name,
		}
		str.Type = getType(byte(r.Type))
		resp = append(resp, str)
	}
	return  nil, resp
}
func (s *Snmp) Walk(oid string) (error, []SnmpResp) {
	res, err := s.GoSnmp.WalkAll( oid )

	if err != nil {
		return  err, nil
	}
	resp := make([]SnmpResp,0)
	if err  == nil && len(res) == 0 {
		return  s.Get(oid)
	}

	for _, r := range  res {
		str := SnmpResp{
			Value: convertValue(r.Type, r.Value),
			HexValue: stringToBytes(fmt.Sprintf("%v", convertValue(r.Type, r.Value))),
			Oid: r.Name,
		}
		str.Type = getType(byte(r.Type))
		resp = append(resp, str)
	}
	return  nil, resp
}
func (s *Snmp) WalkBulk(oid string) (error, []SnmpResp) {
	res, err := s.GoSnmp.BulkWalkAll( oid )

	if err != nil {
		return  err, nil
	}
	resp := make([]SnmpResp,0)
	if err  == nil && len(res) == 0 {
		return  s.Get(oid)
	}

	for _, r := range  res {
		str := SnmpResp{
			Value: convertValue(r.Type, r.Value),
			HexValue: stringToBytes(fmt.Sprintf("%v", convertValue(r.Type, r.Value))),
			Oid: r.Name,
		}
		str.Type = getType(byte(r.Type))
		resp = append(resp, str)
	}
	return  nil, resp
}
