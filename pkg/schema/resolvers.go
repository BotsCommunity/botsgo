package schema

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/botscommunity/botsgo/pkg/converter"
)

const Integer = "int"
const Float = "float"
const String = "string"
const ArrayInteger = "[]int"
const ArrayString = "[]string"
const Boolean = "bool"
const Duration = "time.Time"
const Struct = "struct"

type IType interface {
	doQuery(query url.Values, name string, property any)
}

type IntegerType struct{}

func (i IntegerType) doQuery(query url.Values, name string, property any) {
	query.Set(name, fmt.Sprint(property))
}

type FloatType struct{}

func (i FloatType) doQuery(query url.Values, name string, property any) {
	query.Set(name, fmt.Sprint(property))
}

type StringType struct{}

func (i StringType) doQuery(query url.Values, name string, property any) {
	query.Set(name, fmt.Sprint(property))
}

type ArrayIntegerType struct{}

func (i ArrayIntegerType) doQuery(query url.Values, name string, property any) {
	query.Set(name, converter.SliceToString(property.([]int)))
}

type ArrayStringType struct{}

func (i ArrayStringType) doQuery(query url.Values, name string, property any) {
	query.Set(name, converter.SliceToString(property.([]string)))
}

type BooleanType struct{}

func (i BooleanType) doQuery(query url.Values, name string, property any) {
	query.Set(name, fmt.Sprint(property.(bool)))
}

type DurationType struct{}

func (i DurationType) doQuery(query url.Values, name string, property any) {
	query.Set(name, fmt.Sprint(property))
}

type StructType struct{}

func (i StructType) doQuery(query url.Values, _ string, property any) {
	for i := 0; i < reflect.TypeOf(property).NumField(); i++ {
		name, value := structField(
			reflect.TypeOf(property).Field(i),
			reflect.ValueOf(property).Field(i),
		)

		if name != "" && value != "" {
			query.Set(name, value)
		}
	}
}

func structField(typ reflect.StructField, val reflect.Value) (string, string) {
	if tag := typ.Tag.Get("json"); tag != "" {
		switch typ.Type.Kind().String() {
		case Integer:
			if intValue := val.Int(); intValue != 0 {
				return tag, fmt.Sprint(intValue)
			}
		case String:
			if strValue := val.String(); strValue != "" {
				return tag, strValue
			}
		case Boolean:
			if boolValue := val.Bool(); boolValue {
				switch typ.Tag.Get("to") {
				case "int":
					return tag, fmt.Sprint(converter.BooleanToInteger(boolValue))
				default:
					return tag, fmt.Sprint(boolValue)
				}
			}
		case "slice":
			return sliceField(tag, typ, val)
		case Struct:
			numField := val.NumField()

			return structField(
				reflect.TypeOf(val.Interface()).Field(numField),
				reflect.ValueOf(val.Interface()).Field(numField),
			)
		}
	}

	return "", ""
}

func sliceField(tag string, typ reflect.StructField, val reflect.Value) (string, string) {
	switch typ.Type.String() {
	case "[]int":
		if slice, ok := val.Interface().([]int); ok && len(slice) > 0 {
			return tag, converter.SliceToString(slice)
		}
	case "[]string":
		if slice, ok := val.Interface().([]string); ok && len(slice) > 0 {
			return tag, converter.SliceToString(slice)
		}
	}

	return "", ""
}
