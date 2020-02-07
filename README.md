<p align="center">
    <img src="https://i.imgur.com/NL7tXZj.png" width="500" height="281">
</p>

Ink aims to ease and speedup client-side REST development with Go in a declarative manner. The server side development was too quick, simple and easy and when I tried to create a cli for my project [cloak](https://github.com/codekidX/cloak) it was too time consuming and repetitve. I wanted something that would speed up my client-side development with Go and I ended up writing **Ink** - a HTTP client that let's you write Go REST API calls in a blink.


### Importing

- The old way:

```sh
go get github.com/codekidX/cherry
```

- Inside code:

```
import "github.com/codekidX/cherry"
```

### Getting Started

The first thing we need do is create a base **Cherry** client that will be common across calls.

```go
inkcl := ink.NewClient("http://localhost:8080", time.Second * 30)
```
In order to make any type of API call, you need to create a `RequestEntity`. In **cherry** RequestEntity is nothing but a reusable struct for making `type-safe` request.

Let's create a new RequestEntity


#### Example: Simple GET call

```go
func main() {
    req := cherry.BlankRequestEntity{}
    req.Route("/")
    resp, err := chcl.Get(req)

    if resp.Status == 200 && err == nil {
        // Do something with resp
    }
}
```

#### Example: GET call with additional query

```go
type PancakeRequestEntity struct {
    chcl.RequestEntity
    cakeType int `cherry:"cake_type|query"`
}
```

The bluprint/requirements for a single API endpoint is set up. Now let's do a GET call with `RequestEntity`. The struct tags for chcl is explained below in **Cherry Tags** section.

```go
func main() {
    pancakeReq := PancakeRequestEnttity{
        cakeType: 2,
    }
    pancakeReq.Route("/getCake")
    resp, err := chcl.Get(pancakeReq)
}
```

This defines a **declarative** and **reusable** way to write the requirements of a single API. Here you exactly know what all this API call requires to fulfill it's request almost without a documentation.
The `Route()` tells the RequestEntity that this is the path for current request. The response is of type `cherry.Response` which is explained in the section **Response** below.

### Cherry Tags

Ink tags are really good way to define the necessity of API call, in the above example we defined our tag as `cherry:"cake_type|query"`.

This lets cherry know of 2 things:

1. Key: in which struct value should be assigned
2. Medium: through which the struct value should be passed

In conclution the API call is requested to `http://localhost:8080/getCake?cake_type=2`.

#### Types of `cherry` Tags

- `query`: Query string key value pairs
- `body`: Request body key value pairs
- `form`: Multtipart form-data key value pairs 


### Path Substitution

I had an hard time appending strings just to construct paths that has dynamic params in it, this feature makes it easy and is also a one-liner.

```go
    pancakeReq := PancakeRequestEnttity{
        cakeType: 2,
    }

    cakeId := 1
    pancakeReq.Route("/getCake/$", cakeId)

    resp, err := chcl.Get(pancakeReq)
    if err != nil {
        // Do something with resp ...
    }
```

Notice the dollar sign and an extra param, which just looks as it is. `$` is substituted with each params in the order in which they are supplied to `Route()` method.
