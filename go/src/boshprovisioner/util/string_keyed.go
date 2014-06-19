package util

import (
	"reflect"

	bosherr "bosh/errors"
)

type stringKeyed struct{}

func NewStringKeyed() stringKeyed { return stringKeyed{} }

// convertMapToStringKeyMap converts map to string keyed map.
func (sk stringKeyed) ConvertMap(m map[interface{}]interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	for name, val := range m {
		nameStr, ok := name.(string)
		if !ok {
			return result, bosherr.New("Map contains non-string key %v", name)
		}

		convertedVal, err := sk.ConvertInterface(val)
		if err != nil {
			return result, err
		}

		result[nameStr] = convertedVal
	}

	return result, nil
}

func (sk stringKeyed) ConvertInterface(val interface{}) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.Map:
		valMap, ok := val.(map[interface{}]interface{})
		if !ok {
			return nil, bosherr.New("Converting map %v", val)
		}

		return sk.ConvertMap(valMap)

	case reflect.Slice:
		valSlice, ok := val.([]interface{})
		if !ok {
			return nil, bosherr.New("Converting slice %v", val)
		}

		slice := make([]interface{}, len(valSlice))

		for i, v := range valSlice {
			convertedVal, err := sk.ConvertInterface(v)
			if err != nil {
				return nil, err
			}

			slice[i] = convertedVal
		}

		return slice, nil

	default:
		return val, nil
	}
}
