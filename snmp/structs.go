package snmp

type SnmpResp struct {
	Oid      string      `json:"oid"`
	HexValue string      `json:"hex_value"`
	Value    interface{} `json:"value"`
	Type     string      `json:"type"`
}
