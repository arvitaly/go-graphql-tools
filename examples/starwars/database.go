package starwars

//Based on https://github.com/facebook/relay/blob/master/examples/star-wars/data/database.js
/**
 * This defines a basic set of data for our Star Wars Schema.
 *
 * This data is hard coded for the sake of the demo, but you could imagine
 * fetching this data from a backend service rather than from hardcoded
 * JSON objects in a more complex demo.
 */
var xwing = Ship{
	Id:   "1",
	Name: "X-Wing",
}

var ywing = Ship{
	Id:   "2",
	Name: "Y-Wing",
}

var awing = Ship{
	Id:   "3",
	Name: "A-Wing",
}

// Yeah, technically it's Corellian. But it flew in the service of the rebels,
// so for the purposes of this demo it's a rebel ship.
var falcon = Ship{
	Id:   "4",
	Name: "Millenium Falcon",
}

var homeOne = Ship{
	Id:   "5",
	Name: "Home One",
}

var tieFighter = Ship{
	Id:   "6",
	Name: "TIE Fighter",
}

var tieInterceptor = Ship{
	Id:   "7",
	Name: "TIE Interceptor",
}

var executor = Ship{
	Id:   "8",
	Name: "Executor",
}

var rebels = map[string]interface{}{
	"id":    "1",
	"name":  "Alliance to Restore the Republic",
	"ships": []string{"1", "2", "3", "4", "5"},
}

var empire = map[string]interface{}{
	"id":    "2",
	"name":  "Galactic Empire",
	"ships": []string{"6", "7", "8"},
}

var data = map[string]map[string]interface{}{
	"Faction": map[string]interface{}{
		"1": rebels,
		"2": empire,
	},
	"Ship": map[string]interface{}{
		"1": xwing,
		"2": ywing,
		"3": awing,
		"4": falcon,
		"5": homeOne,
		"6": tieFighter,
		"7": tieInterceptor,
		"8": executor,
	},
}
var nextShipId = "9"
var nextShipName = "New ship"

func GetShips(factionId string) []interface{} {
	ships := []interface{}{}

	for _, s := range data["Faction"][factionId].(map[string]interface{})["ships"].([]string) {
		ships = append(ships, data["Ship"][s].(Ship))
	}
	return ships
}
func GetFaction(Id string) Faction {
	f := data["Faction"][Id].(map[string]interface{})
	return Faction{
		Id:   f["id"].(string),
		Name: f["name"].(string),
	}
}

/*
export function createShip(shipName, factionId) {
  const newShip = {
    id: '' + (nextShip++),
    name: shipName,
  };
  data.Ship[newShip.id] = newShip;
  data.Faction[factionId].ships.push(newShip.id);
  return newShip;
}

export function getShip(id) {
  return data.Ship[id];
}

export function getShips(id) {
  return data.Faction[id].ships.map(shipId => data.Ship[shipId]);
}

*/
