// Path: pkg/core/context.go

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// AppContext is the default implementation of the Context interface.
// It wraps http.Request and http.ResponseWriter with convenient methods
// for handling HTTP requests and responses.
//
// AppContext is designed to be reused across requests using a sync.Pool
// to minimize allocations. Use ReSetHeader() to prepare it for a new request.
type AppContext struct {
	// request is the underlying HTTP request
	request *http.Request

	// response is the underlying HTTP response writer
	response http.ResponseWriter

	// params stores URL path parameters (e.g., :id in /users/:id)
	params map[string]string

	// values stores arbitrary key-value pairs for passing data between middleware
	values map[string]interface{}

	// handlers is the chain of middleware and the final handler
	handlers []HandlerFunc

	// index tracks the current position in the handler chain
	index int

	// statusCode stores the HTTP status code to be written
	statusCode int

	//headerWritten indicates if the header has been written
	headerWritten bool

	// written indicates if the response has been written
	written bool

	// mu protects concurrent access to the context
	mu sync.RWMutex
}

// NewContext creates a new AppContext instance.
// The router typically calls this for each incoming request.
func NewContext(w http.ResponseWriter, r *http.Request) *AppContext {
	return &AppContext{
		request:    r,
		response:   w,
		params:     make(map[string]string),
		values:     make(map[string]interface{}),
		handlers:   make([]HandlerFunc, 0),
		index:      -1,
		statusCode: http.StatusOK,
		written:    false,
	}
}

// Request returns the underlying *http.Request.
func (c *AppContext) Request() *http.Request {
	return c.request
}

// Response returns the underlying http.ResponseWriter.
func (c *AppContext) Response() http.ResponseWriter {
	return c.response
}

// Param returns the value of a URL path parameter by name.
// Returns an empty string if the parameter doesn't exist.
//
// Example:
//
//	For route /users/:id with request /users/123
//	c.Param("id") returns "123"
func (c *AppContext) Param(name string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.params[name]
}

// SetParam sets a URL path parameter.
// The router typically calls this during route matching.
func (c *AppContext) SetParam(name, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.params[name] = value
}

// Query returns the value of a URL query parameter by name.
// Returns an empty string if the parameter doesn't exist.
//
// Example:
//
//	For request /users?name=john&age=30
//	c.Query("name") returns "john"
func (c *AppContext) Query(name string) string {
	return c.request.URL.Query().Get(name)
}

