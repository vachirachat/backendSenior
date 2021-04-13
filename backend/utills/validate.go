package utills

import (
	"log"

	"github.com/go-playground/validator"
)

type StructValidator struct {
	validate *validator.Validate
}

func NewValidator(validate *validator.Validate) *StructValidator {
	return &StructValidator{
		validate: validate,
	}
}

func (val StructValidator) ValidateStruct(variable interface{}) error {
	// returns nil or ValidationErrors ( []FieldError )
	log.Println("validate -> in struct")
	err := val.validate.Struct(variable)
	if err != nil {
		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if err, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}
		// from here you can create your own error messages in whatever language you wish
		return err
	}
	return nil
}

func (val StructValidator) ValidateVariable(variable interface{}, required string) error {
	errs := val.validate.Var(variable, required)
	if errs != nil {
		return errs
	}
	return nil

}
