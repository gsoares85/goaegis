# GoAegis Framework

🚀 A progressive Go framework for building efficient and scalable server-side applications.

## Vision

GoAegis aims to bring the elegant architecture and developer experience of NestJS to the Go ecosystem.

## Features (Roadmap)

### Core Features
- ✅ HTTP Routing with decorators-like pattern
- ✅ Dependency Injection Container
- ✅ Modular Architecture
- ✅ Controllers and Providers
- ✅ Middleware Support
- ✅ Exception Filters
- ✅ Pipes for Validation
- ✅ Guards for Authorization
- ✅ Interceptors for AOP

### Advanced Features
- 📝 Configuration Management
- 📊 Logger Service
- 🔐 Authentication & Authorization
- ✅ Validation with DTOs
- 📚 OpenAPI/Swagger Documentation
- 🧪 Testing Utilities
- 💾 Cache Manager
- ⏰ Task Scheduling
- 📡 Event Emitter
- 🔌 WebSockets Support
- 🗄️ Database Integration

## Quick Start

### Installation
```bash
go get github.com/gsoares85/goaegis
```

### Basic Usage
```go
package main

import (
    "github.com/gsoares85/goaegis/pkg/core"
    "github.com/gsoares85/goaegis/pkg/module"
)

func main() {
    app := core.NewApplication()
    app.RegisterModule(module.NewAppModule())
    app.Listen(":3000")
}
```

## Development
```bash
# Install dependencies
make install

# Run tests
make test

# Build project
make build

# See all commands
make help
```

## Documentation

See [docs/](./docs/) for complete documentation.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the MIT License - see [LICENSE](./LICENSE) for details.

## Acknowledgments

Inspired by [NestJS](https://nestjs.com/)

---

Made with ❤️ for the Go community
