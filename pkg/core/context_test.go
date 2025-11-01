package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)

	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}

	if ctx.Request() != r {
		t.Error("Request() should return the original request")
	}

	if ctx.Response() != w {
		t.Error("Response() should return the original response writer")
	}

	if ctx.GetStatusCode() != http.StatusOK {
		t.Errorf("Initial status code should be 200, got %d", ctx.GetStatusCode())
	}
}

func TestContext_Param(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/123", nil)
	ctx := NewContext(w, r)

	// Set param
	ctx.SetParam("id", "123")

	// Get param
	if got := ctx.Param("id"); got != "123" {
		t.Errorf("Param('id') = %v, want '123'", got)
	}

	// Non-existent param
	if got := ctx.Param("name"); got != "" {
		t.Errorf("Non-existent param should return empty string, got %v", got)
	}
}

func TestContext_Query(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?name=john&age=30", nil)
	ctx := NewContext(w, r)

	if got := ctx.Query("name"); got != "john" {
		t.Errorf("Query('name') = %v, want 'john'", got)
	}

	if got := ctx.Query("age"); got != "30" {
		t.Errorf("Query('age') = %v, want '30'", got)
	}

	if got := ctx.Query("missing"); got != "" {
		t.Errorf("Non-existent query param should return empty string, got %v", got)
	}
}

func TestContext_QueryDefault(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?name=john", nil)
	ctx := NewContext(w, r)

	// Existing param
	if got := ctx.QueryDefault("name", "default"); got != "john" {
		t.Errorf("QueryDefault('name') = %v, want 'john'", got)
	}

	// Missing param with default
	if got := ctx.QueryDefault("age", "25"); got != "25" {
		t.Errorf("QueryDefault('age', '25') = %v, want '25'", got)
	}
}

func TestContext_QueryArray(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?tag=go&tag=web&tag=api", nil)
	ctx := NewContext(w, r)

	tags := ctx.QueryArray("tag")

	if len(tags) != 3 {
		t.Errorf("QueryArray('tag') length = %d, want 3", len(tags))
	}

	expected := []string{"go", "web", "api"}
	for i, tag := range tags {
		if tag != expected[i] {
			t.Errorf("QueryArray('tag')[%d] = %v, want %v", i, tag, expected[i])
		}
	}
}

func TestContext_Body(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Test successful JSON decoding
	t.Run("ValidJSON", func(t *testing.T) {
		jsonData := `{"name":"John","age":30}`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/users", strings.NewReader(jsonData))
		r.Header.Set("Content-Type", "application/json")
		ctx := NewContext(w, r)

		var user User
		if err := ctx.Body(&user); err != nil {
			t.Fatalf("Body() error = %v", err)
		}

		if user.Name != "John" {
			t.Errorf("user.Name = %v, want 'John'", user.Name)
		}

		if user.Age != 30 {
			t.Errorf("user.Age = %v, want 30", user.Age)
		}
	})

	// Test invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		jsonData := `{"name":"John","age":}`
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/users", strings.NewReader(jsonData))
		r.Header.Set("Content-Type", "application/json")
		ctx := NewContext(w, r)

		var user User
		if err := ctx.Body(&user); err == nil {
			t.Error("Body() should return error for invalid JSON")
		}
	})

	// Test unsupported content type
	t.Run("UnsupportedContentType", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/users", strings.NewReader("data"))
		r.Header.Set("Content-Type", "text/plain")
		ctx := NewContext(w, r)

		var user User
		if err := ctx.Body(&user); err == nil {
			t.Error("Body() should return error for unsupported content type")
		}
	})

	// Test empty body
	t.Run("EmptyBody", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/users", nil)
		r.Header.Set("Content-Type", "application/json")
		ctx := NewContext(w, r)

		var user User
		if err := ctx.Body(&user); err == nil {
			t.Error("Body() should return error for empty body")
		}
	})
}

