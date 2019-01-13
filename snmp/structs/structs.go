package structs

type Request struct {
	Ip string `json:"ip"`
	Community string `json:"community"`
	Oid string `json:"oid"`
	Repeats int `json:"repeats"`
	Timeout int `json:"timeout"`
	Type string `json:"type"`
	Value string `json:"value"`
	Version string `json:"version"`

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
	StringValue string `json:"string_value"`
	Type string `json:"type"`
}

