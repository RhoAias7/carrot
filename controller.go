package buddy

import (
	"reflect"
	"fmt"
	"log"
)

type AppController struct {
	persist bool
	Controller ControllerType
	//reqBuffer chan *buddy.Request
	// add a responder here for responding to all, groups, individual etc"
}

type ControllerType interface{}

func (c *AppController) Persist(p bool) {
	c.persist = p
}

/*
		Reflect on the controller and find the correct function to call, then call it
*/
func (c *AppController) Invoke(route *Route, req *Request) {

	fmt.Println(route)

	req.AddMetric(ControllerInvocation)

	// Create a new Value pointer representing the controller type
	ptr := reflect.New(reflect.TypeOf(c.Controller))

	// Look at that value then call the correct method
	method := ptr.MethodByName(route.Function())

	fmt.Println(method)

	if method.IsValid() {
		args := []reflect.Value{reflect.ValueOf(req)}
		method.Call(args)
	} else {
		log.Printf("error: invalid method called")
	}
}