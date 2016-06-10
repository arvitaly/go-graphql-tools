package generator

import (
	"reflect"
	"unicode"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

type RelayGlobalLID struct{}

func _GenerateGraphqlObject(source interface{}, types map[reflect.Type]*graphql.Object) *graphql.Object {
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

		//init new field
		var graphqlField = &graphql.Field{}
		var isNull = false
		//
		var field = sourceType.Field(i)
		var fieldType = field.Type
		var fieldName = field.Name
		var fieldKind = fieldType.Kind()

		if fieldType == reflect.TypeOf(RelayGlobalLID{}) {
			graphqlField = relay.GlobalIDField(sourceType.Field(i).Name, nil)
			continue
		} else {

			if fieldKind == reflect.Ptr {
				isNull = true
				fieldKind = fieldType.Elem().Kind()
				fieldType = fieldType.Elem()
			}
			graphqlField.Type = getTypeByKind(field, fieldType, types)
			if graphqlField.Type == nil {
				continue
			}

			if !isNull {
				graphqlField.Type = graphql.NewNonNull(graphqlField.Type)
			}

			//Resolve
			if method, ok := sourceType.MethodByName("Resolve" + fieldName); ok {
				graphqlField.Resolve = getResolveFunc(sourceType, method)
			}

			///Args
			if _, ok := sourceType.MethodByName("ArgsFor" + fieldName); ok {

			}

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
func getTypeByKind(field reflect.StructField, fieldType reflect.Type, types map[reflect.Type]*graphql.Object) graphql.Output {
	kind := fieldType.Kind()
	if kind == reflect.Struct {
		if fieldObj, ok := types[fieldType]; ok {
			return fieldObj
		} else {
			return _GenerateGraphqlObject(reflect.New(fieldType).Elem().Interface(), types)
		}
	}
	switch kind {
	case reflect.String:
		return graphql.String
	case reflect.Int, reflect.Int32, reflect.Int64:
		return graphql.Int
	case reflect.Float32, reflect.Float64:
		return graphql.Float
	case reflect.Bool:
		return graphql.Boolean
	case reflect.Slice:
		t := getTypeByKind(field, fieldType.Elem(), types)
		return graphql.NewList(t)
	default:

		return nil

	}
}
func getResolveFunc(sourceType reflect.Type, method reflect.Method) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {

		var source reflect.Value
		if reflect.TypeOf(p.Source).Kind() == reflect.Map {
			source = reflect.New(sourceType).Elem()
		} else {
			source = reflect.ValueOf(p.Source)
		}

		values := method.Func.Call([]reflect.Value{source, reflect.ValueOf(p)})
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
func GenerateGraphqlObject(typ interface{}) *graphql.Object {
	types := map[reflect.Type]*graphql.Object{}
	return _GenerateGraphqlObject(typ, types)
}
func lA(s string) string {
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}
