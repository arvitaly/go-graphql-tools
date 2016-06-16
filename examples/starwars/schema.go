package starwars

/**
 * Using our shorthand to describe type systems,
 * the type system for our example will be the following:
 *
 * interface Node {
 *   id: ID!
 * }
 *
 * type Faction : Node {
 *   id: ID!
 *   name: String
 *   ships: ShipConnection
 * }
 *
 * type Ship : Node {
 *   id: ID!
 *   name: String
 * }
 *
 * type ShipConnection {
 *   edges: [ShipEdge]
 *   pageInfo: PageInfo!
 * }
 *
 * type ShipEdge {
 *   cursor: String!
 *   node: Ship
 * }
 *
 * type PageInfo {
 *   hasNextPage: Boolean!
 *   hasPreviousPage: Boolean!
 *   startCursor: String
 *   endCursor: String
 * }
 *
 * type Query {
 *   rebels: Faction
 *   empire: Faction
 *   node(id: ID!): Node
 * }
 *
 * input IntroduceShipInput {
 *   clientMutationId: string!
 *   shipName: string!
 *   factionId: ID!
 * }
 *
 * input IntroduceShipPayload {
 *   clientMutationId: string!
 *   ship: Ship
 *   faction: Faction
 * }
 *
 * type Mutation {
 *   introduceShip(input IntroduceShipInput!): IntroduceShipPayload
 * }
 */

type Node struct {
	Id string `json:"id" "graphql:id"`
}

func (n Node) IsInterface() bool {
	return true
}

type Faction struct {
	Node
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
