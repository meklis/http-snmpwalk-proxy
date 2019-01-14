package snmp

type Request struct {
	Ip string `json:"ip"`
	Community string `json:"community"`
	Oid string `json:"oid"`
	Repeats int `json:"repeats"`
	Timeout int `json:"timeout"`
	Type string `json:"type"`
	Value string `json:"value"`
	UseCache bool `json:"use_cache"`
}

type Response struct {
	Ip string `json:"ip"`
	Error string  `json:"error"`
	Oid string  `json:"oid"`
	Response []SnmpResp `json:"response,omitempty"`
}

type SnmpResp struct {
	Oid string `json:"oid"`
	HexValue string `json:"hex_value"`
	Value interface{} `json:"value"`
	Type string `json:"type"`
}



