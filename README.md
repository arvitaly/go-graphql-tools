# Generator GraphQL schema from Go-structs

#IN WORKING!

Based on https://github.com/graphql-go/graphql library

[![Build Status](https://travis-ci.org/arvitaly/gopherjs-electron.svg?branch=master)](https://travis-ci.org/arvitaly/go-graphql-schema-generator)

# Example

Schema declaration

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

Query

	query Q1{ b{
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
	} }