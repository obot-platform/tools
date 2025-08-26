package env

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

func LoadEnvForStruct[T any](s *T) error {
	v := reflect.ValueOf(s).Elem() // Dereference pointer to get struct value
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct, got %v", t.Kind())
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		varName := field.Tag.Get("env")
		if varName == "" {
			continue
		}

		// Check environment and default if set
		value := os.Getenv(varName)
		if value == "" {
			defaultValue := field.Tag.Get("default")
			if defaultValue != "" {
				value = defaultValue
			} else {
				if field.Tag.Get("optional") == "true" {
					continue
				}
				return fmt.Errorf("missing required environment variable %s", varName)
			}
		}

		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			return fmt.Errorf("cannot set value for field %s", field.Name)
		}

		// We can add more types as needed
		switch field.Type.Kind() {
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				fieldValue.Set(reflect.ValueOf(strings.Split(value, ",")))
			} else {
				return fmt.Errorf("unsupported slice element type %v for field %s", field.Type.Elem().Kind(), field.Name)
			}
		case reflect.Pointer:
			if field.Type.Elem().Kind() == reflect.String {
				fieldValue.Set(reflect.ValueOf(&value))
			} else {
				return fmt.Errorf("unsupported pointer element type %v for field %s", field.Type.Elem().Kind(), field.Name)
			}
		default:
			return fmt.Errorf("unsupported type %v for field %s", field.Type.Kind(), field.Name)
		}
	}

	return nil
}

type ValidationError struct {
	Err error `json:"error"`
}

func (e ValidationError) Error() string {
	return e.Err.Error()
}

type FieldValidationError struct {
	EnvVar    string `json:"envVar"`
	Message   string `json:"message"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive"`
}

func (e FieldValidationError) Error() string {
	if e.Sensitive {
		return fmt.Sprintf("invalid environment variable %s: %s (value: %s)", e.EnvVar, e.Message, "REDACTED")
	}
	return fmt.Sprintf("invalid environment variable %s: %s (value: %s)", e.EnvVar, e.Message, e.Value)
}

func (e FieldValidationError) MarshalJSON() ([]byte, error) {
	// Create a copy of the struct for JSON marshaling
	value := e.Value
	if e.Sensitive {
		value = "REDACTED"
	}

	return json.Marshal(struct {
		EnvVar    string `json:"envVar"`
		Message   string `json:"message"`
		Value     string `json:"value"`
		Sensitive bool   `json:"sensitive"`
	}{
		EnvVar:    e.EnvVar,
		Message:   e.Message,
		Value:     value,
		Sensitive: e.Sensitive,
	})
}

type FieldValidationErrors []FieldValidationError

func (e FieldValidationErrors) Error() string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}
