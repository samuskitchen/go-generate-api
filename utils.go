package generate

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/samuskitchen/go-message-tool/errors"
)

func jsonBind(c echo.Context, payload interface{}) error {
	if err := (&echo.DefaultBinder{}).BindBody(c, payload); err != nil {
		return errores.NewBadRequest(nil, errores.ErrInvalidJSON)
	}

	return nil
}

func search(collection []string, s string) bool {
	for _, item := range collection {
		if item == s {
			return true
		}
	}

	return false
}

func cameCaseToSnake(s string) string {
	for _, reStr := range []string{`([A-Z]+)([A-Z][a-z])`, `([a-z\d])([A-Z])`} {
		re := regexp.MustCompile(reStr)
		s = re.ReplaceAllString(s, "${1}_${2}")
	}

	return strings.ToLower(s)
}

func getMapJsonFieldNameWithModelFieldName(entry interface{}, ignore ...string) map[string]string {
	var responseMap map[string]string = make(map[string]string)

	reflectType := reflect.TypeOf(entry)
	if reflectType == nil {
		return responseMap
	}

	if reflectType.Kind() == reflect.Ptr {
		if reflectType.Elem().Kind() == reflect.Struct {
			reflectType = reflectType.Elem()
		} else {
			return responseMap
		}
	}

	for i := 0; i < reflectType.NumField(); i++ {
		rsf := reflectType.Field(i)
		jsonValue := rsf.Tag.Get("json")
		if jsonValue == "" {
			continue
		}

		jsonValue = strings.Split(jsonValue, ",")[0]
		if !search(ignore, jsonValue) {
			responseMap[jsonValue] = cameCaseToSnake(rsf.Name)
		}
	}

	return responseMap
}

func getIdentifierValues(value reflect.Value) (interface{}, error) {
	switch value.Type().Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		n, ok := (value.Interface()).(int)
		if !ok {
			err := fmt.Errorf("%s, error assertion int", value.Type().Name())
			return nil, errores.NewInternal(err, errores.ErrDatabaseInternal)
		}

		if n == 0 {
			return nil, errores.NewNotFound(nil, "null identifier error")
		}

		return n, nil

	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		n, ok := (value.Interface()).(uint)
		if !ok {
			err := fmt.Errorf("%s, error assertion uint", value.Type().Name())
			return nil, errores.NewInternal(err, errores.ErrDatabaseInternal)
		}

		if n == 0 {
			return nil, errores.NewNotFound(nil, "null identifier error")
		}

		return n, nil

	case reflect.String:
		n, ok := (value.Interface()).(string)
		if !ok {
			err := fmt.Errorf("%s, error assertion string", value.Type().Name())
			return nil, errores.NewInternal(err, errores.ErrDatabaseInternal)
		}

		if n == "" {
			return nil, errores.NewNotFound(nil, "null identifier error")
		}

		return n, nil

	default:
		err := fmt.Errorf("%s, wrong data type", value.Type().Name())
		return nil, errores.NewInternal(err, errores.ErrDatabaseInternal)
	}
}
