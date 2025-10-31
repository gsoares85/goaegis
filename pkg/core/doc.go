// Package core provides the fundamental building blocks of the GoAegis framework.
//
// # Overview
//
// The core package defines the essential interfaces and types that form the foundation
// of the GoAegis framework. These interfaces establish contracts that all framework
// components must follow, ensuring consistency and enabling dependency injection.
//
// # Key Interfaces
//
// Application: The main entry point that manages the framework lifecycle, modules,
// routing, and the HTTP server.
//
// Router: Handles HTTP routing and dispatches requests to appropriate handlers.
// Supports path parameters, query parameters, and all standard HTTP methods.
//
// Context: Provides convenient methods to access request data and write responses.
// It wraps http.Request and http.ResponseWriter with a chainable API.
//
// Controller: Groups related route handlers into cohesive units, providing
// organization and structure to application logic.
//
// Provider: Represents injectable dependencies that can be managed by the
// dependency injection container with configurable lifecycles.
//
// Container: The dependency injection container that manages provider registration,
// resolution, and lifecycle.
//
// Module: The building block of a GoAegis application, grouping controllers,
// providers, and other modules into cohesive functional units.
//
// # Request Pipeline Components
//
// Middleware: Functions that process requests before they reach handlers.
// Can modify requests, responses, or terminate the request chain.
//
// Guard: Determines if a request should be allowed to proceed, commonly used
// for authentication and authorization.
//
// Pipe: Transforms and validates input data before it reaches the handler.
//
// Filter: Handles exceptions thrown during request processing and generates
// appropriate error responses.
//
// Interceptor: Intercepts and transforms the result of handlers, implementing
// aspect-oriented programming patterns.
//
// # Example Usage
//
//	package main
//
//	import (
//	    "github.com/gsoares85/goaegis/pkg/core"
//	)
//
//	// Define a simple handler
//	func helloHandler(ctx core.Context) error {
//	    return ctx.JSON(200, map[string]string{
//	        "message": "Hello, World!",
//	    })
//	}
//
//	// Implement a controller
//	type UserController struct {
//	    prefix string
//	}
//
//	func (c *UserController) GetPrefix() string {
//	    return c.prefix
//	}
//
//	func (c *UserController) GetMiddleware() []core.Middleware {
//	    return nil
//	}
//
//	func (c *UserController) RegisterRoutes(router core.Router) error {
//	    router.GET("/", c.getUsers)
//	    router.GET("/:id", c.getUserById)
//	    router.POST("/", c.createUser)
//	    return nil
//	}
//
//	func (c *UserController) getUsers(ctx core.Context) error {
//	    users := []string{"Alice", "Bob", "Charlie"}
//	    return ctx.JSON(200, users)
//	}
//
//	func (c *UserController) getUserById(ctx core.Context) error {
//	    id := ctx.Param("id")
//	    return ctx.JSON(200, map[string]string{"id": id})
//	}
//
//	func (c *UserController) createUser(ctx core.Context) error {
//	    var user map[string]interface{}
//	    if err := ctx.Body(&user); err != nil {
//	        return err
//	    }
//	    return ctx.JSON(201, user)
//	}
//
// # Dependency Injection
//
// The framework uses dependency injection to manage component lifecycles and
// dependencies. Providers can be registered with different scopes:
//
// - Singleton: One instance shared across the entire application
// - Transient: New instance created every time it's resolved
// - Request: One instance per HTTP request, shared within that request
//
// # Lifecycle Hooks
//
// Modules can implement lifecycle hooks to perform initialization and cleanup:
//
// - OnModuleInit(): Called when the module is initialized
// - OnModuleDestroy(): Called when the module is destroyed
//
// # Type Safety
//
// The framework leverages Go's type system to provide compile-time safety
// while maintaining flexibility through interfaces. All core components
// are designed to be mockable for testing.
//
// # Error Handling
//
// The framework provides structured error handling through:
//
// - ErrorResponse: Standard error response format
// - ValidationErrors: Field-level validation error details
// - Exception Filters: Customizable error handling and formatting
//
// # Performance
//
// The framework is designed for performance with:
//
// - Minimal allocations in hot paths
// - Efficient routing using radix trees (implementation in router package)
// - Request pooling to reduce GC pressure
// - Concurrent request handling with goroutines
//
// For more information, see the documentation for individual types and interfaces.
package core
