# Stack <br> [![Build Status](https://travis-ci.org/alexedwards/stack.svg?branch=master)](https://travis-ci.org/alexedwards/stack)  [![Coverage](http://gocover.io/_badge/github.com/alexedwards/stack?0)](http://gocover.io/github.com/alexedwards/stack)  [![GoDoc](http://godoc.org/github.com/alexedwards/stack?status.png)](http://godoc.org/github.com/alexedwards/stack)

Stack makes it easy to chain your HTTP middleware and handlers together, and to pass request-scoped context between them. It's basically a context-aware version of [Alice](https://github.com/justinas/alice).

*This branch is undergoing some major API changes. For a stable version please use [http://gopkg.in/alexedwards/stack.v0](http://gopkg.in/alexedwards/stack.v0).*

## Features

- An inbuilt request-scoped and concurrency-safe map for sharing data (or *context*) between your HTTP middleware and handlers.
- Make reusable and append-able handler chains in the Alice style.
- Lets you use middleware with the common `func(http.Handler) http.Handler`  pattern.
- Nicely compatible with the standard `net/http` package.
- Straightforward and non-magic code.

### TODO 

- Improve README clarity
- Make a `chain.Merge()` method
- Mirror master in v1 branch (and mention gopkg.in in README)
- Add benchmarks
- Add usage examples section to documentation
