<p align="center">
    <img src="https://i.imgur.com/WzoCr7D.png" width="500" height="281">
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
inkcl := ink.New("http://localhost:8080", time.Second * 30)
```

Making a `GET` request is a one-liner with **Ink**

```go
resp, err := inkcl.Get("/pancakes").Call()

if err == nil {
    // Dp something with resp ...
}
```

#### Path Substitution

I call it magic because it really is, I had an hard time appending strings just to construct paths that has dynamic params in it, this feature makes it easy and is also a one-liner.

```go
var cakeType = "1"
var addon = "honey"
resp, err := inkcl.Get("/pancake/$/addon/$", cakeType, addon).Call()

if err == nil {
    // Dp something with resp ...
}
```

Notice the dollar signs and an extra params, which just looks as it is. `$`s are substituted with each params in the order in which they are supplied to function.