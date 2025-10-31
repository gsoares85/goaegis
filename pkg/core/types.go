package core

import "net/http"

// HTTPMethod represents HTTP request methods.
type HTTPMethod string

// HTTP method constants following the HTTP specification.
const (
	MethodGET     HTTPMethod = http.MethodGet
	MethodPOST    HTTPMethod = http.MethodPost
	MethodPUT     HTTPMethod = http.MethodPut
	MethodDELETE  HTTPMethod = http.MethodDelete
	MethodPATCH   HTTPMethod = http.MethodPatch
	MethodHEAD    HTTPMethod = http.MethodHead
	MethodOPTIONS HTTPMethod = http.MethodOptions
	MethodCONNECT HTTPMethod = http.MethodConnect
	MethodTRACE   HTTPMethod = http.MethodTrace
)

// String returns the string representation of the HTTP method.
func (m HTTPMethod) String() string {
	return string(m)
}

// RouteMetadata holds metadata about a route including its path, method, and handlers.
type RouteMetadata struct {
	// Method is the HTTP method for this route (GET, POST, etc.)
	Method HTTPMethod
	// Path is the URL path pattern for this route (e.g., "/users/:id")
	Path string
	// Handler is the main handler function for this route
	Handler HandlerFunc
	// Middleware is a list of middleware applied to this specific route
	Middleware []Middleware
	// Guards are authorization guards applied to this route
	Guards []Guard
	// Pipes are data transformation/validation pipes applied to this route
	Pipes []Pipe
	// Filters are exception filters applied to this route
	Filters []Filter
	// Interceptors are interceptors applied to this route
	Interceptors []Interceptor
}

// ControllerMetadata holds metadata about a controller including its prefix and routes.
type ControllerMetadata struct {
	// Prefix is the base path for all routes in this controller
	Prefix string
	// Routes are the route definitions for this controller
	Routes []RouteMetadata
	// Middleware applied to all routes in this controller
	Middleware []Middleware
	// Guards applied to all routes in this controller
	Guards []Guard
}

// ModuleMetadata holds configuration and metadata for a module.
type ModuleMetadata struct {
	// Controllers are the controllers defined in this module
	Controllers []Controller
	// Providers are the services/providers defined in this module
	Providers []Provider
	// Imports are other modules that this module depends on
	Imports []Module
	// Exports are providers that should be available to importing modules
	Exports []interface{}
	// Middleware applied to all routes in this module
	Middleware []Middleware
	// IsGlobal indicates if this module's providers should be available globally
	IsGlobal bool
}

// ProviderMetadata holds metadata about a provider registration.
type ProviderMetadata struct {
	// Token is the unique identifier for this provider
	Token interface{}
	// Factory is the function that creates instances
	Factory ProviderFactory
	// Scope is the lifecycle scope of the provider
	Scope ProviderScope
	// Dependencies are the tokens of providers this provider depends on
	Dependencies []interface{}
}

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
	// StatusCode is the HTTP status code
	StatusCode int `json:"statusCode"`
	// Message is the error message
	Message string `json:"message"`
	// Error is the error type/name
	Error string `json:"error,omitempty"`
	// Path is the request path where the error occurred
	Path string `json:"path"`
	// Timestamp is when the error occurred
	Timestamp string `json:"timestamp,omitempty"`
}

// SuccessResponse represents a standard success response structure.
type SuccessResponse struct {
	// StatusCode is the HTTP status code
	StatusCode int `json:"statusCode"`
	// Message is an optional success message
	Message string `json:"message"`
	// Data is the response data
	Data interface{} `json:"data"`
}

// PaginationMetadata holds metadata for paginated responses.
type PaginationMetadata struct {
	// Page is the current page number (1-indexed)
	Page int `json:"page"`
	// Limit is the number of items per page
	Limit int `json:"limit"`
	// Total is the total number of items
	Total int `json:"total"`
	// TotalPages is the total number of pages
	TotalPages int `json:"totalPages"`
	// HasNext indicates if there's a next page
	HasNext bool `json:"hasNext"`
	// HasPrev indicates if there's a previous page
	HasPrev bool `json:"hasPrev"`
}

// PaginatedResponse represents a paginated response structure.
type PaginatedResponse struct {
	// Data is the array of items for the current page
	Data interface{} `json:"data"`
	// Meta contains pagination metadata
	Meta PaginationMetadata `json:"meta"`
}

// ConfigOptions holds configuration options for the application.
type ConfigOptions struct {
	// Port is the HTTP server port
	Port int
	// Host is the HTTP server host
	Host string
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout int
	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout int
	// MaxHeaderBytes is the maximum size of request headers
	MaxHeaderBytes int
	// EnableCORS enables Cross-Origin Resource Sharing
	EnableCors bool
	// TrustProxy enables trusting proxy headers (X-Forwarded-*)
	TrustProxy bool
	// Environment is the application environment (development, production, etc.)
	Environment string
}

// DefaultConfigOptions returns default configuration options.
func DefaultConfigOptions() ConfigOptions {
	return ConfigOptions{
		Port:           3000,
		Host:           "0.0.0.0",
		ReadTimeout:    30,
		WriteTimeout:   30,
		MaxHeaderBytes: 1 << 20,
		EnableCors:     false,
		TrustProxy:     false,
		Environment:    "development",
	}
}

// ValidationError represents a validation error with field-level details.
type ValidationError struct {
	// Field is the name of the field that failed validation
	Field string `json:"field"`
	// Message is the validation error message
	Message string `json:"message"`
	// Value is the value that failed validation (optional)
	Value interface{} `json:"value"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors.
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "validation failed"
	}
	return v[0].Message
}

// MiddlewareFunc is an adapter to allow ordinary functions to be used as middleware.
// This allows for more flexible middleware definition.
type MiddlewareFunc func(ctx Context, next HandlerFunc) error

// RouteOptions holds options for route registration.
type RouteOptions struct {
	// Middleware specific to this route
	Middleware []Middleware
	// Guards specific to this route
	Guards []Guard
	// Pipes specific to this route
	Pipes []Pipe
	// Filters specific to this route
	Filters []Filter
	// Interceptors specific to this route
	Interceptors []Interceptor
}

// LifecycleHook represents a hook that can be executed at various lifecycle stages.
type LifecycleHook interface {
	// OnInit is called when the component is initialized
	OnInit() error
	// OnDestroy is called when the component is destroyed
	OnDestroy() error
}

// LogLevel represents the severity level of a log message.
type LogLevel int

const (
	// LogLevelDebug is for debug messages
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for informational messages
	LogLevelInfo
	// LogLevelWarn is for warning messages
	LogLevelWarn
	// LogLevelError is for error messages
	LogLevelError
	// LogLevelFatal is for fatal error messages
	LogLevelFatal
)

// String returns the string representation of the LogLevel.
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}
