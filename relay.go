package tools

import (
	"reflect"

	"github.com/graphql-go/relay"
)

func ResolveGlobalId(params ResolveParams) (interface{}, error) {
	var rawId reflect.Value

	switch reflect.TypeOf(params.Params.Source).Kind() {
	case reflect.Struct:
		rawId = reflect.ValueOf(params.Params.Source).FieldByName(params.FieldInfo.Name)
	case reflect.Map:
		rawId = reflect.ValueOf(params.Params.Source).MapIndex(reflect.ValueOf(params.FieldInfo.Name))
	}
	if rawId.Interface() == nil {
		return nil, nil
	}
	return relay.ToGlobalID(reflect.TypeOf(params.Params.Source).Name(), rawId.Interface().(string)), nil

}