// QueryDefault returns the value of a URL query parameter or a default value.
func (c *AppContext) QueryDefault(name, defaultValue string) string {
	value := c.Query(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// QueryArray returns all values of a URL query parameter.
// Useful for parameters that can appear multiple times.
//
// Example:
//
//	For request /search?tag=go&tag=web&tag=framework
//	c.QueryArray("tag") returns ["go", "web", "framework"]
func (c *AppContext) QueryArray(name string) []string {
	return c.request.URL.Query()[name]
}

// Body decodes the request body into the provided interface.
// Currently, supports JSON decoding.
//
// Example:
//
//	var user
//	if err := c.Body(&user); err != nil {
//	    return err
//	}
func (c *AppContext) Body(v interface{}) error {
	if c.request.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		decoder := json.NewDecoder(c.request.Body)
		return decoder.Decode(v)
	}

	return fmt.Errorf("unsupported content type: %s", contentType)
}

// JSON writes a JSON response with the specified status code.
// Automatically sets the Content-Type header to application/json.
//
// Example:
//
//	return c.JSON(200, map[string]string{
//	    "message": "Success",
//	})
func (c *AppContext) JSON(statusCode int, data interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.Status(statusCode)
	c.writeHeaderOnce()

	encoder := json.NewEncoder(c.response)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	c.written = true
	return nil
}

// String writes a plain text response with the specified status code.
// Automatically sets the Content-Type header to text/plain.
// Supports format string with variadic values like fmt.Sprintf.
//
// Example:
//
//	return c.String(200, "Hello, %s!", "World")
//	return c.String(200, "User ID: %d", 123)
func (c *AppContext) String(statusCode int, format string, values ...interface{}) error {
	c.SetHeader("Content-Type", "text/plain; charset=utf-8")
	c.Status(statusCode)
	c.writeHeaderOnce()

	text := fmt.Sprintf(format, values...)
	if _, err := c.response.Write([]byte(text)); err != nil {
		return fmt.Errorf("failed to write string response: %w", err)
	}

	c.written = true
	return nil
}

// HTML writes an HTML response with the specified status code.
// Automatically sets the Content-Type header to text/html.
//
// Example:
//
//	return c.HTML(200, "<h1>Welcome</h1>")
func (c *AppContext) HTML(statusCode int, html string) error {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	c.Status(statusCode)
	c.writeHeaderOnce()

	if _, err := c.response.Write([]byte(html)); err != nil {
		return fmt.Errorf("failed to write HTML response: %w", err)
	}

	c.written = true
	return nil
}

// Data writes raw binary data with the specified content type.
//
// Example:
//
//	data := []byte{...}
//	return c.Data(200, "application/octet-stream", data)
func (c *AppContext) Data(statusCode int, contentType string, data []byte) error {
	c.SetHeader("Content-Type", contentType)
	c.Status(statusCode)
	c.writeHeaderOnce()

	if _, err := c.response.Write(data); err != nil {
		return fmt.Errorf("failed to write data response: %w", err)
	}

	c.written = true
	return nil
}

// NoContent sends a response with no body content.
// Commonly used for DELETE operations or 204 responses.
//
// Example:
//
//	return c.NoContent(204)
func (c *AppContext) NoContent(statusCode int) error {
	c.Status(statusCode)
	c.writeHeaderOnce()
	c.written = true
	return nil
}

// Redirect sends an HTTP redirect response.
//
// Example:
//
//	return c.Redirect(302, "/login")
func (c *AppContext) Redirect(statusCode int, location string) error {
	if statusCode < 300 || statusCode >= 400 {
		return fmt.Errorf("invalid redirect status code: %d", statusCode)
	}

	c.SetHeader("Location", location)
	c.Status(statusCode)
	c.writeHeaderOnce()
	c.written = true
	return nil
}

// Status sets the HTTP status code for the response.
// Must be called before writing the response body.
func (c *AppContext) Status(statusCode int) Context {
	c.statusCode = statusCode
	return c
}

// GetStatusCode returns the current HTTP status code.
func (c *AppContext) GetStatusCode() int {
	return c.statusCode
}

// SetHeader sets a response header.
// This is an alias for SetHeader() for compatibility.
// Must be called before writing the response body.
//
// Example:
//
//	c.SetHeader("X-Custom-Header", "value")
func (c *AppContext) SetHeader(key, value string) Context {
	c.response.Header().Set(key, value)
	return c
}

// GetHeader returns the value of a request header.
// This is an alias for Get() for compatibility.
//
// Example:
//
//	authHeader := c.GetHeader("Authorization")
func (c *AppContext) GetHeader(key string) string {
	return c.request.Header.Get(key)
}

// GetValue retrieves a value stored in the context by key.
// Returns nil if the key doesn't exist.
//
// This is useful for passing data between middleware.
//
// Example:
//
//	user := c.GetValue("user").(*User)
func (c *AppContext) GetValue(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

// SetValue stores a value in the context by key.
// This is useful for passing data between middleware.
//
// Example:
//
//	c.SetValue("user", user)
func (c *AppContext) SetValue(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// Method returns the HTTP method of the request (GET, POST, etc.).
func (c *AppContext) Method() string {
	return c.request.Method
}

// Path returns the request URL path.
//
// Example:
//
//	For request /users/123?name=john
//	c.Path() returns "/users/123"
func (c *AppContext) Path() string {
	return c.request.URL.Path
}

// URL returns the full request URL.
func (c *AppContext) URL() *url.URL {
	return c.request.URL
}

// Host returns the host from the request.
//
// Example:
//
//	c.Host() might return "example.com:8080"
func (c *AppContext) Host() string {
	return c.request.Host
}

// ClientIP attempts to get the real client IP address.
// It checks X-Forwarded-For, X-Real-IP headers and falls back to RemoteAddr.
func (c *AppContext) ClientIP() string {
	// Check X-Forwarded-For header
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// Take the first IP in the list
		if idx := strings.Index(ip, ","); idx != -1 {
			return strings.TrimSpace(ip[:idx])
		}
		return strings.TrimSpace(ip)
	}

	// Check X-Real-IP header
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(c.request.RemoteAddr, ":"); idx != -1 {
		return c.request.RemoteAddr[:idx]
	}

	return c.request.RemoteAddr
}

// UserAgent returns the User-Agent header value.
func (c *AppContext) UserAgent() string {
	return c.GetHeader("User-Agent")
}

// FormValue returns the value of a form field.
// It checks both URL query parameters and POST form data.
func (c *AppContext) FormValue(key string) string {
	return c.request.FormValue(key)
}

// FormFile retrieves a file from multipart form data.
//
// Example:
//
//	file, header, err := c.FormFile("upload")
//	if err != nil {
//	    return err
//	}
//	defer file.Close()
func (c *AppContext) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.request.FormFile(name)
}

// MultipartForm returns the parsed multipart form, including file uploads.
func (c *AppContext) MultipartForm() (*multipart.Form, error) {
	if err := c.request.ParseMultipartForm(32 << 20); err != nil { // 32 MB
		return nil, err
	}
	return c.request.MultipartForm, nil
}

// Cookie returns the value of a cookie by name.
func (c *AppContext) Cookie(name string) (string, error) {
	cookie, err := c.request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// SetCookie sets a cookie in the response.
//
// Example:
//
//	c.SetCookie(&http.Cookie{
//	    Name:     "session",
//	    Value:    "abc123",
//	    MaxAge:   3600,
//	    HttpOnly: true,
//	})
func (c *AppContext) SetCookie(cookie *http.Cookie) Context {
	http.SetCookie(c.response, cookie)
	return c
}

// IsWebSocket checks if the request is a WebSocket upgrade request.
func (c *AppContext) IsWebSocket() bool {
	return c.GetHeader("Upgrade") == "websocket"
}

// IsAjax checks if the request is an AJAX request.
// It checks for the X-Requested-With header.
func (c *AppContext) IsAjax() bool {
	return c.GetHeader("X-Requested-With") == "XMLHttpRequest"
}

// Accepts checks if the client accepts the specified content type.
//
// Example:
//
//	if c.Accepts("application/json") {
//	    return c.JSON(200, data)
//	}
func (c *AppContext) Accepts(contentType string) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, contentType) || strings.Contains(accept, "*/*")
}

