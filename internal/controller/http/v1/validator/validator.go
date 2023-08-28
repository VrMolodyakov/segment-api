package validator

import (
	validate "github.com/go-playground/validator/v10"
)

var (
	validator = validate.New()
)

type ValidateError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Param string `json:"param"`
}

func Validate(r any) []ValidateError {
	var errors []ValidateError
	err := validator.Struct(r)
	if err != nil {
		for _, err := range err.(validate.ValidationErrors) {
			var re ValidateError
			re.Field = err.Field()
			re.Tag = err.Tag()
			re.Param = err.Param()
			errors = append(errors, re)
		}
	}
	return errors
}
