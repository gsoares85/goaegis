package core

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// Application represents the main application instance that manages the framework lifecycle.
// It coordinates modules, dependency injection, routing, and the HTTP server.
type Application interface {
	// RegisterModule registers a module with the application.
	// Modules are the building blocks of a GoAegis application.
	RegisterModule(module Module) error

	// Use registers a global middleware that will be applied to all routes.
	Use(middleware Middleware) Application

	// Listen starts the HTTP server on the specified address.
	// The address must be in the form "host:port", e.g., ":3000" or "localhost:3000".
	Listen(addr string) error

	// ListenTLS starts the HTTPS server with the provided certificate and key files on the specified address.
	ListenTLS(addr string, certFile, keyFile string) error

	// Shutdown gracefully shuts down the application without interrupting any active connections.
	Shutdown(ctx context.Context) error

	// GetContainer returns the dependency injection container.
	GetContainer() Container

	// GetRouter returns the application's router.
	GetRouter() Router
}

// Router handles HTTP routing and dispatches requests to the appropriate handler.
// It supports path parameters, query parameters, middlewares and various HTTP methods, e.g., GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, CONNECT, TRACE,.
type Router interface {
	// GET registers a route for HTTP GET requests
	GET(path string, handler HandlerFunc) Router

	// POST registers a route for HTTP POST requests
	POST(path string, handler HandlerFunc) Router

	// PUT registers a route for HTTP PUT requests
	PUT(path string, handler HandlerFunc) Router

	// DELETE registers a route for HTTP DELETE requests
	DELETE(path string, handler HandlerFunc) Router

	// PATCH registers a route for HTTP PATCH requests
	PATCH(path string, handler HandlerFunc) Router

	// OPTIONS register a route for HTTP OPTIONS requests
	OPTIONS(path string, handler HandlerFunc) Router

	// HEAD registers a route for HTTP HEAD requests
	HEAD(path string, handler HandlerFunc) Router

	// Handle registers a route with a custom HTTP method
	Handle(method, path string, handler HandlerFunc) Router

	// Group creates a route group with a common prefix and optional middleware.
	Group(prefix string, middleware ...Middleware) Router

	// Use adds middleware to the router.
	Use(middleware ...Middleware) Router

	// ServeHTTP implements the http.Handler interface.
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// HandlerFunc defines the signature for route handlers.
// It receives a Context which provides access to request data and response methods.
type HandlerFunc func(ctx Context) error

// Context provides methods to access request data and write responses.
// It wraps http.Request and http.ResponseWriter with convenient methods.
type Context interface {
	// Request returns the underlying http.Request.
	Request() *http.Request

	// Response returns the underlying http.ResponseWriter.
	Response() http.ResponseWriter

	// Param returns the value of a URL parameter by name
	// For route like /users/:id, the value of :id will be returned by Param("id")
	Param(name string) string

	// SetParam sets a URL parameter.
	SetParam(name, value string)

	// Query returns the value of a URL query parameter by name
	// for route like /users?id=1, the value of id will be returned by Query("id")
	Query(name string) string

	// QueryDefault returns the query parameter value or a default if not present
	QueryDefault(name, defaultValue string) string

	// QueryArray returns all values of a URL query parameter.as []string
	QueryArray(name string) []string

	// Body binds the request body to a struct using JSON decoding.
	Body(v interface{}) error

	// JSON sends a JSON response with the given status code.
	JSON(statusCode int, data interface{}) error

	// String sends a string response with the given status code.
	String(statusCode int, format string, args ...interface{}) error

	// HTML sends an HTML response with the given status code.
	HTML(statusCode int, html string) error

	// Data writes raw binary data with the specified content type.
	Data(statusCode int, contentType string, data []byte) error

	// NoContent sends a response with no body content.
	NoContent(statusCode int) error

	// Redirect sends an HTTP redirect response.
	Redirect(statusCode int, location string) error

	// Status sets the HTTP status code for the response.
	Status(statusCode int) Context

	// GetStatusCode returns the current HTTP status code.
	GetStatusCode() int

	// SetHeader sets a response header.
	SetHeader(key, value string) Context

	// GetHeader returns a request header value.
	GetHeader(key string) string

	// SetValue stores a value in the context for later retrieval.
	// This is useful for passing data between middleware and handlers.
	SetValue(key string, value interface{})

	// GetValue retrieves a value from the context.
	GetValue(key string) interface{}

	// Method returns the HTTP method of the request (GET, POST, etc.).
	Method() string

	// Path returns the request URL path.
	Path() string

	// URL returns the full request URL.
	URL() *url.URL

	// Host returns the host from the request.
	Host() string

	// ClientIP attempts to get the real client IP address.
	ClientIP() string

	// UserAgent returns the User-Agent header value.
	UserAgent() string

	// FormValue returns the value of a form field.
	FormValue(key string) string

	// FormFile retrieves a file from multipart form data.
	FormFile(name string) (multipart.File, *multipart.FileHeader, error)

	// MultipartForm returns the parsed multipart form, including file uploads.
	MultipartForm() (*multipart.Form, error)

	// Cookie returns the value of a cookie by name.
	Cookie(name string) (string, error)

	// SetCookie sets a cookie in the response.
	SetCookie(cookie *http.Cookie) Context

	// IsWebSocket checks if the request is a WebSocket upgrade request.
	IsWebSocket() bool

	// IsAjax checks if the request is an AJAX request.
	IsAjax() bool

	// Accepts checks if the client accepts the specified content type.
	Accepts(contentType string) bool

	// Next advances the request pipeline to the next handler.
	Next() error

	// SetHandlers sets the handler chain for this context.
	SetHandlers(handlers []HandlerFunc)

	// Reset resets the context for reuse with a new request/response pair.
	Reset(w http.ResponseWriter, r *http.Request)

	// IsWritten returns true if the response has been written.
	IsWritten() bool

	// Context returns the request's context.Context for cancellation and deadlines.
	Context() context.Context

	// WithContext returns a shallow copy of the AppContext with a new context.Context.
	WithContext(ctx context.Context) Context

	// Write implements io.Writer interface.
	Write(data []byte) (int, error)

	// Stream sends a streaming response.
	Stream(statusCode int, contentType string, step func(io.Writer) error) error

	// Err returns any error stored in the request context.
	Err() error
}

// Controller represents a controller that groups related route handlers.
// Controllers organize application logic into cohesive units.
type Controller interface {
	// GetPrefix returns the base path prefix for all routes in this controller.
	// For example, a UserController might return "/users".
	GetPrefix() string

	// GetMiddleware returns middleware that should be applied to all controller routes.
	GetMiddleware() []Middleware

	// RegisterRoutes registers the controller's routes with the router.
	RegisterRoutes(router Router) error
}

// Provider represents a service or component that can be injected as a dependency.
// Providers are registered in the dependency injection container.
type Provider interface {
	// GetToken returns a unique identifier for this provider.
	// This is typically the provider's type or a custom string token.
	GetToken() interface{}

	// GetScope returns the provider's lifecycle scope (Singleton, Transient, or Request).
	GetScope() ProviderScope

	// GetFactory returns a function that creates instances of the provider.
	GetFactory() ProviderFactory
}

// ProviderFactory is a function that creates instances of a provider.
// It receives the container to resolve dependencies.
type ProviderFactory func(container Container) (interface{}, error)

// Container is the dependency injection container that manages providers and their instances.
type Container interface {
	// Register registers a provider in the container.
	Register(provider Provider) error

	// Resolve resolves a dependency by its token and returns the instance.
	Resolve(token interface{}) (interface{}, error)

	// Has checks if a provider with the given token is registered.
	Has(token interface{}) bool

	// Clear removes all registered providers and cached instances.
	Clear()
}

// Module represents a cohesive unit of functionality that groups related components.
// Modules are the building blocks of a GoAegis application.
type Module interface {
	// GetControllers returns the controllers defined in this module.
	GetControllers() []Controller

	// GetProviders returns the providers (services) defined in this module.
	GetProviders() []Provider

	// GetImports returns other modules that this module depends on.
	GetImports() []Module

	// GetExports returns providers that should be available to other modules.
	// Only exported providers can be injected in other modules that import this one.
	GetExports() interface{}

	// GetMiddleware returns middleware that applies to all module routes.
	GetMiddleware() []Middleware

	// OnModuleInit is called when the module is initialized.
	// Use this to perform setup tasks like connecting to databases.
	OnModuleInit() error

	// OnModuleDestroy is called when the module is destroyed.
	// Use this to perform cleanup tasks like closing database connections.
	OnModuleDestroy() error
}

// Middleware is a function that can process requests before they reach handlers.
// Middleware can modify the request, response, or terminate the request chain.
type Middleware func(ctx Context, next HandlerFunc) error

// Guard is a middleware that determines if a request should be allowed to proceed.
// Guards are commonly used for authentication and authorization.
type Guard interface {
	// CanActivate returns true if the request should be allowed, false otherwise.
	CanActivate(ctx Context) (bool, error)
}

// Pipe transforms and validates input data before it reaches the handler.
// Pipes can parse parameters, validate data, or transform data types.
type Pipe interface {
	// Transform processes and transforms the input value.
	Transform(value interface{}, metadata PipeMetadata) (interface{}, error)
}

// PipeMetadata provides context about where a pipe is being applied.
type PipeMetadata struct {
	// Type indicates the type of parameter (body, query, param, etc.)
	Type string

	// Data contains additional metadata about the parameter
	Data interface{}
}

// Filter handles exceptions thrown during request processing.
// Filters can catch errors and return appropriate error responses.
type Filter interface {
	// Catch processes an error and generates an appropriate response.
	Catch(err error, ctx Context) error
}

// Interceptor can intercept and transform the result of handlers.
// Interceptors implement aspect-oriented programming patterns.
type Interceptor interface {
	// Intercept wraps the execution of a handler and can modify its result.
	Intercept(ctx Context, next HandlerFunc) (interface{}, error)
}

// ProviderScope defines the lifecycle of a provider instance.
type ProviderScope int

const (
	// SingletonScope means one instance is created and shared across the entire application.
	SingletonScope ProviderScope = iota
	// TransientScope means a new instance is created every time the provider is resolved.
	TransientScope
	// RequestScope means one instance is created per HTTP request and shared within that request.
	RequestScope
)

// String returns the string representation of the ProviderScope.
func (s ProviderScope) String() string {
	switch s {
	case SingletonScope:
		return "Singleton"
	case TransientScope:
		return "Transient"
	case RequestScope:
		return "Request"
	default:
		return "Unknown"
	}
}
