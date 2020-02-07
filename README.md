<p align="center">
    <img src="https://i.imgur.com/NL7tXZj.png" width="500" height="281">
</p>

Ink aims to ease and speedup client-side development with Go. The server side development was too quick, simple and easy and when I tried to create a cli for my project [cloak](https://github.com/codekidX/cloak) it was too time consuming and repetitve. I wanted something that would speed up my client-side development with Go and here I would like to introduce to you **Ink** a HTTP client that let's me write Go API calls in breeze.

### Features

- Easy to use APIs (wrapper around net/http with zero third-party modules)
- (Magic) path substitution
- Inline query and body builder
- Response Type-Inference
- Client wise abstraction
- Async API calls with non-blocking I/O
- Easy request cancellations


### Importing

- The old way:

```sh
go get github.com/codekidX/ink
```

- Inside code:

```
import "github.com/codekidX/ink"
```

### Getting Started

The first thing we need do is create a base **Ink** client that will be common across calls.

```go
inkcl := ink.NewClient("http://localhost:8080", time.Second * 30)
```
In order to make any type of API call, you need to create a `RequestEntity`. In **ink** RequestEntity is nothing but a reusable struct for making `type-safe` request.

Let's create a new RequestEntity

```go
type PancakeRequestEntity struct {
    ink.RequestEntity
    cakeType int `ink:"cake_type|query"`
}
```

This defines your bluprint/requirements for a single API endpoint. Now let's do  a GET call with `RequestEntity`. The struct tags for ink is explained below.

```go
func main() {
    pancakeReq := PancakeRequestEnttity{
        cakeType: 2,
    }
    pancakeReq.Route("/getCake")
    resp, err := inkcl.Get(pancakeReq)
}
```

This defines a **declarative** and **reusable** way to write the requirements of a single API. Here you exactly know what all this API call requires to fulfill it's request almost without a documentation.
The `Route()` tells the RequestEntity that this is the path for current request. The response is of type `ink.Response` which is explained in the section **Response** below.

### Ink Tags

Ink tags are really good way to define the necessity of API call, in the above example we defined our tag as `ink:"cake_type|query"`.

This lets ink know of 2 things:

1. Key: in which struct value should be assigned
2. Medium: through which the struct value should be passed

In conclution the API call is requested to `http://localhost:8080/getCake?cake_type=2`.


#### Path Substitution

I had an hard time appending strings just to construct paths that has dynamic params in it, this feature makes it easy and is also a one-liner.

```go
    pancakeReq := PancakeRequestEnttity{
        cakeType: 2,
    }

    cakeId := 1
    pancakeReq.Route("/getCake/$", cakeId)

    resp, err := inkcl.Get(pancakeReq)
    if err != nil {
        // Do something with resp ...
    }
```

Notice the dollar signs and an extra param, which just looks as it is. `$`s are substituted with each params in the order in which they are supplied to function.
