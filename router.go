package tools

import (
	"errors"
	"log"
	"reflect"

	"golang.org/x/net/context"

	"github.com/graphql-go/graphql"
)
import "github.com/graphql-go/graphql/language/ast"

const (
	ArgTypeParams  = 1
	ArgTypeSource  = 2
	ArgTypeArgs    = 3
	ArgTypeContext = 4
)

type ResolveParams struct {
	FieldInfo FieldInfo
	Source    interface{}
	Args      interface{}
	Context   interface{}
	Params    graphql.ResolveParams
}

type Router struct {
	queries map[string]RouteParams
	uses    []UseFn
}
type RouteParams struct {
	Args    []int
	Context reflect.Type
	Handle  interface{}
}

func NewRouter() *Router {
	router := Router{
		queries: map[string]RouteParams{},
		uses:    []UseFn{},
	}
	return &router
}

func (r *Router) Routes() map[string]interface{} {
	routes := map[string]interface{}{}
	for k, _ := range r.queries {
		routes[k] = r.Resolve
	}
	return routes
}
func (r *Router) IsResolve(sourceType reflect.Type, field reflect.StructField) bool {
	path := sourceType.Name() + "." + field.Name
	if _, ok := r.queries[path]; ok {
		return true
	}
	if field.Tag.Get("resolve") != "" {
		return true
	}
	return false
}
func (r *Router) Resolve(fieldInfo FieldInfo, p graphql.ResolveParams) (interface{}, error) {

	sourceType := reflect.TypeOf(fieldInfo.Source)
	sourceValueType := reflect.TypeOf(p.Source)
	//Change ptr to elem
	if sourceType.Kind() == reflect.Ptr {
		sourceType = sourceType.Elem()
	}
	var source interface{}
	//Check type of source
	if sourceValueType.Kind() != reflect.Struct {
		if sourceValueType.Kind() == reflect.Map {

			source = reflect.New(sourceType).Elem().Interface()
		} else {
			return nil, InvalidSourceError{RouterError{Text: "Source for resolve query should be struct or pointer to struct, has " + sourceType.Kind().String()}}
		}
	} else {
		source = p.Source
	}

	for _, useFn := range r.uses {

		res, next, err := useFn(ResolveParams{
			FieldInfo: fieldInfo,
			Params:    p,
			Source:    source,
		})

		if err != nil {
			return nil, err
		}
		if !next {
			return res, nil
		}
	}

	if p.Info.Operation.GetOperation() == ast.OperationTypeQuery {
		res, err := r.ResolveQuery(fieldInfo, p)
		if err != nil {
			return nil, err
		}
		/*retType := reflect.TypeOf(res)
		ret := reflect.ValueOf(res)
		if retType.Kind() != reflect.Struct {
			if retType.Kind() == reflect.Ptr {
				res = ret.Elem().Interface()
			}
		}*/
		return res, nil
	}

	return nil, errors.New("Unsupported resolve")
}
func (r *Router) ResolveQuery(fieldInfo FieldInfo, p graphql.ResolveParams) (interface{}, error) {
	sourceType := reflect.TypeOf(fieldInfo.Source)
	sourceValueType := reflect.TypeOf(p.Source)
	//Change ptr to elem
	if sourceType.Kind() == reflect.Ptr {
		sourceType = sourceType.Elem()
	}
	var source interface{}
	//Check type of source
	if sourceValueType.Kind() != reflect.Struct {
		if sourceValueType.Kind() == reflect.Map {

			source = reflect.New(sourceType).Elem().Interface()
		} else {
			return nil, InvalidSourceError{RouterError{Text: "Source for resolve query should be struct or pointer to struct, has " + sourceType.Kind().String()}}
		}
	} else {
		source = p.Source
	}

	sourceTypeName := sourceType.Name()

	sourceFieldName := lU(p.Info.FieldName)

	path := sourceTypeName + "." + sourceFieldName

	query, ok := r.queries[path]

	if !ok {
		return nil, NotFoundRoute{RouterError{Text: "Not found route for path " + path}}
	}

	var args interface{}
	if fieldInfo.Args != nil {
		args = getArgsForResolve(p.Args, reflect.TypeOf(fieldInfo.Args)).Interface()
		log.Println(fieldInfo.Name, fieldInfo.Args, args)
	} else {
		log.Println(fieldInfo.Name, fieldInfo.Args)
		args = nil
	}

	params := ResolveParams{
		Source:    source,
		Args:      args,
		Context:   p.Context,
		FieldInfo: fieldInfo,
		Params:    p,
	}
	argsCall := []reflect.Value{}

	var contextCall reflect.Value
	if query.Context != nil {
		contextCall = getContextForResolve(p.Context, query.Context)
	} else {
		contextCall = reflect.ValueOf(p.Context)
	}

	for i := 0; i < len(query.Args); i++ {
		switch query.Args[i] {
		case ArgTypeSource:
			argsCall = append(argsCall, reflect.ValueOf(source))
			break
		case ArgTypeArgs:
			if args == nil {

				argsCall = append(argsCall, reflect.ValueOf(map[string]interface{}{}))
			} else {
				log.Println(777)
				argsCall = append(argsCall, reflect.ValueOf(args))
			}

			break
		case ArgTypeContext:

			argsCall = append(argsCall, contextCall)
			break
		case ArgTypeParams:
			log.Println("MMMM", params.Args)
			argsCall = append(argsCall, reflect.ValueOf(params))
			break
		}
	}
	handleValue := reflect.ValueOf(query.Handle)

	resValue := handleValue.Call(argsCall)
	if resValue[1].Interface() != nil {
		return nil, resValue[1].Interface().(error)
	}

	return resValue[0].Interface(), nil
}

func getArgsForResolve(args map[string]interface{}, typ reflect.Type) reflect.Value {

	var output = reflect.New(typ)
	for key, value := range args {
		n := lU(key)
		if _, ok := typ.FieldByName(n); ok {
			field := output.Elem().FieldByName(n)
			if field.CanInterface() {

				if field.Kind() == reflect.Ptr {
					v := reflect.ValueOf(value)

					if v.Type().Kind() == reflect.Ptr {
						field.Set(v)
					} else {
						field.Set(reflect.New(field.Type().Elem()))
						field.Elem().Set(v)
					}

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

type UseFn func(params ResolveParams) (interface{}, bool, error)

func (r *Router) Use(fn UseFn) {
	r.uses = append(r.uses, fn)
}
func (r *Router) Query(path string, handle interface{}) {
	handleType := reflect.TypeOf(handle)
	if handleType.Kind() != reflect.Func {
		panic("Invalid query handle, expected func, has " + handleType.Kind().String())
	}
	if handleType.NumOut() != 2 {
		panic("Invalid query handle, func should return 2 parameters interface, error")
	}
	args := []int{}
	current := 0
	params := RouteParams{}
	for i := 0; i < handleType.NumIn(); i++ {

		if handleType.In(i) == reflect.TypeOf(ResolveParams{}) {
			args = append(args, ArgTypeParams)
		} else {
			switch current {
			case 0:
				args = append(args, ArgTypeSource)
				break
			case 1:
				args = append(args, ArgTypeArgs)
				break
			case 2:
				params.Context = handleType.In(i)
				args = append(args, ArgTypeContext)
				break
			}
			current++
		}
	}
	params.Handle = handle
	params.Args = args
	r.queries[path] = params
}

type RouterError struct {
	Text string
}

type InvalidSourceError struct {
	RouterError
}
type NotFoundRoute struct {
	RouterError
}

func (e RouterError) Error() string {
	return e.Text
}
