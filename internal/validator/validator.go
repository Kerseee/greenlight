package validator

import "regexp"

// Validator contains a map of validation error.
type Validator struct {
	Errors map[string]string
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

const ErrMsgMustBeProvided = "must be provided"

// New return a Validator.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid return true if there is no error in Validator v.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error if the key does not exist. Otherwise does nothing.
func (v *Validator) AddError(key, msg string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = msg
	}
}

// Check adds an error if ok is false. Should be called during validation check.
func (v *Validator) Check(ok bool, key, msg string) {
	if !ok {
		v.AddError(key, msg)
	}
}

// In returns true if val is in a list of string.
func In(val string, list ...string) bool {
	for _, v := range list {
		if val == v {
			return true
		}
	}
	return false
}

// Matches returns true if val matches rx.
func Matches(val string, rx *regexp.Regexp) bool {
	return rx.MatchString(val)
}

// Unique returns true if all value in vals are unique.
func Unique(vals []string) bool {
	unique := make(map[string]bool)
	for _, val := range vals {
		if _, ok := unique[val]; ok {
			return false
		}
		unique[val] = true
	}
	return true
}
