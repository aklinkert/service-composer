# service-composer

[![test](https://github.com/aklinkert/service-composer/actions/workflows/test.yml/badge.svg)](https://github.com/aklinkert/service-composer/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/aklinkert/service-composer.svg)](https://pkg.go.dev/github.com/aklinkert/service-composer)

A lightweight helper library for composing modular services in Go applications using dependency injection.

## Overview

`service-composer` provides a simple way to structure and compose services in Go applications by leveraging the
[`do`](https://github.com/samber/do) dependency injection container. It helps you organize your application into
manageable modules with clear lifecycle management.

## Installation

```shell
go get github.com/aklinkert/service-composer
```

Requires Go 1.26 or newer.

## Features

- **Modular Architecture**: Organize your application into self-contained modules
- **Dependency Injection**: Built on top of the `do` DI container
- **Dependency Ordering**: Modules declare `DependsOn()`; initialization runs in a stable topological order and
  fails on dependency cycles
- **Lifecycle Management**: Clean startup and shutdown of services
- **Minimal Boilerplate**: Focus on business logic, not wiring

## Usage

Documentation and examples can be found in the [GoDoc](https://pkg.go.dev/github.com/aklinkert/service-composer).
For a complete usage example, see [`example_test.go`](./example_test.go#L68).

## License

```
MIT License

Copyright (c) 2025 - present Alex Klinkert
```
