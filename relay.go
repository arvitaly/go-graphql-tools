package tools

import (
	"reflect"

	"github.com/graphql-go/relay"
)

func UseGlobalId(params ResolveParams) (interface{}, bool, error) {

	if params.FieldInfo.ResolveTag == "globalid" {
		res, err := ResolveGlobalId(params)
		if err != nil {
			return nil, false, err
		}
		return res, false, nil
	}
	return nil, true, nil
}
func ResolveGlobalId(params ResolveParams) (interface{}, error) {
	var rawId interface{}

	switch reflect.TypeOf(params.Params.Source).Kind() {
	case reflect.Struct:
		rawId = reflect.ValueOf(params.Params.Source).FieldByName(params.FieldInfo.Name).Interface()
	case reflect.Map:
		rawId = reflect.ValueOf(params.Params.Source).MapIndex(reflect.ValueOf(params.FieldInfo.Name)).Interface()
	}

	if rawId != nil {
		return relay.ToGlobalID(reflect.TypeOf(params.Params.Source).Name(), rawId.(string)), nil
	}
	return nil, nil
}
