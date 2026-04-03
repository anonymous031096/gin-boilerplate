package validator

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("vn_phone", vnPhone)
		v.RegisterValidation("strong_password", strongPassword)
	}
}

// vn_phone validates Vietnamese phone numbers
// Example: 0912345678, 0387654321
func vnPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	matched, _ := regexp.MatchString(`^(0[3|5|7|8|9])\d{8}$`, phone)
	return matched
}

// strong_password requires min 8 chars, 1 uppercase, 1 lowercase, 1 special character
func strongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*(),.?":{}|<>]`, password)
	return hasUpper && hasLower && hasSpecial
}
