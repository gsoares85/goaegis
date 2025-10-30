# Core Package

The `core` package contains the fundamental interfaces and types that define the architecture of the GoAegis framework.

## 📦 Contents

- **interfaces.go** - Core interface definitions
- **types.go** - Supporting types and constants
- **doc.go** - Package documentation
- **types_test.go** - Unit tests

## 🎯 Purpose

This package establishes the contracts that all framework components must follow, enabling:

- **Dependency Injection** - Type-safe DI container
- **Modular Architecture** - Composable application modules
- **Request Pipeline** - Middleware, guards, pipes, filters, interceptors
- **Type Safety** - Compile-time safety with Go's type system

## 🔑 Key Interfaces

### Application
Main application instance managing lifecycle, modules, and HTTP server.

```go
type Application interface {
    RegisterModule(module Module) error
    Use(middleware Middleware) Application
    Listen(addr string) error
    Shutdown(ctx context.Context) error
}
```

### Router
HTTP routing with support for all methods and middleware.

```go
type Router interface {
    GET(path string, handler HandlerFunc) Router
    POST(path string, handler HandlerFunc) Router
    Group(prefix string, middleware ...Middleware) Router
}
```

### Context
Request/response context with convenient methods.

```go
type Context interface {
    Param(name string) string
    Query(name string) string
    Body(v interface{}) error
    JSON(statusCode int, data interface{}) error
}
```

### Controller
Groups related routes into cohesive units.

```go
type Controller interface {
    GetPrefix() string
    RegisterRoutes(router Router) error
}
```

### Provider
Injectable dependency with configurable lifecycle.

```go
type Provider interface {
    GetToken() interface{}
    GetScope() ProviderScope
    GetFactory() ProviderFactory
}
```

### Module
Building block grouping controllers, providers, and imports.

```go
type Module interface {
    GetControllers() []Controller
    GetProviders() []Provider
    GetImports() []Module
    OnModuleInit() error
}
```

## 🔄 Request Pipeline

The framework supports a complete request processing pipeline:

1. **Middleware** - Pre/post processing
2. **Guards** - Authentication/authorization
3. **Pipes** - Data validation/transformation
4. **Handler** - Route handler
5. **Interceptors** - Response transformation
6. **Filters** - Exception handling

## 🏗️ Provider Scopes

Three lifecycle scopes for dependency injection:

- **Singleton** - One instance for entire app
- **Transient** - New instance each resolve
- **Request** - One instance per HTTP request

## 📊 Standard Response Types

```go
// Success
type SuccessResponse struct {
    StatusCode int
    Data       interface{}
    Message    string
}

// Error
type ErrorResponse struct {
    StatusCode int
    Message    string
    Error      string
    Path       string
    Timestamp  string
}

// Paginated
type PaginatedResponse struct {
    Data interface{}
    Meta PaginationMetadata
}
```

## 🧪 Testing

Run tests:

```bash
# Using PowerShell script
.\make.ps1 test

# Or directly
go test ./pkg/core/...
```

## 📝 Example Usage

```go
package main

import "github.com/yourusername/gonest/pkg/core"

// Simple handler
func helloHandler(ctx core.Context) error {
    return ctx.JSON(200, map[string]string{
        "message": "Hello, World!",
    })
}

// Controller example
type UserController struct {
    prefix string
}

func (c *UserController) GetPrefix() string {
    return "/users"
}

func (c *UserController) GetMiddleware() []core.Middleware {
    return nil
}

func (c *UserController) RegisterRoutes(router core.Router) error {
    router.GET("/", c.listUsers)
    router.GET("/:id", c.getUser)
    router.POST("/", c.createUser)
    return nil
}
```

## 🔗 Related Packages

- **di/** - Dependency injection implementation
- **router/** - HTTP routing implementation
- **module/** - Module system implementation
- **controller/** - Controller utilities

## 📚 Documentation

See [doc.go](./doc.go) for comprehensive package documentation.

Run `go doc` to view documentation:

```bash
go doc github.com/yourusername/gonest/pkg/core
```

## ✅ Task Status

- [x] Task 2: Core interfaces defined
- [ ] Task 3: Context implementation
- [ ] Task 4: HTTP methods
- [ ] Task 5: RouteDefinition

## 🎯 Next Steps

- **Task 3**: Implement Context struct
- **Task 4**: Create HTTP method system
- **Task 5**: Implement RouteDefinition

---

Part of the [GoNest Framework](../../../README.md)
