package starwars

import (
	"github.com/arvitaly/go-graphql-tools"
	"github.com/graphql-go/relay"
)

func NewRouter() *tools.Router {
	router := tools.NewRouter()
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
	return router
}
