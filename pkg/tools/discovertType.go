package tools

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

const GetTypeUnHandledErrMsg = "UnHandled Type"

func GetOpenApiType(t any) (string, error) {
	switch v := reflect.ValueOf(t); v.Kind() {
	case reflect.Bool:
		return fmt.Sprint("boolean"), nil
	case reflect.String:
		return fmt.Sprint("string"), nil
	case reflect.Ptr:
		if v.IsNil() {
			fmt.Printf("##kind %v : Uninitialized Pointer to type: %v\n", v.Kind(), v.Type().Elem())
		} else {
			fmt.Printf("##kind %v : Pointer to value: >>%v<< of type %v , element kind: %v \n", v.Kind(), v.Elem(), v.Type().Elem(), v.Type().Elem().Kind())
		}
		switch v.Type().Elem().Kind() {
		case reflect.String:
			return fmt.Sprint("string, nullable"), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return fmt.Sprint("integer, nullable"), nil
		case reflect.Float32, reflect.Float64:
			return fmt.Sprint("number, nullable"), nil
		case reflect.Struct:
			return fmt.Sprint("object, nullable"), nil
		default:
			log.Printf("WARNING: GetOpenApiType(%v[%T]) unhandled case result : Pointer to %s", v, v, reflect.TypeOf(t))
			return fmt.Sprintf("%s", reflect.TypeOf(t)), errors.New(GetTypeUnHandledErrMsg)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint("integer"), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprint("number"), nil
	case reflect.Array:
		return fmt.Sprint("array"), nil
	case reflect.Struct:
		return fmt.Sprint("object"), nil
	default:
		log.Printf("WARNING: GetOpenApiType(%v[%T]) unhandled case result : %s", v, v, reflect.TypeOf(t))
		return fmt.Sprintf("%s", reflect.TypeOf(t)), errors.New(GetTypeUnHandledErrMsg)
	}
}
