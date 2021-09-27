module github.com/meklis/http-snmpwalk-proxy

go 1.16

replace github.com/meklis/http-snmpwalk-proxy => ./

require (
	github.com/gin-gonic/gin v1.7.4
	github.com/gosnmp/gosnmp v1.32.0
	github.com/meklis/go-cache v2.1.0+incompatible
	github.com/satori/go.uuid v1.2.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0
	gopkg.in/yaml.v2 v2.4.0
)
