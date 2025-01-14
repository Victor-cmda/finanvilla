package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func Init() {
	validate = validator.New()

	// Registro de validações customizadas
	_ = validate.RegisterValidation("cpf", validateCPF)
	_ = validate.RegisterValidation("phone", validatePhone)
}

func Validate(i interface{}) error {
	return validate.Struct(i)
}

func validateCPF(fl validator.FieldLevel) bool {
	cpf := fl.Field().String()
	regex := regexp.MustCompile(`^\d{3}\.\d{3}\.\d{3}-\d{2}$`)
	return regex.MatchString(cpf)
}

func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	regex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return regex.MatchString(phone)
}
