package routes

//import "github.com/senior-buddy/buddy"
import "strings"

const (
	routeDelimiter = "_"
	streamIdentifier = "StreamController"
	controllerIdentifier = "Controller"
)

// Hopefully this will be replaced in the future with autogenerated routes
var routeTableRaw = map[string]string{
	"draw": "DrawingStreamController_Draw",
	"place_sphere": "SphereController_Place",
}

var routingTable = map[string]Route{}

func init() {
	for k, v := range routeTableRaw {
		routingTable[k] = parse(v)
	}
}

type Route struct {
	controller string
	function string
	persist bool
}

func parse(s string) Route {
	var c, f string
	var p bool
	// ["DrawingStreamController", "Draw"]
	// ["PlaceObjectController", "Place"]
	pair := strings.Split(s, routeDelimiter)
	f = pair[1]

	delimiterReplacer := strings.NewReplacer(streamIdentifier, "", controllerIdentifier, "")
	c = delimiterReplacer.Replace(pair[0])

	if strings.Contains(pair[0], streamIdentifier) {
		p = true
	} else {
		p = false
	}

	return Route{
		controller: c,
		function: f,
		persist: p,
	}
}

type pair struct {
	controller, method string
}

func Lookup(route string) Route {
	return routingTable[route]
}

