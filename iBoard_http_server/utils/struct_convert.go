package utils

import (
	"reflect"
	"unicode"
)

func CamelToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func GetLowerCamelCase(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

type S2MKeyType int

const (
	S2MKeyTypeSnakeCase      S2MKeyType = 1
	S2MKeyTypeLowerCamelCase S2MKeyType = 2
)

func StructToMap(s interface{}, keyTypeOptional ...S2MKeyType) map[string]interface{} {
	keyType := S2MKeyTypeSnakeCase
	if len(keyTypeOptional) > 0 {
		keyType = S2MKeyType(keyTypeOptional[0])
	}

	result := make(map[string]interface{})
	val := reflect.ValueOf(s)

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		valueField := val.Field(i)

		// pass if tag s2m is "-"
		if typeField.Tag.Get("s2m") == "-" {
			continue
		}
		// handle bool
		if valueField.Kind() == reflect.Bool {
			result[CamelToSnakeCase(typeField.Name)] = valueField.Interface()
			continue
		}
		// handle Zero value
		if typeField.Tag.Get("zero") != "-" && valueField.Interface() == reflect.Zero(valueField.Type()).Interface() {
			continue
		}

		var key string
		if typeField.Tag.Get("form") != "" {
			key = typeField.Tag.Get("form")
		} else if typeField.Tag.Get("json") != "" {
			key = typeField.Tag.Get("json")
		} else {
			key = typeField.Name
		}

		// if pointer, get the value
		if valueField.Kind() == reflect.Ptr {
			valueField = valueField.Elem()
		}

		if keyType == S2MKeyTypeSnakeCase {
			result[CamelToSnakeCase(key)] = valueField.Interface()
		} else if keyType == S2MKeyTypeLowerCamelCase {
			result[GetLowerCamelCase(key)] = valueField.Interface()
		} else {
			return nil
		}
	}

	return result
}

func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