func TestContext_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	data := map[string]string{
		"message": "Hello, World!",
	}

	err := ctx.JSON(200, data)
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	// Check status code
	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	// Check Content-Type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %s, want 'application/json'", contentType)
	}

	// Check response body
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["message"] != "Hello, World!" {
		t.Errorf("Response message = %v, want 'Hello, World!'", response["message"])
	}

	// Check IsWritten
	if !ctx.IsWritten() {
		t.Error("IsWritten() should be true after JSON()")
	}
}

func TestContext_String(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	// Test simple string
	err := ctx.String(200, "Hello, %s!", "World")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Errorf("Content-Type = %s, want 'text/plain'", contentType)
	}

	if w.Body.String() != "Hello, World!" {
		t.Errorf("Response body = %v, want 'Hello, World!'", w.Body.String())
	}

	if !ctx.IsWritten() {
		t.Error("IsWritten() should be true after String()")
	}
}

func TestContext_StringWithFormatting(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	// Test with multiple format values
	err := ctx.String(200, "User %s has ID %d", "John", 123)
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	expected := "User John has ID 123"
	if w.Body.String() != expected {
		t.Errorf("Response body = %v, want %v", w.Body.String(), expected)
	}
}

func TestContext_HTML(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	html := "<h1>Welcome</h1>"
	err := ctx.HTML(200, html)
	if err != nil {
		t.Fatalf("HTML() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Content-Type = %s, want 'text/html'", contentType)
	}

	if w.Body.String() != html {
		t.Errorf("Response body = %v, want %v", w.Body.String(), html)
	}
}

func TestContext_Data(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	data := []byte("binary data here")
	err := ctx.Data(200, "application/octet-stream", data)

	if err != nil {
		t.Fatalf("Data() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/octet-stream" {
		t.Errorf("Content-Type = %s, want 'application/octet-stream'", contentType)
	}

	if !bytes.Equal(w.Body.Bytes(), data) {
		t.Error("Response body doesn't match sent data")
	}
}

func TestContext_NoContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/users/123", nil)
	ctx := NewContext(w, r)

	err := ctx.NoContent(204)
	if err != nil {
		t.Fatalf("NoContent() error = %v", err)
	}

	if w.Code != 204 {
		t.Errorf("Status code = %d, want 204", w.Code)
	}

	if w.Body.Len() != 0 {
		t.Error("Response body should be empty")
	}

	if !ctx.IsWritten() {
		t.Error("IsWritten() should be true after NoContent()")
	}
}

func TestContext_Redirect(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		location   string
		wantErr    bool
	}{
		{"Valid302", 302, "/login", false},
		{"Valid301", 301, "https://example.com", false},
		{"Invalid200", 200, "/test", true},
		{"Invalid404", 404, "/test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			ctx := NewContext(w, r)

			err := ctx.Redirect(tt.statusCode, tt.location)

			if (err != nil) != tt.wantErr {
				t.Errorf("Redirect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if w.Header().Get("Location") != tt.location {
					t.Errorf("Location header = %v, want %v", w.Header().Get("Location"), tt.location)
				}

				if w.Code != tt.statusCode {
					t.Errorf("Status code = %d, want %d", w.Code, tt.statusCode)
				}
			}
		})
	}
}

func TestContext_Status(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	ctx.Status(201)

	if ctx.GetStatusCode() != 201 {
		t.Errorf("GetStatusCode() = %d, want 201", ctx.GetStatusCode())
	}
}

