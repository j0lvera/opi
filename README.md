# opi

An opinionated set of tools for REST APIs written in Go, focused on CRUD operations.

## Overview

`opi` provides a clean, type-safe approach to building REST APIs in Go with built-in pagination, validation, and error handling. It's designed to be router-agnostic, allowing you to use it with your preferred HTTP router.

## Features

- ✨ Generic CRUD handlers
- 📝 Built-in pagination
- ✅ Request validation
- 🚦 Structured error handling
- 🔌 Router agnostic
- 💪 Type-safe operations

## Installation

```bash
go get github.com/j0lvera/opi
```

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/j0lvera/opi/crud"
	// ... other imports
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserQuery struct {
    crud.PaginatedQuery
    Name string `form:"name"`
}

func main() {
	querier := NewUserQuerier()
	writer := NewJSONWriter()
	handler := crud.NewListHandler[User, UserQuery](querier, writer)
	
    // use with any router
    http.HandleFunc("/users", handler.Handle)
}
```

## Current Status

Work in Progress! Currently implemented:

- ✅ List (GET) operations with pagination
- ⏳ Create (POST) - Coming soon
- ⏳ Update (PATCH) - Coming soon
- ⏳ Delete (DELETE) - Coming soon

## Design Philosophy

This library is intentionally opinionated to provide a consistent approach to building REST APIs. While it's primarily designed for personal projects, it's open source to allow others to fork and adapt it to their needs.

## License

MIT License

## Contributing

Feel free to open issues and pull requests. As this is an opinionated library, please open an issue to discuss significant changes before submitting PRs.