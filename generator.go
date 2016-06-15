package tools

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/graphql-go/graphql"
)

type Resolver interface {
	IsResolve(sourceType reflect.Type, field reflect.StructField) bool
	Resolve(FieldInfo, graphql.ResolveParams) (interface{}, error)
}
type Generator struct {
	Types    map[reflect.Type]graphql.Output
	Resolver Resolver
}

func NewGenerator(resolver Resolver) *Generator {
	generator := Generator{}
	generator.Types = map[reflect.Type]graphql.Output{}
	generator.Resolver = resolver
	return &generator
}
func (generator *Generator) Generate(typ interface{}) interface{} {
	return generator._GenerateGraphqlObject(typ)
}
func (generator *Generator) GenerateObject(typ interface{}) *graphql.Object {
	return generator.Generate(typ).(*graphql.Object)
}
func (generator *Generator) generateInterface() {

}

type FieldInfo struct {
	Name       string
	Type       reflect.Type
	Source     interface{}
	Args       interface{}
	ResolveTag string
}

func (generator *Generator) _GenerateGraphqlObject(source interface{}) graphql.Output {
	types := generator.Types

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

	//If type is interface
	IsInterface := false
	if method, ok := sourceType.MethodByName("IsInterface"); ok {
		IsInterface = method.Func.Call([]reflect.Value{reflect.ValueOf(source)})[0].Interface().(bool)
	}

	//get fields
	var graphqlFields = graphql.Fields{} //init graphql fields
	var graphqlInterfaces = []*graphql.Interface{}
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

		graphqlField.Type = generator.getGraphQLType(fieldType, graphqlTagType)

		if graphqlField.Type == nil {
			continue
		}
		/*if graphqlTagType == "globalid" {
			graphqlField.Resolve = getGlobalIdResolveFunc(sourceType.Name(), fieldName)
		}*/

		/*if method, ok := sourceType.MethodByName("Resolve" + fieldName); ok {
			graphqlField.Resolve = getResolveFunc(sourceType, method.Func)
		}*/

		///Args
		var argsI interface{} = nil
		if method, ok := sourceType.MethodByName("ArgsFor" + fieldName); ok {
			args := method.Func.Call([]reflect.Value{reflect.ValueOf(source)})[0]
			graphqlField.Args = generator.getArgs(args)
			argsI = args.Interface()
		}

		//Resolve
		resolveTag := sourceType.Field(i).Tag.Get("resolve")
		if generator.Resolver != nil && generator.Resolver.IsResolve(sourceType, field) {
			fieldInfo := FieldInfo{
				Name:       fieldName,
				Type:       fieldType,
				Source:     source,
				Args:       argsI,
				ResolveTag: resolveTag,
			}
			graphqlField.Resolve = func(p graphql.ResolveParams) (interface{}, error) {
				return generator.Resolver.Resolve(fieldInfo, p)
			} //getResolveFunc(sourceType, reflect.ValueOf(route)) // getResolveFunc(sourceType, reflect.ValueOf(route))
		}

		/*if route, ok := routes[name+"."+fieldName]; ok || resolveTag != "" {


		}*/

		graphqlField.Name = lA(fieldName)
		if field.Anonymous {
			IsInterface = false
			graphqlInterfaces = append(graphqlInterfaces, graphqlField.Type.(*graphql.Interface))
		} else {
			graphqlFields[lA(fieldName)] = graphqlField
		}

	}

	config := graphql.ObjectConfig{
		Name:        name,
		Description: description,
	}
	if len(graphqlFields) > 0 {
		config.Fields = graphqlFields
	}
	if len(graphqlInterfaces) > 0 {
		config.Interfaces = graphqlInterfaces
	}

	var obj graphql.Output
	if IsInterface {
		obj = graphql.NewInterface(graphql.InterfaceConfig{
			Name:        config.Name,
			Fields:      config.Fields,
			Description: config.Description,
			ResolveType: generator.ResolveType,
		})
	} else {
		obj = graphql.NewObject(config)
	}

	types[sourceType] = obj

	return obj
}
func (generator *Generator) ResolveType(p graphql.ResolveTypeParams) *graphql.Object {
	return generator.Types[reflect.TypeOf(p.Value)].(*graphql.Object)
}

/**/
func (generator *Generator) getArgs(sourceValue reflect.Value) graphql.FieldConfigArgument {
	args := graphql.FieldConfigArgument{}
	sourceType := sourceValue.Type()
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		if !sourceValue.Field(i).CanInterface() {
			continue
		}
		args[lA(field.Name)] = &graphql.ArgumentConfig{
			Type:         generator.getGraphQLType(field.Type, field.Tag.Get("graphql")),
			Description:  field.Name,
			DefaultValue: sourceValue.Field(i).Interface(),
		}
	}
	return args
}
func (generator *Generator) getGraphQLType(fieldType reflect.Type, graphqlTagType string) graphql.Output {
	types := generator.Types
	if graphqlTagType == "globalid" {
		return graphql.ID
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
			return generator._GenerateGraphqlObject(reflect.New(fieldType).Elem().Interface())
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
		t := generator.getGraphQLType(fieldType.Elem(), graphqlTagType)
		graphqlType = graphql.NewList(t)
	default:

		return nil

	}

	if !isNull {
		return graphql.NewNonNull(graphqlType)
	}
	return graphqlType
}

/*func getResolveFunc(sourceType reflect.Type, fun reflect.Value) func(p graphql.ResolveParams) (interface{}, error) {
	funcArgsTypes := []string{"", "", "", ""}
	funArgsNum := fun.Type().NumIn()
	if funArgsNum == 0 {
		panic("Invalid resolve func, expected 1 parameter to be graphql.ResolveParams or " + sourceType.Name())
	}
	if fun.Type().In(0) == reflect.TypeOf(graphql.ResolveParams{}) {
		funcArgsTypes[0] = "params"
	} else {
		funcArgsTypes[0] = "obj"
	}
	if funArgsNum > 1 {
		if fun.Type().In(1) == reflect.TypeOf(graphql.ResolveParams{}) {
			funcArgsTypes[1] = "params"
		} else {
			funcArgsTypes[1] = "args"
		}
	}
	if funArgsNum > 2 {
		funcArgsTypes[2] = "context"
	}
	return func(p graphql.ResolveParams) (interface{}, error) {
		var source reflect.Value
		if reflect.TypeOf(p.Source).Kind() == reflect.Map {
			source = reflect.New(sourceType).Elem()
		} else {
			source = reflect.ValueOf(p.Source)
		}

		argumentsForCall := []reflect.Value{}
		for _, funcArgsType := range funcArgsTypes {
			switch funcArgsType {
			case "params":
				argumentsForCall = append(argumentsForCall, reflect.ValueOf(p))
				break
			case "obj":
				argumentsForCall = append(argumentsForCall, source)
				break
			case "args":
				argumentsForCall = append(argumentsForCall, getArgsForResolve(p.Args, fun.Type().In(1)))
				break
			case "context":
				argumentsForCall = append(argumentsForCall, getContextForResolve(p.Context, fun.Type().In(2)))
				break
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
}*/

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
