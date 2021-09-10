package validator

import (
	"fmt"
	valid "gopkg.in/go-playground/validator.v9"
	"log"
	"regexp"
)

type Validator struct {
	Name  string
	Regex string
}

var validators = make([]Validator, 0)

func init() {
	validators = append(validators,
		Validator{
			Name:  "ip_address",
			Regex: `^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`,
		},
		Validator{
			Name:  "exclude_specials",
			Regex: `^[a-zA-Z0-9._@]{1,40}$`,
		},
		Validator{
			Name:  "oid",
			Regex: `^\.[[0-9|\.]{1,100}$`,
		},
		Validator{
			Name:  "zero_or_number",
			Regex: `^0$|^$|^[0-9]{1,4}$`,
		},
		Validator{
			Name:  "zero",
			Regex: `^0$|^$`,
		},
		Validator{
			Name:  "zero_or_email",
			Regex: `^$|^\S+@\S+\.\S+$`,
		},
	)
}

func GetValidator(tagName string) *valid.Validate {
	val := valid.New()
	val.SetTagName(tagName)
	for _, cust := range validators {
		custom := Validator{
			Regex: cust.Regex,
			Name:  cust.Name,
		}
		val.RegisterValidation(custom.Name, func(fl valid.FieldLevel) bool {
			if matched, err := regexp.MatchString(custom.Regex, fmt.Sprintf("%v", fl.Field())); err != nil {
				log.Panic("REGEX NOT CORRECT", custom.Regex, custom.Name, fmt.Sprintf("%v", fl.Field()))
			} else if matched {
				return true
			}
			return false
		})
	}

	return val
}
