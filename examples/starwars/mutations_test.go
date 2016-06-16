package starwars

import (
	"testing"
)

func TestIntroduceShip(t *testing.T) {
	q := `
	mutation M1{ introduceShip(input:{shipName:"New shippy"}){
		ship{
			id
			name
		}
	} }
	`
	expectedRes := map[string]interface{}{
		"introduceShip": map[string]interface{}{
			"ship": map[string]interface{}{
				"id":   nextShipId,
				"name": "New shippy",
			},
		},
	}
	//var shipName = "Ship1"
	variables := map[string]interface{}{}
	DoQueryWithCheck(q, variables, expectedRes, t)
}