func TestContext_Headers(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Authorization", "Bearer token123")
	ctx := NewContext(w, r)

	// Test Get (interface method)
	if got := ctx.GetHeader("Authorization"); got != "Bearer token123" {
		t.Errorf("Get('Authorization') = %v, want 'Bearer token123'", got)
	}

	// Test GetHeader (alias)
	if got := ctx.GetHeader("Authorization"); got != "Bearer token123" {
		t.Errorf("GetHeader('Authorization') = %v, want 'Bearer token123'", got)
	}

	// Test Set (interface method)
	ctx.SetHeader("X-Custom-Header", "custom-value")
	if got := w.Header().Get("X-Custom-Header"); got != "custom-value" {
		t.Errorf("Set didn't set header correctly, got %v", got)
	}

	// Test SetHeader (alias)
	ctx.SetHeader("X-Another-Header", "another-value")
	if got := w.Header().Get("X-Another-Header"); got != "another-value" {
		t.Errorf("SetHeader didn't set header correctly, got %v", got)
	}
}

func TestContext_Values(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	// Set value
	ctx.SetValue("user", "john")
	ctx.SetValue("role", "admin")

	// Get value
	if got := ctx.GetValue("user"); got != "john" {
		t.Errorf("GetValue('user') = %v, want 'john'", got)
	}

	if got := ctx.GetValue("role"); got != "admin" {
		t.Errorf("GetValue('role') = %v, want 'admin'", got)
	}

	// Get non-existent value
	if got := ctx.GetValue("missing"); got != nil {
		t.Errorf("GetValue('missing') should return nil, got %v", got)
	}
}

func TestContext_Method(t *testing.T) {
	tests := []struct {
		method string
	}{
		{"GET"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, "/test", nil)
			ctx := NewContext(w, r)

			if got := ctx.Method(); got != tt.method {
				t.Errorf("Method() = %v, want %v", got, tt.method)
			}
		})
	}
}

func TestContext_Path(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/123?name=john", nil)
	ctx := NewContext(w, r)

	if got := ctx.Path(); got != "/users/123" {
		t.Errorf("Path() = %v, want '/users/123'", got)
	}
}

func TestContext_Host(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com:8080/test", nil)
	ctx := NewContext(w, r)

	if got := ctx.Host(); got != "example.com:8080" {
		t.Errorf("Host() = %v, want 'example.com:8080'", got)
	}
}

func TestContext_ClientIP(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(*http.Request)
		expectedIP string
	}{
		{
			name: "X-Forwarded-For single IP",
			setupFunc: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			setupFunc: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
			},
			expectedIP: "203.0.113.1",
		},
		{
			name: "X-Real-IP",
			setupFunc: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "198.51.100.1")
			},
			expectedIP: "198.51.100.1",
		},
		{
			name: "RemoteAddr fallback",
			setupFunc: func(r *http.Request) {
				r.RemoteAddr = "192.0.2.1:12345"
			},
			expectedIP: "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			tt.setupFunc(r)
			ctx := NewContext(w, r)

			if got := ctx.ClientIP(); got != tt.expectedIP {
				t.Errorf("ClientIP() = %v, want %v", got, tt.expectedIP)
			}
		})
	}
}

func TestContext_UserAgent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0 (Test Browser)")
	ctx := NewContext(w, r)

	if got := ctx.UserAgent(); got != "Mozilla/5.0 (Test Browser)" {
		t.Errorf("UserAgent() = %v, want 'Mozilla/5.0 (Test Browser)'", got)
	}
}

func TestContext_FormValue(t *testing.T) {
	// Test URL query parameter
	t.Run("QueryParameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test?name=john", nil)
		ctx := NewContext(w, r)

		if got := ctx.FormValue("name"); got != "john" {
			t.Errorf("FormValue('name') = %v, want 'john'", got)
		}
	})

	// Test POST form data
	t.Run("PostFormData", func(t *testing.T) {
		formData := "name=john&age=30"
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/test", strings.NewReader(formData))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		ctx := NewContext(w, r)

		if got := ctx.FormValue("name"); got != "john" {
			t.Errorf("FormValue('name') = %v, want 'john'", got)
		}
	})
}