// Next executes the next handler in the middleware chain.
// This is used within middleware to pass control to the next handler.
//
// Example middleware:
//
//	func LoggerMiddleware(ctx Context) error {
//	    log.Println("Before handler")
//	    if err := ctx.Next(); err != nil {
//	        return err
//	    }
//	    log.Println("After handler")
//	    return nil
//	}
func (c *AppContext) Next() error {
	c.index++
	if c.index < len(c.handlers) {
		return c.handlers[c.index](c)
	}
	return nil
}

// SetHandlers sets the handler chain for this context.
// The router typically calls this.
func (c *AppContext) SetHandlers(handlers []HandlerFunc) {
	c.handlers = handlers
	c.index = -1
}

// Reset resets the context for reuse with a new request/response pair.
// This is called by the context pool to prepare the context for the next request.
func (c *AppContext) Reset(w http.ResponseWriter, r *http.Request) {
	c.request = r
	c.response = w
	c.statusCode = http.StatusOK
	c.headerWritten = false
	c.written = false
	c.index = -1

	// Clear maps
	for k := range c.params {
		delete(c.params, k)
	}
	for k := range c.values {
		delete(c.values, k)
	}

	// Reset handlers slice but keep capacity
	c.handlers = c.handlers[:0]
}

// IsWritten returns true if the response has been written.
func (c *AppContext) IsWritten() bool {
	return c.written
}

// Context returns the request's context.Context for cancellation and deadlines.
func (c *AppContext) Context() context.Context {
	return c.request.Context()
}

// WithContext returns a shallow copy of the AppContext with a new context.Context.
func (c *AppContext) WithContext(ctx context.Context) Context {
	if ctx == nil {
		return c
	}

	newRequest := c.request.WithContext(ctx)

	c.mu.RLock()
	paramsCopy := make(map[string]string, len(c.params))
	for k, v := range c.params {
		paramsCopy[k] = v
	}
	valuesCopy := make(map[string]interface{}, len(c.values))
	for k, v := range c.values {
		valuesCopy[k] = v
	}
	c.mu.RUnlock()

	handlersCopy := make([]HandlerFunc, len(c.handlers))
	copy(handlersCopy, c.handlers)

	return &AppContext{
		request:       newRequest,
		response:      c.response,
		statusCode:    c.statusCode,
		headerWritten: c.headerWritten,
		written:       c.written,
		index:         c.index,
		params:        paramsCopy,
		values:        valuesCopy,
		handlers:      handlersCopy,
	}
}

// Write implements io.Writer interface.
// This allows Context to be used directly with functions expecting io.Writer.
func (c *AppContext) Write(data []byte) (int, error) {
	c.writeHeaderOnce()
	c.written = true
	return c.response.Write(data)
}

// Stream sends a streaming response.
// The step function is called repeatedly until it returns io.EOF or an error.
//
// Example:
//
//	return c.Stream(200, "text/event-stream", func(w io.Writer) error {
//	    for i := 0; i < 10; i++ {
//	        fmt.Fprintf(w, "data: %d\n\n", i)
//	        time.Sleep(time.Second)
//	    }
//	    return io.EOF
//	})
func (c *AppContext) Stream(statusCode int, contentType string, step func(io.Writer) error) error {
	c.SetHeader("Content-Type", contentType)
	c.SetHeader("Transfer-Encoding", "chunked")
	c.Status(statusCode)

	c.writeHeaderOnce()
	c.written = true

	if err := step(c.response); err != nil && err != io.EOF {
		return fmt.Errorf("streaming error: %w", err)
	}

	return nil
}

// Err returns any error stored in the request context.
func (c *AppContext) Err() error {
	return c.request.Context().Err()
}

func (c *AppContext) writeHeaderOnce() {
	if !c.headerWritten {
		c.Response().WriteHeader(c.statusCode)
		c.headerWritten = true
	}
}
