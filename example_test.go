package rubik_test

import "github.com/rubikorg/rubik"

func ExampleE() {
	r := rubik.Route{
		Path: "/test",
		Controller: func(req *rubik.Request) {
			req.Throw(403, rubik.E("invalid token or something"))
		},
	}
	rubik.UseRoute(r)
}
func ExampleRequest_Throw() {
	func(req *rubik.Request) {
		req.Throw(403, rubik.E("invalid token or something"))
	}(&rubik.Request{})
}