func TestContext_FormFile(t *testing.T) {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add a file
	fileWriter, err := writer.CreateFormFile("upload", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	fileContent := []byte("test file content")
	if _, err := fileWriter.Write(fileContent); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	writer.Close()

	// Create request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := NewContext(w, r)

	// Get file
	file, header, err := ctx.FormFile("upload")
	if err != nil {
		t.Fatalf("FormFile() error = %v", err)
	}
	defer file.Close()

	if header.Filename != "test.txt" {
		t.Errorf("Filename = %v, want 'test.txt'", header.Filename)
	}

	// Read and verify content
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !bytes.Equal(content, fileContent) {
		t.Error("File content doesn't match")
	}
}

func TestContext_Cookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	ctx := NewContext(w, r)

	// Get cookie
	value, err := ctx.Cookie("session")
	if err != nil {
		t.Fatalf("Cookie() error = %v", err)
	}

	if value != "abc123" {
		t.Errorf("Cookie value = %v, want 'abc123'", value)
	}

	// Get non-existent cookie
	_, err = ctx.Cookie("missing")
	if err == nil {
		t.Error("Cookie() should return error for non-existent cookie")
	}
}

func TestContext_SetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	cookie := &http.Cookie{
		Name:     "session",
		Value:    "abc123",
		MaxAge:   3600,
		HttpOnly: true,
	}

	ctx.SetCookie(cookie)

	// Check if cookie was set in response
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	if cookies[0].Name != "session" || cookies[0].Value != "abc123" {
		t.Error("Cookie was not set correctly")
	}
}

