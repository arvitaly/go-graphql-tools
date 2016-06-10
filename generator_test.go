package generator

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
)

var str1 = "Hello"
var int1 = 12
var intPtr1 = 156
var float1 = 123.45
var bool1 = true

type A struct {
	Antt int
	B    B

	B2    B
	B2Ptr *B
	Id    RelayGlobalLID
}
type B struct {
	Str1 *string
	C    C
}
type CArgs struct {
	Token string
	x     C
}
type C struct {
	Str2   string `graphql:"-"`
	Int1   int
	Float1 float64
	Bool1  bool
	Int2   *int
	Int3   *int
	Arr1   *[]string
	Map1   map[string]interface{}
	D      DConnection
}
type D struct {
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
	return CArgs{
		Token: "111",
	}
}

func (a A) ResolveB(p graphql.ResolveParams) (B, error) {
	return B{}, nil
}

func (b B) ResolveStr1(p graphql.ResolveParams) (*string, error) {
	return &str1, nil
}
func (b B) ResolveC(p graphql.ResolveParams) (C, error) {
	return C{Int1: int1, Float1: float1, Str2: "Hi", Int3: &intPtr1, Arr1: &[]string{"test"}}, nil
}
func (c C) ResolveBool1(p graphql.ResolveParams) (bool, error) {
	return bool1, nil
}
func (c C) ResolveD(p graphql.ResolveParams) (*relay.Connection, error) {
	return relay.ConnectionFromArray([]interface{}{
		D{Field1: "c1"},
		D{Field1: "c2"},
		D{Field1: "c3"},
	}, relay.NewConnectionArguments(p.Args)), nil
}

func TestGenerateGraphqlObject(t *testing.T) {
	obj := GenerateGraphqlObject(A{})
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
		c{
			int1
			int2
			int3
			float1
			bool1
			arr1
			d{
				edges{
					node{
						field1
						}
					}
				}
		}
	} }`
	res := graphql.Do(graphql.Params{
		Schema:        Schema,
		RequestString: query,
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
		log.Println(output.B.Str1, string(*output.B.Str1))
		t.Fatal("Invalid response, expected A.B.Str1 == " + str1 + ", has " + *output.B.Str1)
	}
	if output.B.C.Bool1 != bool1 {
		t.Fatal("Invalid response, expected output.B.C.bool1 == " + strconv.FormatBool(bool1) + ", has " + strconv.FormatBool(output.B.C.Bool1))
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
}
