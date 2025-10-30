package core

import "testing"

func TestHTTPMethod_String(t *testing.T) {
	tests := []struct {
		name     string
		method   HTTPMethod
		expected string
	}{
		{"GET method", MethodGET, "GET"},
		{"POST method", MethodPOST, "POST"},
		{"PUT method", MethodPUT, "PUT"},
		{"DELETE method", MethodDELETE, "DELETE"},
		{"PATCH method", MethodPATCH, "PATCH"},
		{"OPTIONS method", MethodOPTIONS, "OPTIONS"},
		{"HEAD method", MethodHEAD, "HEAD"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.method.String(); got != test.expected {
				t.Errorf("HTTPMethod.String() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestProviderScope_String(t *testing.T) {
	tests := []struct {
		name     string
		scope    ProviderScope
		expected string
	}{
		{"Singleton scope", SingletonScope, "Singleton"},
		{"Transient scope", TransientScope, "Transient"},
		{"Request scope", RequestScope, "Request"},
		{"Unknown scope", ProviderScope(999), "Unknown"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.scope.String(); got != test.expected {
				t.Errorf("ProviderScope.String() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"Debug level", LogLevelDebug, "DEBUG"},
		{"Info level", LogLevelInfo, "INFO"},
		{"Warn level", LogLevelWarn, "WARN"},
		{"Error level", LogLevelError, "ERROR"},
		{"Fatal level", LogLevelFatal, "FATAL"},
		{"Unknown level", LogLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.level.String(); got != test.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestDefaultConfigOptions(t *testing.T) {
	opts := DefaultConfigOptions()

	if opts.Port != 3000 {
		t.Errorf("DefaultConfigOptions().Port = %v, want %v", opts.Port, 3000)
	}
	if opts.Host != "0.0.0.0" {
		t.Errorf("DefaultConfigOptions().Host = %v, want %v", opts.Host, "0.0.0.0")
	}
	if opts.Environment != "development" {
		t.Errorf("DefaultConfigOptions().Environment = %v, want %v", opts.ReadTimeout, "development")
	}
	if opts.EnableCors != false {
		t.Errorf("DefaultConfigOptions().EnableCors = %v, want %v", opts.EnableCors, false)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errors:   ValidationErrors{},
			expected: "validation failed",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "email", Message: "email is required"},
			},
			expected: "email is required",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "email", Message: "email is required"},
				{Field: "name", Message: "name is required"},
			},
			expected: "email is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.errors.Error(); got != test.expected {
				t.Errorf("ValidationErrors.Error() = %v, want %v", got, test.expected)
			}
		})
	}
}

func TestRouteMetadata(t *testing.T) {
	meta := PaginationMetadata{
		Page:       1,
		Limit:      10,
		Total:      100,
		TotalPages: 10,
		HasNext:    true,
		HasPrev:    false,
	}

	if meta.Page != 1 {
		t.Errorf("PaginationMetadata.Page = %v, want %v", meta.Page, 1)
	}

	if meta.TotalPages != 10 {
		t.Errorf("PaginationMetadata.TotalPages = %v, want %v", meta.TotalPages, 10)
	}
	if meta.HasNext != true {
		t.Errorf("PaginationMetadata.HasNext = %v, want %v", meta.HasNext, true)
	}
	if meta.HasPrev != false {
		t.Errorf("PaginationMetadata.HasPrev = %v, want %v", meta.HasPrev, false)
	}
}

func TestErrorResponse(t *testing.T) {
	err := ErrorResponse{
		StatusCode: 404,
		Message:    "Not Found",
		Error:      "NotFoundException",
		Path:       "/users/123",
		Timestamp:  "2025-10-30T00:00:00Z",
	}

	if err.StatusCode != 404 {
		t.Errorf("ErrorResponse.StatusCode = %v, want %v", err.StatusCode, 404)
	}
	if err.Message != "Not Found" {
		t.Errorf("ErrorResponse.Message = %v, want %v", err.Message, "Not Found")
	}
	if err.Error != "NotFoundException" {
		t.Errorf("ErrorResponse.Error = %v, want %v", err.Error, "NotFoundException")
	}
	if err.Path != "/users/123" {
		t.Errorf("ErrorResponse.Path = %v, want %v", err.Path, "/users/123")
	}
}

func TestSuccessResponse(t *testing.T) {
	resp := SuccessResponse{
		StatusCode: 200,
		Message:    "Success",
		Data:       map[string]string{"id": "123"},
	}

	if resp.StatusCode != 200 {
		t.Errorf("SuccessResponse.StatusCode = %v, want %v", resp.StatusCode, 200)
	}
	if resp.Message != "Success" {
		t.Errorf("SuccessResponse.Message = %v, want %v", resp.Message, "Success")
	}
}
