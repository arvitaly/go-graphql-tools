package starwars

import (
	"testing"
)

func TestGetFactions(t *testing.T) {

	q := `query Q1{ 
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
		}`

	expectedRes := map[string]interface{}{
		"rebels": map[string]interface{}{
			"id":   rebels["id"],
			"name": rebels["name"],
			"ships": map[string]interface{}{
				"edges": []interface{}{
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":   xwing.Id,
							"name": xwing.Name,
						},
					},
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":   ywing.Id,
							"name": ywing.Name,
						},
					},
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":   awing.Id,
							"name": awing.Name,
						},
					},
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":   falcon.Id,
							"name": falcon.Name,
						},
					},
					map[string]interface{}{
						"node": map[string]interface{}{
							"id":   homeOne.Id,
							"name": homeOne.Name,
						},
					},
				},
			},
		},
		"empire": map[string]interface{}{
			"id":   empire["id"],
			"name": empire["name"],
		},
	}

	DoQueryWithCheck(q, nil, expectedRes, t)
}
