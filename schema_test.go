package tools

import (
	"log"
	"reflect"
	"sort"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
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

type Mutation struct {
	Mutation1 Mutation1Payload `json:"mutation1"`
}
type Mutation1Input struct {
	In1 string `json:"in1"`
}
type Mutation1Args struct {
	Input *Mutation1Input `json:"input" graphql:"input" description:"Mutation1Input"`
}

func (m Mutation) ArgsForMutation1() Mutation1Args {
	return Mutation1Args{}
}

type Mutation1Payload struct {
	Out1 string `json:"out1"`
}

func CompareIntrospectionResult(res1 graphql.Result, res2 graphql.Result) bool {
	if res1.HasErrors() {
		if !res2.HasErrors() {
			return false
		}
	}

	var types1 = SortBy(res1.Data.(map[string]interface{})["__schema"].(map[string]interface{})["types"].([]interface{}), "name")
	var types2 = SortBy(res2.Data.(map[string]interface{})["__schema"].(map[string]interface{})["types"].([]interface{}), "name")
	for i, type1 := range types1 {
		if types1[i]["fields"] != nil {
			types1[i]["fields"] = SortBy(type1["fields"].([]interface{}), "name")
		}

	}
	for i, type2 := range types2 {
		if types2[i]["fields"] != nil {
			types2[i]["fields"] = SortBy(type2["fields"].([]interface{}), "name")
		}

	}
	for i, _ := range types1 {
		if !reflect.DeepEqual(types1[i], types2[i]) {
			log.Println("Not equal ", i)
			log.Println("Has")
			log.Println(types1[i])
			log.Println("Expected")
			log.Println(types2[i])

			for key1, value1 := range types1[i] {
				t2 := types2[i]
				if !reflect.DeepEqual(value1, t2[key1]) {
					log.Println("Has")
					log.Println(key1, value1)
					log.Println("Expected")
					log.Println(key1, t2[key1])
					return false
				}
			}

		}
	}

	if !reflect.DeepEqual(types1, types2) {
		log.Println("Has")
		log.Println(types1)
		log.Println("Expected")
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
	obj := generator.GenerateObject(Query{})

	Schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    obj,
		Mutation: generator.GenerateObject(Mutation{}),
	})

	if err != nil {
		t.Fatal(err)
	}
	nodeType := graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "Node",
		Description: "Node",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Name:        "id",
				Description: "Id",
				Type:        graphql.NewNonNull(graphql.ID),
			},
		},
		//ResolveType: generator.ResolveType,
	})
	nodeIType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "NodeI",
		Description: "NodeI",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Name:        "name",
				Description: "Name",
				Type:        graphql.NewNonNull(graphql.String),
			},
			"id": &graphql.Field{
				Name:        "id",
				Description: "Id",
				Type:        graphql.NewNonNull(graphql.ID),
			},
		},
		Interfaces: []*graphql.Interface{nodeType},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Query",
		Description: "Query",
		Fields: graphql.Fields{
			"antt": &graphql.Field{
				Name:        "antt",
				Description: "Antt",
				Type:        graphql.NewNonNull(graphql.Int),
			},
			"node": &graphql.Field{
				Name:        "node",
				Description: "Node",
				Type:        nodeType,
			},

			"b": &graphql.Field{
				Name:        "b",
				Description: "B",
				Type:        nodeIType,
			},
		},
	})

	/*mutation1 := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Mutation1Payload",
		Description: "Mutation1Payload",
		Fields: graphql.Fields{
			"out1": &graphql.Field{
				Description: "Out1",
				Type:        graphql.NewNonNull(graphql.String),
			},
		},
	})*/

	mutation1 := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Mutation1Payload",
		Description: "Mutation1Payload",
		Fields: graphql.Fields{
			"out1": &graphql.Field{
				Description: "Out1",
				Type:        graphql.NewNonNull(graphql.String),
			},
		},
	})
	mutation1Field := &graphql.Field{
		Name:        "mutation1",
		Description: "Mutation1",
		Type:        mutation1,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Description: "Mutation1Input",
				Type: graphql.NewInputObject(graphql.InputObjectConfig{
					Name:        "Mutation1Input",
					Description: "Input",
					Fields: graphql.InputObjectConfigFieldMap{
						"in1": &graphql.InputObjectFieldConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
				}),
			},
		},
	}

	/*mutation1 := relay.MutationWithClientMutationID(relay.MutationConfig{
		Name:        "Mutation1",
		Description: "Mutation1",
		InputFields: graphql.InputObjectConfigFieldMap{
			"in1": &graphql.InputObjectFieldConfig{
				Description:  "In1",
				DefaultValue: map[string]interface{}{},
				Type:         graphql.NewNonNull(graphql.String),
			},
		},
		OutputFields: graphql.Fields{
			"out1": &graphql.Field{
				Description: "Out1",
				Type:        graphql.NewNonNull(graphql.String),
			},
		},
		MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			return nil, nil
		},
	})*/

	mutation := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Mutation",
		Description: "Mutation",
		Fields: graphql.Fields{
			"mutation1": mutation1Field,
		},
	})

	expectedSchema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutation,
	})
	if err != nil {
		t.Fatalf("Invalid expected schema, %v", err)
	}

	res := graphql.Do(graphql.Params{
		Schema:        Schema,
		RequestString: testutil.IntrospectionQuery,
	})
	expectedRes := graphql.Do(graphql.Params{
		Schema:        expectedSchema,
		RequestString: testutil.IntrospectionQuery,
	})
	if res.HasErrors() {
		t.Fatalf("Invalid result, %v", res.Errors)
	}
	if expectedRes.HasErrors() {
		t.Fatalf("Invalid expected result, %v", expectedRes.Errors)
	}

	if !CompareIntrospectionResult(*res, *expectedRes) {
		t.Fatalf("Invalid result, expected \n%#v \nhas \n%#v", res, expectedRes)
	}
}
