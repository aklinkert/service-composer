# go-service-composer

A lightweight helper library for composing modular services in Go applications using dependency injection.

## Overview

`go-service-composer` provides a simple way to structure and compose services in Go applications by leveraging the [
`do`](https://github.com/samber/do) dependency injection container. It helps you organize your application into
manageable modules with clear lifecycle management.

## Installation

```shell
go get github.com/aklinkert/go-service-composer
```

## Features

- **Modular Architecture**: Organize your application into self-contained modules
- **Dependency Injection**: Built on top of the `do` DI container
- **Lifecycle Management**: Clean startup and shutdown of services
- **Minimal Boilerplate**: Focus on business logic, not wiring

## Usage

For a complete usage example, see [`example_test.go`](./example_test.go).

## License

```
MIT License

Copyright (c) 2025 - present Alex Klinkert
```
