## 🚀 go-cloud-k8s-common


[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=lao-tseu-is-alive_go-cloud-k8s-common&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=lao-tseu-is-alive_go-cloud-k8s-common)
[![cve-trivy-scan](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common/actions/workflows/cve-trivy-scan.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common/actions/workflows/cve-trivy-scan.yml)
[![codecov](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-common/branch/main/graph/badge.svg)](https://codecov.io/gh/lao-tseu-is-alive/go-cloud-k8s-common)
[![Go Test](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common/actions/workflows/go-test.yml/badge.svg)](https://github.com/lao-tseu-is-alive/go-cloud-k8s-common/actions/workflows/go-test.yml)


Common Golang packages for other MicroServices in our goeland team.

**A modular, production-ready toolkit for building cloud-native Go web services and APIs.**

---

## Overview

`go-cloud-k8s-common` is a collection of Go packages designed to accelerate the development of robust, observable, and secure microservices, especially those targeting cloud and Kubernetes environments. It provides a powerful, modular HTTP server that enhances Go's standard `net/http` library with essential production-grade features right out of the box.

Whether you need a simple web server with structured logging or a full-fledged microservice with JWT authentication and database integration, this library offers the flexibility to use only the components you need, thanks to the functional options pattern.

---

## ✨ Features

* **Modular HTTP Server**: Build anything from a minimal web server to a complex API.
* **SSL Support**: HTTPS with autocert or manual config options with your certificate. 
* **JWT Authentication**: Secure your endpoints with a built-in JWT middleware.
* **Prometheus Metrics**: Export critical metrics for monitoring and alerting.
* **Structured Logging**: Keep your logs clean and easy to parse with `golog`.
* **Configuration Management**: Easily configure your application using environment variables.
* **Database Helpers**: Simplify your database interactions with helpers for PostgreSQL.
* **Kubernetes Awareness**: Automatically gather information about the Kubernetes environment.
* **Graceful Shutdown**: Ensure your services shut down cleanly without dropping connections.

---

## 🏁 Getting Started

Here's how to create a minimal, production-ready web server in just a few lines of code.

### Minimal Server Example

This example creates a simple web server that listens on port `8080` and responds with "Hello, World!" at the `/hello` endpoint.

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/gohttp"
    "github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
)

func main() {
    // 1. Initialize the logger
    l, err := golog.NewLogger("simple", golog.DebugLevel, "my-minimal-app")
    if err != nil {
        log.Fatalf("unable to create logger: %v", err)
    }

    // 2. Create a new server
    server := gohttp.NewGoHttpServer(":8080", l)

    // 3. Add a simple handler
    server.AddRoute("GET /hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    }))

    // 4. Start the server
    l.Info("Starting server on :8080")
    server.StartServer()
}
````

-----

## 🛠️ Advanced Usage

For more complex applications, you can easily add authentication, versioning, and protected routes using the functional options pattern.

### Full-Fledged Server with Protected Routes

This example demonstrates how to create a server with JWT authentication and custom protected routes.

```go
package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/config"
    "github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/gohttp"
    "github.com/lao-tseu-is-alive/go-cloud-k8s-common/pkg/golog"
)

func main() {
    // 1. Initialize logger and other components
    l, err := golog.NewLogger("simple", golog.DebugLevel, "my-app")
    if err != nil {
        log.Fatalf("unable to create logger: %v", err)
    }

    jwtContextKey := config.GetJwtContextKeyFromEnvOrPanic()
    myJwt := gohttp.NewJwtChecker(
        config.GetJwtSecretFromEnvOrPanic(),
        config.GetJwtIssuerFromEnvOrPanic(),
        "my-app",
        jwtContextKey,
        config.GetJwtDurationFromEnvOrPanic(60),
        l)

    myAuthenticator := gohttp.NewSimpleAdminAuthenticator(
        config.GetAdminUserFromEnvOrPanic("admin"),
        config.GetAdminPasswordFromEnvOrPanic(),
        config.GetAdminEmailFromEnvOrPanic("admin@example.com"),
        config.GetAdminIdFromEnvOrPanic(1),
        myJwt)

    myVersionReader := gohttp.NewSimpleVersionReader("my-app", "1.0.0", "", "", "", "/login")

    // 2. Define your protected routes
    protectedAPI := map[string]http.HandlerFunc{
        "GET /api/v1/data": func(w http.ResponseWriter, r *http.Request) {
            claims := myJwt.GetJwtCustomClaimsFromContext(r.Context())
            fmt.Fprintf(w, "Hello, %s! Here is your protected data.", claims.User.UserName)
        },
    }

    // 3. Create the server with all the options
    server := gohttp.NewGoHttpServer(
        ":8080",
        l,
        gohttp.WithAuthentication(myAuthenticator),
        gohttp.WithProtectedRoutes(myJwt, protectedAPI),
        gohttp.WithVersionReader(myVersionReader),
    )

    // 4. Start the server
    l.Info("Starting server on :8080")
    server.StartServer()
}
```

-----

## ⚙️ Configuration

The library is configured through environment variables. For a complete list of available variables, please refer to the `.env_sample` file.

-----

## 🙌 Contributing

Contributions are welcome\! If you have a suggestion or find a bug, please open an issue or submit a pull request.

```
```