// +build test

package finder

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"

	"github.com/rancher/octopus/test/util/fuzz"
)

// Parse fills the env value to the given input pointer.
func Parse(ptr interface{}) error {
	if ptr == nil {
		return errors.New("cannot parse nil input")
	}

	// check the input is pointer or not
	var ptrValue = reflect.ValueOf(ptr)
	if ptrValue.Kind() != reflect.Ptr {
		return errors.New("cannot parse none pointer input")
	}
	if ptrValue.IsNil() {
		return errors.New("cannot parse nil pointer")
	}

	// parse the "env" tag
	var ptrTypeElem = reflect.TypeOf(ptr).Elem()
	var ptrValueElem = ptrValue.Elem()
	for i := 0; i < ptrValueElem.NumField(); i++ {
		if err := parseEnvFieldValue(ptrTypeElem.Field(i), ptrValueElem.Field(i)); err != nil {
			return errors.Wrap(err, "failed to parse")
		}
	}

	return nil
}

func parseEnvFieldValue(fieldDescriptor reflect.StructField, fieldValue reflect.Value) error {
	var envName, envDefaultValue string

	if tag, existed := fieldDescriptor.Tag.Lookup("env"); existed {
		var parts = strings.Split(tag, ",")
		for _, part := range parts {
			// env name
			if strings.HasPrefix(part, "name=") {
				envName = part[5:]
			}
			// default value
			if strings.HasPrefix(part, "default=") {
				envDefaultValue = part[8:]
			}
			// fuzz mode
			if strings.HasPrefix(part, "fuzz") {
				switch part {
				case "fuzzPort":
					var port, err = fuzz.FreePort()
					if err != nil {
						return errors.Wrapf(err, "failed to fuzz a new port for field %s", fieldDescriptor.Name)
					}
					envDefaultValue = fmt.Sprint(port)
				case "fuzzString":
					envDefaultValue = fuzz.String(10)
				case "fuzzFile":
					var path, _, err = fuzz.File(1 * fuzz.MB)
					if err != nil {
						return errors.Wrapf(err, "failed to fuzz a new file for field %s", fieldDescriptor.Name)
					}
					envDefaultValue = path
				default:
					return errors.Errorf("cannot identify fuzz mode %q", part)
				}
			}
		}
		if envName == "" {
			envName = strcase.ToScreamingSnake(fieldDescriptor.Name)
		}

		var err = fillEnvFieldValue(fieldValue, envName, envDefaultValue)
		if err != nil {
			return errors.Wrapf(err, "failed to fill field %s", fieldDescriptor.Name)
		}
	}

	return nil
}

func fillEnvFieldValue(fieldValue reflect.Value, envName, envDefaultValue string) error {
	var envValue, existed = os.LookupEnv(envName)
	if !existed {
		envValue = envDefaultValue
	}
	if envValue == "" {
		// quickly return
		return nil
	}

	var valueType = fieldValue.Type()
	switch valueType.Kind() {
	case reflect.Ptr:
		var ptrType = valueType.Elem()
		var ptrValue = reflect.New(ptrType)
		if err := fillEnvFieldValue(ptrValue.Elem(), envName, envDefaultValue); err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to bool", envName, envValue)
		}
		fieldValue.Set(ptrValue)
	case reflect.String:
		fieldValue.SetString(envValue)
	case reflect.Bool:
		var ret, err = strconv.ParseBool(envValue)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to bool", envName, envValue)
		}
		fieldValue.SetBool(ret)
	case reflect.Float32:
		var ret, err = strconv.ParseFloat(envValue, 32)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to float32", envName, envValue)
		}
		fieldValue.SetFloat(ret)
	case reflect.Float64:
		var ret, err = strconv.ParseFloat(envValue, 64)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to float64", envName, envValue)
		}
		fieldValue.SetFloat(ret)
	case reflect.Uint64:
		var ret, err = strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to uint64", envName, envValue)
		}
		fieldValue.SetUint(ret)
	case reflect.Uint8:
		var ret, err = strconv.ParseUint(envValue, 10, 8)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to uint8", envName, envValue)
		}
		fieldValue.SetUint(ret)
	case reflect.Uint16:
		var ret, err = strconv.ParseUint(envValue, 10, 16)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to uint16", envName, envValue)
		}
		fieldValue.SetUint(ret)
	case reflect.Uint, reflect.Uint32:
		var ret, err = strconv.ParseUint(envValue, 10, 32)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to uint/uint32", envName, envValue)
		}
		fieldValue.SetUint(ret)
	case reflect.Int64:
		switch fieldValue.Interface().(type) {
		case time.Duration:
			var duration, err = time.ParseDuration(envValue)
			if err != nil {
				return errors.Wrapf(err, "failed to parse env %s value %q to time.Duration", envName, envValue)
			}
			fieldValue.SetInt(int64(duration))
		default:
			var ret, err = strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "failed to parse env %s value %q to int64", envName, envValue)
			}
			fieldValue.SetInt(ret)
		}
	case reflect.Int8:
		var ret, err = strconv.ParseInt(envValue, 10, 8)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to int8", envName, envValue)
		}
		fieldValue.SetInt(ret)
	case reflect.Int16:
		var ret, err = strconv.ParseInt(envValue, 10, 16)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to int16", envName, envValue)
		}
		fieldValue.SetInt(ret)
	case reflect.Int, reflect.Int32:
		var ret, err = strconv.ParseInt(envValue, 10, 32)
		if err != nil {
			return errors.Wrapf(err, "failed to parse env %s value %q to int/int32", envName, envValue)
		}
		fieldValue.SetInt(ret)
	}

	return nil
}
