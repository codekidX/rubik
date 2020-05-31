package rubik_test

import "github.com/rubikorg/rubik"

func ExampleE() {
	func(req *rubik.Request) {
		req.Throw(403, rubik.E("invalid token or something"))
	}
}
func ExampleData_Throw() {
	func(req *rubik.Request) {
		req.Throw(403, rubik.E("invalid token or something"))
	}(&rubik.Request{})
}
