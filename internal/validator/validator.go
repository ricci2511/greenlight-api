package validator

import "regexp"

var (
	// https://html.spec.whatwg.org/#valid-e-mail-address
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

// Initializes a new Validator instance.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Adds an error message to the map of errors if it doesn't already exist.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Adds an error entry to the map if the validation check passed to ok is false.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Checks if the passed value is contained in the permittedValues slice.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for _, permittedValue := range permittedValues {
		if value == permittedValue {
			return true
		}
	}

	return false
}

// Checks if the passed value matches the passed regex.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Checks if all values in the passed slice are unique.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}

	return len(values) == len(uniqueValues)
}
