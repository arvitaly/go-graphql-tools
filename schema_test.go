package generator

import (
	"log"
	"reflect"
	"sort"
	"testing"

	"github.com/graphql-go/graphql"
)

type Query struct {
	Antt int
	Node *Node
	B    NodeI
}
type NodeI struct {
	Node
	Id   string `json:"id" graphql:"id"`
	Name string
}

func CompareIntrospectionResult(res1 graphql.Result, res2 graphql.Result) bool {
	if res1.HasErrors() {
		if !res2.HasErrors() {
			return false
		}
	}

	var types1 = SortBy(res1.Data.(map[string]interface{})["__schema"].(map[string]interface{})["types"].([]interface{}), "name")
	var types2 = SortBy(res2.Data.(map[string]interface{})["__schema"].(map[string]interface{})["types"].([]interface{}), "name")
	if !reflect.DeepEqual(types1, types2) {
		log.Println(types1)
		log.Println(types2)
		return false
	}
	return true
}

func SortBy(input []interface{}, key string) []map[string]interface{} {
	output := []map[string]interface{}{}
	keys := []string{}
	objs := map[string]int{}
	for index, r := range input {
		row := r.(map[string]interface{})
		k := row[key].(string)
		keys = append(keys, k)
		objs[k] = index
	}
	sort.Strings(keys)
	for _, key := range keys {
		output = append(output, input[objs[key]].(map[string]interface{}))
	}
	return output
}
func TestSchema(t *testing.T) {

	generator := NewGenerator(nil)
	obj := generator.Generate(Query{})

	Schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: obj.(*graphql.Object),
	})

	if err != nil {
		t.Fatal(err)
	}

	nodeIType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "NodeI",
		Description: "NodeI",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name: "name",
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Query",
		Description: "Query",
		Fields: graphql.Fields{
			"antt": &graphql.Field{
				Name: "antt",
				Type: graphql.NewNonNull(graphql.Int),
			},
			"node": &graphql.Field{
				Name:        "node",
				Description: "Node",
				Type: graphql.NewInterface(graphql.InterfaceConfig{
					Name:        "Node",
					Description: "Node",
					Fields: graphql.Fields{
						"id": &graphql.Field{
							Name: "id",
							Type: graphql.NewNonNull(graphql.ID),
						},
					},
					//ResolveType: generator.ResolveType,
				}),
			},

			"b": &graphql.Field{
				Name: "b",
				Type: nodeIType,
			},
		},
	})

	expectedSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		panic(err)
	}

	var IntrospectionQuery = `
  query IntrospectionQuery {
    __schema {
      queryType { name }
	  types{
		kind
    	name
	  }
	}
	}
	`
	res := graphql.Do(graphql.Params{
		Schema:        Schema,
		RequestString: IntrospectionQuery,
	})
	expectedRes := graphql.Do(graphql.Params{
		Schema:        expectedSchema,
		RequestString: IntrospectionQuery,
	})

	if !CompareIntrospectionResult(*res, *expectedRes) {
		t.Fatalf("Invalid result, expected \n%#v \nhas \n%#v", res, expectedRes)
	}
}
