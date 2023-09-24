<img src="https://avatars3.githubusercontent.com/u/61872650?s=60&v=4">

# Rubik

## A very simple and composible framework for Go

[Homepage](https://rubikorg.github.io) -
[API Documentation](https://pkg.go.dev/github.com/rubikorg/rubik?tab=doc)

```go
package main

import "github.com/rubikorg/rubik"

func helloResponder(rc *rubik.Context) {
    rc.Text("hello!")
}

func main() {
    // GET
    rubik.GET("/", tellHello)
    // POST
    rubik.POST("/poster", tellHello)
    // Start the server
    rubik.Run()
}
```

### License

Rubik is released under the
[Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0)
