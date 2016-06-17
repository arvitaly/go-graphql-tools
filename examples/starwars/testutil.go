package starwars

import (
	"testing"

	"github.com/arvitaly/go-graphql-tools"
	"github.com/arvitaly/graphql"
	"github.com/graphql-go/graphql/testutil"
)

func DoQueryWithCheck(q string, variables map[string]interface{}, expectedRes interface{}, t *testing.T) {
	router := NewRouter()
	gen := tools.NewGenerator(router)
	query := gen.GenerateObject(Query{})
	mutation := gen.GenerateObject(Mutation{})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
	if err != nil {
		t.Fatal(err)
	}

	res := graphql.Do(graphql.Params{
		Schema:         schema,
		RequestString:  q,
		VariableValues: variables,
	})
	if res.HasErrors() {
		t.Fatalf("Result has errors: %v", res.Errors)
	}

	diff := testutil.Diff(expectedRes, res.Data)

	if len(diff) > 0 {
		t.Log(diff)
		t.Fatalf("Invalid result, expected\n%v\nhas\n%v", expectedRes, res.Data)
	}
}
