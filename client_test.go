package cherry

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	// "github.com/codekidX/ink/testutils"
	// "encoding/json"
)

type PancakeResponse struct {
	Hey string `json:"hey"`
}

type PancakeRequestEntity struct {
	RequestEntity
	CakeType int    `ink:"type|body,optional" json:"omitempty"`
	Topping  string `ink:"topping|body"`
	Addon    string `ink:"addon|query"`
	PFile    []File `ink:"sage|form"`
}

// func TestParamSubstitution(t *testing.T) {
// 	pancakeReq := PancakeRequestEntity{
// 		Addon: "honey",
// 		CakeType: 2,
// 	}
// 	pancakeReq.Route("/$/$", "somevalue")
// 	_, err := inkcl.Get(pancakeReq)
// 	if err == nil {
// 		t.Error("There should be an error because 2 $'s are passed and single param")
// 	}
// }

func TestSimpleGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		// equals(t, req.URL.String(), "/some/path")
		// Send response to be tested
		rw.Write([]byte("OK"))
	}))
	defer server.Close()
	var inkcl = NewClient(server.URL, time.Second*30)
	// dir, _ := os.Getwd()
	// gomodFile, _ := os.Open(dir + "/go.mod")
	// var pancakeRes PancakeResponse
	reqE := BlankRequestEntity{}
	reqE.Route("/")
	// pancakeReq := PancakeRequestEntity{
	// 	CakeType: 1,
	// 	Addon:    "honey",
	// 	Topping:  "chocolate",
	// 	PFile: []File{
	// 		File{
	// 			OSFile: gomodFile,
	// 		},
	// 		File{
	// 			Path: dir + "/LICENSE",
	// 		},
	// 	},
	// }
	// pancakeReq.Route("/$", "1")
	// pancakeReq.FormData = true
	// pancakeReq.Infer = &pancakeRes
	// pancakeReq.Infer = pancakeRes
	resp, err := inkcl.Get(reqE)
	// if err == nil {
	// 	fmt.Println(pancakeRes.Hey)
	// }

	if err != nil || resp.Status != 200 || resp.StringBody != "OK" {
		t.Error(err)
	}
}