func TestContext_IsWebSocket(t *testing.T) {
	tests := []struct {
		name        string
		setupHeader func(*http.Request)
		want        bool
	}{
		{
			name: "WebSocket request",
			setupHeader: func(r *http.Request) {
				r.Header.Set("Upgrade", "websocket")
			},
			want: true,
		},
		{
			name: "Regular HTTP request",
			setupHeader: func(r *http.Request) {
				// No upgrade header
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			tt.setupHeader(r)
			ctx := NewContext(w, r)

			if got := ctx.IsWebSocket(); got != tt.want {
				t.Errorf("IsWebSocket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_IsAjax(t *testing.T) {
	tests := []struct {
		name        string
		setupHeader func(*http.Request)
		want        bool
	}{
		{
			name: "AJAX request",
			setupHeader: func(r *http.Request) {
				r.Header.Set("X-Requested-With", "XMLHttpRequest")
			},
			want: true,
		},
		{
			name: "Regular request",
			setupHeader: func(r *http.Request) {
				// No X-Requested-With header
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			tt.setupHeader(r)
			ctx := NewContext(w, r)

			if got := ctx.IsAjax(); got != tt.want {
				t.Errorf("IsAjax() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContext_Accepts(t *testing.T) {
	tests := []struct {
		name         string
		acceptHeader string
		contentType  string
		want         bool
	}{
		{"JSON accepted", "application/json", "application/json", true},
		{"Wildcard accepted", "*/*", "text/html", true},
		{"Not accepted", "application/json", "text/html", false},
		{"Multiple types", "text/html, application/json", "application/json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			r.Header.Set("Accept", tt.acceptHeader)
			ctx := NewContext(w, r)

			if got := ctx.Accepts(tt.contentType); got != tt.want {
				t.Errorf("Accepts(%v) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestContext_Next(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	var executionOrder []string

	// Create handler chain
	middleware1 := func(ctx Context) error {
		executionOrder = append(executionOrder, "middleware1-before")
		err := ctx.Next()
		executionOrder = append(executionOrder, "middleware1-after")
		return err
	}

	middleware2 := func(ctx Context) error {
		executionOrder = append(executionOrder, "middleware2-before")
		err := ctx.Next()
		executionOrder = append(executionOrder, "middleware2-after")
		return err
	}

	finalHandler := func(ctx Context) error {
		executionOrder = append(executionOrder, "handler")
		return nil
	}

	handlers := []HandlerFunc{middleware1, middleware2, finalHandler}
	ctx.SetHandlers(handlers)

	// Execute
	if err := ctx.Next(); err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	// Verify execution order
	expected := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	if len(executionOrder) != len(expected) {
		t.Fatalf("Execution order length = %d, want %d", len(executionOrder), len(expected))
	}

	for i, step := range executionOrder {
		if step != expected[i] {
			t.Errorf("Execution order[%d] = %v, want %v", i, step, expected[i])
		}
	}
}

func TestContext_Reset(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	// Set some state
	ctx.SetParam("id", "123")
	ctx.SetValue("user", "john")
	ctx.Status(201)

	// Reset with new request/response
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("POST", "/test2", nil)
	ctx.Reset(w2, r2)

	// Verify reset
	if ctx.Request() != r2 {
		t.Error("Request should be updated after Reset")
	}

	if ctx.Response() != w2 {
		t.Error("Response should be updated after Reset")
	}

	if ctx.GetStatusCode() != http.StatusOK {
		t.Errorf("Status code should be reset to 200, got %d", ctx.GetStatusCode())
	}

	if ctx.Param("id") != "" {
		t.Error("Params should be cleared after Reset")
	}

	if ctx.GetValue("user") != nil {
		t.Error("Values should be cleared after Reset")
	}

	if ctx.IsWritten() {
		t.Error("IsWritten should be reset to false")
	}
}

func TestContext_WithContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	// Create new context with timeout
	newCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create new AppContext with new context.Context
	ctxWithContext := ctx.WithContext(newCtx)

	if ctxWithContext.Context() != newCtx {
		t.Error("WithContext() should return AppContext with new context")
	}

	// Verify original context is unchanged
	if ctx.Context() == newCtx {
		t.Error("Original context should not be modified")
	}
}

func TestContext_Write(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	data := []byte("Hello, World!")
	n, err := ctx.Write(data)

	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if n != len(data) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(data))
	}

	if !ctx.IsWritten() {
		t.Error("IsWritten() should be true after Write()")
	}

	if !bytes.Equal(w.Body.Bytes(), data) {
		t.Error("Response body doesn't match written data")
	}
}

func TestContext_Stream(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/stream", nil)
	ctx := NewContext(w, r)

	step := func(w io.Writer) error {
		var err error
		for i := 0; i < 4; i++ {
			_, err = w.Write([]byte(fmt.Sprintf("chunk%d,", i)))
		}
		return err
	}

	err := ctx.Stream(200, "text/plain", step)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	expected := "chunk0,chunk1,chunk2,chunk3,"
	if w.Body.String() != expected {
		t.Errorf("Response body = %s, want %s", w.Body.String(), expected)
	}

	if !ctx.IsWritten() {
		t.Error("IsWritten() should be true after Stream()")
	}
}

func TestContext_Err(t *testing.T) {
	// Test with cancelled context
	t.Run("CancelledContext", func(t *testing.T) {
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil).WithContext(cancelledCtx)
		ctx := NewContext(w, r)

		if ctx.Err() != context.Canceled {
			t.Errorf("Err() = %v, want context.Canceled", ctx.Err())
		}
	})

	// Test with normal context
	t.Run("NormalContext", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := NewContext(w, r)

		if ctx.Err() != nil {
			t.Errorf("Err() = %v, want nil", ctx.Err())
		}
	})
}

// Benchmark tests
func BenchmarkContext_Param(b *testing.B) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/123", nil)
	ctx := NewContext(w, r)
	ctx.SetParam("id", "123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.Param("id")
	}
}

func BenchmarkContext_JSON(b *testing.B) {
	data := map[string]string{
		"message": "Hello, World!",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/test", nil)
		ctx := NewContext(w, r)
		_ = ctx.JSON(200, data)
	}
}

func BenchmarkContext_SetValue(b *testing.B) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.SetValue("key", "value")
	}
}

func BenchmarkContext_GetValue(b *testing.B) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx := NewContext(w, r)
	ctx.SetValue("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.GetValue("key")
	}
}
