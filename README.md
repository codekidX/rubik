<img src="https://avatars3.githubusercontent.com/u/61872650?s=120&v=4">

# Rubik

## An efficient, scalable Web Framework

[Homepage](https://rubikorg.github.io) - [API Documentation](https://pkg.go.dev/github.com/rubikorg/rubik?tab=doc)

For writing REST client and server-side applications faster. It provides a pluggable
layer of abstraction over `net/http` and enables automation of development environment through extensive tooling.

Even though the goal of Rubik is set it'll take a lot of time to achieve it, that being said you must
not use this framework for any production use **yet**. There can be a lot of edge cases missed and
bug fixes beyond the grasps which needs to be fixed before it is **production ready**.

### Framework Components

- Core _(this repository)_
- [CLI](https://github.com/rubikorg/okrubik)
- [Blocks](https://github.com/rubikorg/blocks)

### Quickstart

- Install Rubik CLI _(supports Linux and OSX 64-bit versions only)_

```bash
curl https://rubik.ashishshekar.com/install | sh
```

- Create a new project

```bash
okrubik new hello-world
```

- Change directory to your project name & run the project

```bash
cd hello-world
okrubik run
```

### Questions

Co-ordination with suggestions, opinions and questions can be done through this  
[Mailing list](https://groups.google.com/forum/#!forum/rubikorg)

### Contributing

We encourage you to read this [Contributing to Rubik Guidelines](https://github.com/rubikorg/rubik/blob/master/CONTRIBUTING.md) for ensuring smooth development flow.

### Core Goals

- [x] Make Rubik fun to work with!
- [ ] Provide a great tooling for Rubik
- [ ] Make client-server development easier in Go
- [ ] Concurrent messgage passing

### Things Rubik can't do yet

Here is the [list](https://github.com/rubikorg/rubik/blob/master/_scripts/incomplete.md) of things that rubik can't do yet and needs fixing/development in these areas.

### License

Rubik is released under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0)
