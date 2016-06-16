package starwars

import (
	"testing"

	"github.com/arvitaly/go-graphql-tools"
	"github.com/graphql-go/graphql"
)

func TestSchema(t *testing.T) {
	gen := tools.NewGenerator(nil)
	query := gen.GenerateObject(Query{})
	mutation := gen.GenerateObject(Mutation{})

	_, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    query,
		Mutation: mutation,
	})
	if err != nil {
		t.Fatal(err)
	}

}
