<img src="https://avatars3.githubusercontent.com/u/61872650?s=60&v=4">

# Rubik

## A very simple and composible framework for Go

[Homepage](https://rubikorg.github.io) -
[API Documentation](https://pkg.go.dev/github.com/rubikorg/rubik?tab=doc)

```go
func tellHello(rc *rubik.Context) {
    rc.Text("Hello World")
}

// GET
index := rubik.Route{
    Path: "/",
    Responders: []rubik.Responder {tellHello}
}
// POST
poster := rubik.Route{
    Path: "/poster",
    Method: []string{http.MethodPost},
    Responders: []rubik.Responder {tellHello}
}
rubik.Use(index, poster)
```

### License

Rubik is released under the
[Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0)
