package tools

import (
	"testing"

	"github.com/arvitaly/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

type X struct {
	Z string
}

func TestResolveQuery(t *testing.T) {
	x := X{
		Z: "Hello world",
	}
	router := NewRouter()
	router.Query("X.Y", func(rp ResolveParams) (interface{}, error) {
		return rp.Source.(X).Z, nil
	})

	graphqlParams := graphql.ResolveParams{
		Source: x,
		Info: graphql.ResolveInfo{
			FieldName: "y",
			Operation: &ast.OperationDefinition{
				Operation: "query",
			},
		},
	}
	res, err := router.Resolve(FieldInfo{Source: X{}, Path: "X.Y"}, graphqlParams)
	if err != nil {
		t.Fatalf("Invalid result, error not nil, has %v", err)
	}
	if res != x.Z {
		t.Fatalf("Invalid result, expected %v, has %v", x.Z, res)
	}

}
