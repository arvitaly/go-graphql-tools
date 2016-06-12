package generator

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"golang.org/x/net/context"
)

func _GenerateGraphqlObject(source interface{}, types map[reflect.Type]*graphql.Object, routes map[string]interface{}) *graphql.Object {
	sourceType := reflect.TypeOf(source)
	//get name
	var name = sourceType.Name()

	//get description
	method, ok := sourceType.MethodByName("Description")
	var description string
	if ok {
		description = method.Func.Call([]reflect.Value{reflect.ValueOf(source)})[0].Interface().(string)
	} else {
		description = name
	}
	//get fields
	var graphqlFields = graphql.Fields{} //init graphql fields
	for i := 0; i < sourceType.NumField(); i++ {
		var sourceFieldGraphqlTag = sourceType.Field(i).Tag.Get("graphql")
		if sourceFieldGraphqlTag == "-" {
			continue
		}
		sourceFieldGraphqlTagParams := strings.Split(sourceFieldGraphqlTag, ",")
		var graphqlTagType string
		if len(sourceFieldGraphqlTagParams) > 0 {
			graphqlTagType = strings.ToLower(sourceFieldGraphqlTagParams[0])
		}

		//init new field
		var graphqlField = &graphql.Field{}

		//
		var field = sourceType.Field(i)
		var fieldType = field.Type
		var fieldName = field.Name

		graphqlField.Type = getGraphQLType(fieldType, graphqlTagType, types, routes)

		if graphqlField.Type == nil {
			continue
		}
		if graphqlTagType == "globalid" {
			graphqlField.Resolve = getGlobalIdResolveFunc(sourceType.Name(), fieldName)
		}
		//Resolve
		if route, ok := routes[name+"."+fieldName]; ok {

			graphqlField.Resolve = getResolveFunc(sourceType, reflect.ValueOf(route))
		}
		if method, ok := sourceType.MethodByName("Resolve" + fieldName); ok {
			graphqlField.Resolve = getResolveFunc(sourceType, method.Func)
		}

		///Args
		if method, ok := sourceType.MethodByName("ArgsFor" + fieldName); ok {

			graphqlField.Args = getArgs(method.Func.Call([]reflect.Value{reflect.ValueOf(source)})[0], types, routes)

		}

		graphqlFields[lA(fieldName)] = graphqlField
	}
	config := graphql.ObjectConfig{
		Name:        name,
		Description: description,
	}
	if len(graphqlFields) > 0 {
		config.Fields = graphqlFields
	}
	obj := graphql.NewObject(config)
	types[sourceType] = obj

	return obj
}

func getGlobalIdResolveFunc(typeName string, fieldName string) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		var rawId interface{}
		switch reflect.TypeOf(p.Source).Kind() {
		case reflect.Struct:
			rawId = reflect.ValueOf(p.Source).FieldByName(fieldName).Interface()
		case reflect.Map:
			rawId = reflect.ValueOf(p.Source).MapIndex(reflect.ValueOf(fieldName)).Interface()
		}
		if rawId != nil {
			return relay.ToGlobalID(typeName, rawId.(string)), nil
		}
		return nil, nil
	}
}
func getArgs(sourceValue reflect.Value, types map[reflect.Type]*graphql.Object, routes map[string]interface{}) graphql.FieldConfigArgument {
	args := graphql.FieldConfigArgument{}
	sourceType := sourceValue.Type()
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		if !sourceValue.Field(i).CanInterface() {
			continue
		}
		args[lA(field.Name)] = &graphql.ArgumentConfig{
			Type:         getGraphQLType(field.Type, "", types, routes),
			Description:  field.Name,
			DefaultValue: sourceValue.Field(i).Interface(),
		}
	}
	return args
}
func getGraphQLType(fieldType reflect.Type, graphqlTagType string, types map[reflect.Type]*graphql.Object, routes map[string]interface{}) graphql.Output {
	if graphqlTagType == "globalid" {
		return graphql.NewNonNull(graphql.ID)
	}
	if graphqlTagType == "id" {
		return graphql.ID
	}
	if graphqlTagType == "enum" {
		var configMap = graphql.EnumValueConfigMap{}

		if method, ok := fieldType.MethodByName("Values"); ok {
			res := method.Func.Call([]reflect.Value{reflect.New(fieldType).Elem()})
			for _, key := range res[0].MapKeys() {
				configMap[key.String()] = &graphql.EnumValueConfig{Value: res[0].MapIndex(key).Interface()}
			}
		}
		return graphql.NewEnum(graphql.EnumConfig{
			Name:   fieldType.Name(),
			Values: configMap})

	}
	fieldKind := fieldType.Kind()
	var isNull = false
	if fieldKind == reflect.Ptr {
		isNull = true
		fieldKind = fieldType.Elem().Kind()
		fieldType = fieldType.Elem()
	}

	kind := fieldType.Kind()
	if kind == reflect.Struct {
		if fieldObj, ok := types[fieldType]; ok {
			return fieldObj
		} else {
			return _GenerateGraphqlObject(reflect.New(fieldType).Elem().Interface(), types, routes)
		}
	}
	var graphqlType graphql.Output
	switch kind {
	case reflect.String:
		graphqlType = graphql.String
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint:
		graphqlType = graphql.Int
	case reflect.Float32, reflect.Float64:
		graphqlType = graphql.Float
	case reflect.Bool:
		graphqlType = graphql.Boolean
	case reflect.Slice:
		t := getGraphQLType(fieldType.Elem(), graphqlTagType, types, routes)
		graphqlType = graphql.NewList(t)
	default:

		return nil

	}

	if !isNull {
		return graphql.NewNonNull(graphqlType)
	}
	return graphqlType
}

