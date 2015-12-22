package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/spf13/viper"
)

func tmplFuncs(v *viper.Viper) template.FuncMap {
	return template.FuncMap{
		"join":       strings.Join,
		"concat":     concat,
		"sum":        sum,
		"flag":       flag(v),
		"flagString": flagString(v),
		"dockerLink": dockerLink,
	}
}

func concat(items ...interface{}) string {
	result := ""
	for _, v := range items {
		if s, ok := v.(string); ok {
			result += s
		} else {
			return ""
		}
	}
	return result
}

func sum(items ...interface{}) (int, error) {
	result := 0
	for _, v := range items {
		i, err := toInt(v)
		if err != nil {
			return 0, maskAny(err)
		}
		result += i
	}
	return result, nil
}

func toInt(v interface{}) (int, error) {
	switch v.(type) {
	case string:
		i, err := strconv.Atoi(reflect.ValueOf(v).String())
		if err != nil {
			return 0, maskAny(err)
		}
		return i, nil
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int()), nil
	case float32, float64:
		return int(reflect.ValueOf(v).Float()), nil
	}
	return 0, nil
}

func flag(v *viper.Viper) func(string) (interface{}, error) {
	return func(flag string) (interface{}, error) {
		if v.IsSet(flag) {
			return v.Get(flag), nil
		}
		return nil, maskAny(flagNotFoundError)
	}
}

func flagString(v *viper.Viper) func(string) (string, error) {
	return func(flag string) (string, error) {
		if v.IsSet(flag) {
			return v.GetString(flag), nil
		}
		return "", maskAny(flagNotFoundError)
	}
}

func dockerLink(service, alias string) string {
	// TODO fix service name when scaled
	return fmt.Sprintf("--link %s.service:%s", service, alias)
}

func dict(items ...interface{}) (map[string]interface{}, error) {
	if len(items)%2 != 0 {
		return nil, maskAny(errors.New("invalid dict call"))
	}
	dict := make(map[string]interface{}, len(items)/2)
	for i := 0; i < len(items); i += 2 {
		key, ok := items[i].(string)
		if !ok {
			return nil, maskAny(errors.New("dict keys must be strings"))
		}
		dict[key] = items[i+1]
	}
	return dict, nil
}
