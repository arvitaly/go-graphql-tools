# Generator GraphQL schema from Go-structs

#IN WORKING!

Based on https://github.com/graphql-go/graphql library

[![Build Status](https://travis-ci.org/arvitaly/gopherjs-electron.svg?branch=master)](https://travis-ci.org/arvitaly/go-graphql-schema-generator)

# Example of Star Wars, based on https://github.com/facebook/relay/blob/master/examples/star-wars/data/schema.js

Query's example
	query Q1{ 
			rebels{
				id 
				name
				ships{
					edges{
						node{
							id
							name
						}
					}
				}
			} 
			empire{
				id 
				name
			}				 
		}
Mutation's example
	mutation M1{ introduceShip(input:{shipName:"New shippy"}){
		ship{
			id
			name
		}
	} }	

Schema declaration

	type Node struct {
		Id string `json:"id" "graphql:id"`
	}
	
	func (n Node) IsInterface() bool {
		return true
	}
	
	type Faction struct {
		Node  `graphql:"interface"`
		Id    string         `json:"id" "graphql:id"`
		Name  string         `json:"name"`
		Ships ShipConnection `json:"ships"`
	}
	type Ship struct {
		Id   string `json:"id" graphql:"id"`
		Name string `json:"name"`
	}
	type ShipConnection struct {
		Edges    []ShipEdge `json:"edges"`
		PageInfo *PageInfo  `json:"pageInfo"`
	}
	
	type ShipEdge struct {
		Cursor *string `json:"cursor"`
		Node   Ship    `json:"node"`
	}
	
	type PageInfo struct {
		HasNextPage     *bool  `json:"hasNextPage"`
		HasPreviousPage *bool  `json:"hasPreviousPage"`
		StartCursor     string `json:"startCursor"`
		EndCursor       string `json:"endCursor"`
	}
	type Query struct {
		Rebels Faction `json:"rebels"`
		Empire Faction `json:"empire"`
		Node   Node    `json:"node"`
	}
	type QueryNodeArgs struct {
		Id string `json:"id" graphql:"id"`
	}
	
	func (q Query) ArgsForNode() QueryNodeArgs {
		return QueryNodeArgs{}
	}
	
	type IntroduceShipInput struct {
		ClientMutationId *string `json:"clientMutationId"`
		ShipName         *string `json:"shipName"`
		FactionId        *string `json:"factionId" graphql:"id"`
	}
	
	type IntroduceShipPayload struct {
		ClientMutationId *string `json:"clientMutationId"`
		Ship             Ship    `json:"ship"`
		Faction          Faction `json:"faction"`
	}
	
	type Mutation struct {
		IntroduceShip IntroduceShipPayload `json:"introduceShip"`
	}
	type MutationIntroduceShipArgs struct {
		Input *IntroduceShipInput `json:"input" graphql:"input"`
	}
	
	func (m Mutation) ArgsForIntroduceShip() MutationIntroduceShipArgs {
		return MutationIntroduceShipArgs{}
	}

Resolve

	router.Query("Query.Rebels", func(query Query) (Faction, error) {
		return GetFaction("1"), nil
	})
	router.Query("Query.Empire", func(query Query) (Faction, error) {
		return GetFaction("2"), nil
	})
	router.Query("Faction.Ships", func(faction Faction, p tools.ResolveParams) (*relay.Connection, error) {
		return relay.ConnectionFromArray(GetShips(faction.Id), relay.NewConnectionArguments(p.Params.Args)), nil
	})
	router.Query("Mutation.IntroduceShip", func(m Mutation, args MutationIntroduceShipArgs) (interface{}, error) {
		return IntroduceShipPayload{
			Ship: Ship{
				Id:   nextShipId,
				Name: *args.Input.ShipName,
			},
		}, nil
	})
	