func getArgsForResolve(args map[string]interface{}, typ reflect.Type) reflect.Value {
	var output = reflect.New(typ)
	for key, value := range args {
		n := lU(key)
		if _, ok := typ.FieldByName(n); ok {
			field := output.Elem().FieldByName(n)
			if field.CanInterface() {

				if field.Kind() == reflect.Ptr {
					field.Set(reflect.New(field.Type().Elem()))
					field.Elem().Set(reflect.ValueOf(value))
				} else {
					field.Set(reflect.ValueOf(value))
				}

			}
		}
	}
	return output.Elem()
}
func getContextForResolve(context context.Context, typ reflect.Type) reflect.Value {
	var output = reflect.New(typ)

	for i := 0; i < typ.NumField(); i++ {
		if !output.Elem().Field(i).CanInterface() {
			continue
		}
		value := context.Value(lA(typ.Field(i).Name))
		if value == nil {
			continue
		}
		output.Elem().Field(i).Set(reflect.ValueOf(value))
	}

	return output.Elem()
}
func getResolveFunc(sourceType reflect.Type, fun reflect.Value) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {

		var source reflect.Value
		if reflect.TypeOf(p.Source).Kind() == reflect.Map {
			source = reflect.New(sourceType).Elem()
		} else {
			source = reflect.ValueOf(p.Source)
		}

		argumentsForCall := []reflect.Value{source}

		if fun.Type().NumIn() > 1 {
			if fun.Type().In(1) == reflect.TypeOf(graphql.ResolveParams{}) {
				argumentsForCall = append(argumentsForCall, reflect.ValueOf(p))
			} else {
				//args
				argumentsForCall = append(argumentsForCall, getArgsForResolve(p.Args, fun.Type().In(1)))

				//context
				if fun.Type().NumIn() > 2 {
					argumentsForCall = append(argumentsForCall, getContextForResolve(p.Context, fun.Type().In(2)))
				}
			}
		}

		values := fun.Call(argumentsForCall)
		if len(values) != 2 {
			panic("Resolve func should return 2 values: interface{}, error")
		}
		err := values[1].Interface()

		var ret interface{}
		retType := values[0].Type()
		if values[0].Kind() == reflect.Struct {

			ret2 := map[string]interface{}{}
			for i := 0; i < retType.NumField(); i++ {

				//Check for exported field
				if !values[0].Field(i).CanInterface() {
					continue
				}

				if retType.Field(i).Type.Kind() == reflect.Ptr {

					if !values[0].Field(i).IsNil() {

						ret2[lA(retType.Field(i).Name)] = values[0].Field(i).Elem().Interface()
					}

				} else {

					ret2[lA(retType.Field(i).Name)] = values[0].Field(i).Interface()

				}

			}
			ret = ret2

		} else {

			if values[0].Kind() == reflect.Ptr {
				ret = values[0].Elem().Interface()
			} else {
				ret = values[0].Interface()
			}

		}

		if err == nil {
			return ret, nil
		} else {

			return ret, values[1].Interface().(error)
		}

	}
}
func GenerateGraphqlObject(typ interface{}, routes *map[string]interface{}) *graphql.Object {
	types := map[reflect.Type]*graphql.Object{}
	if routes != nil {
		return _GenerateGraphqlObject(typ, types, *routes)
	} else {
		return _GenerateGraphqlObject(typ, types, map[string]interface{}{})
	}

}
func lA(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}
func lU(s string) string {
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}
