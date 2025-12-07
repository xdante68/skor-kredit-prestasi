package helper

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func FormatValidationErrors(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errMsg string
		for i, e := range validationErrors {
			if i > 0 {
				errMsg += "; "
			}
			switch e.Tag() {
			case "required":
				errMsg += e.Field() + " wajib diisi"
			case "email":
				errMsg += e.Field() + " harus email"
			case "min":
				errMsg += e.Field() + " minimal " + e.Param() + " karakter"
			case "max":
				errMsg += e.Field() + " maksimal " + e.Param() + " karakter"
			case "oneof":
				errMsg += e.Field() + " harus salah satu: " + e.Param()
			case "gt":
				errMsg += e.Field() + " harus lebih besar dari " + e.Param()
			default:
				errMsg += e.Field() + " tidak valid"
			}
		}
		return errMsg
	}
	return err.Error()
}
