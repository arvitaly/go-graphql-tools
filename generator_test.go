package tools

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"golang.org/x/net/context"
)

var str1 = "Hello"
var int1 = 12
var intPtr1 = 156
var float1 = 123.45
var bool1 = true

type A struct {
	Antt int
	Node *Node
	B    B

	B2    B
	B2Ptr *B
}
type B struct {
	Str1 *string
	Str8 string
	C    C
}
type CArgs struct {
	Token *string
	After *string
	x     C
}

type Enum1 int

var Enum1Value1 Enum1 = 1
var Enum1Value2 Enum1 = 2
var Enum1Value3 Enum1 = 3

type C struct {
	Id      int                    `json:"id,string" graphql:"ID"`
	Ignore1 string                 `graphql:"-"`
	Str2    string                 `json:"str2"`
	Int1    int                    `json:"int1"`
	Float1  float64                `json:"float1"`
	Bool1   bool                   `json:"bool1"`
	Int2    *int                   `json:"int2"`
	Int3    *int                   `json:"int3"`
	Arr1    *[]string              `json:"arr1"`
	Map1    map[string]interface{} `json:"map1"`
	D       DConnection            `json:"d"`
	Enum1   Enum1                  `json:"enum1,string" graphql:"enum"`
}

func (e *Enum1) UnmarshalJSON(b []byte) error {
	if value, ok := e.Values()[string(b)]; ok {
		*e = value
	}
	return nil
}
func (e Enum1) Values() map[string]Enum1 {
	return map[string]Enum1{
		"Enum1": Enum1Value1,
		"Enum2": Enum1Value2,
		"Enum3": Enum1Value3,
	}
}

type Node struct {
	Id string `graphql:"id"`
}

func (n Node) IsInterface() bool {
	return true
}

type D struct {
	Node
	Id     string `json:"id" graphql:"id" resolve:"globalid"`
	Field1 string `json:"field1"`
}
type DConnection struct {
	Edges []DConnectionEdge
}
type DConnectionEdge struct {
	Node D
}

func (a A) Description() string {
	return "AAA"
}
func (b B) ArgsForC() CArgs {
	var x = "" + str1
	return CArgs{
		Token: &x,
	}
}

type Context1 struct {
	Context1 string
}

func TestGenerateGraphqlObject(t *testing.T) {

	router := NewRouter()
	router.Use(UseGlobalId)
	router.Query("A.B", func() (interface{}, error) {
		return B{Str8: "Hello"}, nil
	})
	router.Query("B.C", func(b B, argsC CArgs, ctx Context1) (interface{}, error) {
		return C{Enum1: Enum1Value1, Id: 13, Int1: int1, Float1: float1, Str2: b.Str8 + *argsC.Token + ctx.Context1, Int3: &intPtr1, Arr1: &[]string{"test"}}, nil
	})
	router.Query("B.Str1", func(p ResolveParams) (interface{}, error) {
		return &str1, nil
	})
	router.Query("C.Bool1", func(p ResolveParams) (interface{}, error) {
		return bool1, nil
	})
	router.Query("C.D", func(p ResolveParams) (interface{}, error) {

		return relay.ConnectionFromArray([]interface{}{
			D{Field1: "c1", Id: "1001"},
			D{Field1: "c2", Id: "1002"},
			D{Field1: "c3", Id: "1003"},
		}, relay.NewConnectionArguments(p.Params.Args)), nil
	})

	generator := NewGenerator(router)
	obj := generator.GenerateObject(A{})
	if obj.Name() != "A" {
		t.Fatal("Invalid name for graphql object, expected A")
	}

	Schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: obj,
	})

	if err != nil {
		t.Fatal(err)
	}
	query := `query Q1{ b{
		str1
		c(token:"token1"){
			id
			enum1
			str2
			int1
			int2
			int3
			float1
			bool1
			arr1
			d{
				edges{
					node{
						id
						field1
						}
					}
				}
		}
	} }`

	ctx := context.Background()
	ctx = context.WithValue(ctx, "context1", "context1value")
	res := graphql.Do(graphql.Params{
		Schema:        Schema,
		RequestString: query,
		Context:       ctx,
	})
	if res.HasErrors() {
		t.Fatal(res.Errors)
	}
	b, err := json.Marshal(res.Data)
	if err != nil {
		t.Fatal(err)
	}

	var output A
	err = json.Unmarshal(b, &output)
	if err != nil {
		t.Fatal(err)
	}
	if *output.B.Str1 != str1 {
		t.Fatal("Invalid response, expected A.B.Str1 == " + str1 + ", has " + *output.B.Str1)
	}
	if output.B.C.Bool1 != bool1 {
		t.Fatal("Invalid response, expected output.B.C.bool1 == " + strconv.FormatBool(bool1) + ", has " + strconv.FormatBool(output.B.C.Bool1))
	}
	if output.B.C.Id != 13 {
		t.Fatal("Invalid output.B.C.Id, expected 13, has " + strconv.Itoa(output.B.C.Id))
	}
	if output.B.C.Int1 != int1 {
		t.Fatal("Invalid response, expected output.B.C.Int1 == " + strconv.Itoa(int1) + ", has " + strconv.Itoa(output.B.C.Int1))
	}
	if output.B.C.Int2 != nil {
		t.Fatal("Invalid response, expected output.B.C.int2 == nil, has: " + strconv.Itoa(*output.B.C.Int2))
	}
	if output.B.C.Int3 == nil {
		t.Fatal("Invalid response, expected output.B.C.int3 == " + strconv.Itoa(intPtr1) + ", has: nil")
	}
	if *output.B.C.Int3 != intPtr1 {
		t.Fatal("Invalid response, expected output.B.C.Int3 == " + strconv.Itoa(intPtr1) + ", has " + strconv.Itoa(*output.B.C.Int3))
	}
	if (*output.B.C.Arr1)[0] != "test" {
		t.Fatal("Invalid response, expected output.B.C.Arr1[0] == test, has " + (*output.B.C.Arr1)[0])
	}
	if len(output.B.C.D.Edges) != 3 {
		t.Fatal("Waiting for 3 edges in output.B.C.D.Edges")
	}

	if output.B.C.D.Edges[2].Node.Field1 != "c3" {
		t.Fatal("Invalid value output.B.C.D.Edges[2].Node.Field1, expected c3, has " + output.B.C.D.Edges[2].Node.Field1)
	}

	if output.B.C.D.Edges[1].Node.Id != relay.ToGlobalID("D", "1002") {
		t.Fatal("Invalid global id value output.B.C.D.Edges[0].Node.Id, expected " + relay.ToGlobalID("D", "1001") + ", has " + output.B.C.D.Edges[0].Node.Id)
	}

	//Check args
	if output.B.C.Str2 != "Hellotoken1context1value" {
		t.Fatal("Invalid provide args, expected output.B.C.Str2 to be token1, has " + output.B.C.Str2)
	}

	//Check enum
	if output.B.C.Enum1 != Enum1Value1 {
		t.Fatal("Invalid value for output.B.C.Enum1, expected " + strconv.Itoa(int(Enum1Value1)) + ", has " + strconv.Itoa(int(output.B.C.Enum1)))
	}
}
