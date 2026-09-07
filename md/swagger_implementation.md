# Swagger Implementation Guide for Golang (Echo v5)

This guide provides a step-by-step walkthrough of how Swagger (OpenAPI 2.0) was implemented in this project. You can follow these exact steps to implement Swagger in any future Golang projects.

## Step 1: Install Dependencies

First, you need to install the `swag` Command Line Interface (CLI) tool globally, and the `http-swagger` library in your project.

```bash
# Install the Swag CLI tool
go install github.com/swaggo/swag/cmd/swag@latest

# Install the swagger HTTP handler library
go get -u github.com/swaggo/http-swagger/v2
```

## Step 2: Add General API Info to `main.go`

In your application's entry point (`cmd/main.go`), you need to define the general information about your API. These are special comments that `swag` reads.

**File:** `cmd/main.go`
```go
package main

import (
    // ... your imports
)

// @title           GoTickets API
// @version         1.0
// @description     A professional backend service for a ticket booking platform.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@gotickets.local

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:5000
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
    // ... your initialization code
}
```
*Note: The `@securityDefinitions.apikey` block is critical if your API uses JWT/Bearer tokens.*

## Step 3: Annotate Your HTTP Handlers

For every endpoint you want to document, you must add annotations immediately above the handler function. 

**File:** `internal/domain/user/handler.go` *(Example)*
```go
// CreateUser godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns access and refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateRequest  true  "User Registration Details"
// @Success      200      {object}  dto.Response
// @Failure      400      {object}  httpresponse.Error
// @Router       /api/v1/auth/register [post]
func (h *handler) CreateUser(c *echo.Context) error {
    // ... your logic
}
```

**For Protected Routes (Requires JWT):**
Add the `@Security` tag to enforce authentication in the Swagger UI.
```go
// GetMyBookings godoc
// @Summary      List my bookings
// @Tags         Bookings
// @Security     BearerAuth
// @Success      200      {array}   dto.Response
// @Router       /api/v1/bookings/me [get]
func (h *handler) GetMyBookings(c *echo.Context) error {
    // ...
}
```

## Step 4: Configure the Swagger Route in Echo

Next, we need to expose the Swagger UI dashboard through a route in your server setup file. 

**File:** `internal/server/http.go`
```go
package server

import (
    // 1. Import the generated docs folder (even if it doesn't exist yet, we will generate it in Step 5)
    _ "gotickets/docs"
    
    // 2. Import the http-swagger library
    httpSwagger "github.com/swaggo/http-swagger/v2"
    "github.com/labstack/echo/v5"
)

func Start(...) {
    e := echo.New()
    
    // ... your normal routes

    // 3. Mount the Swagger UI Handler using Echo's WrapHandler
    e.GET("/swagger/*", echo.WrapHandler(httpSwagger.Handler()))
    
    // ... start the server
}
```

## Step 5: Generate the Documentation

Finally, you must run the `swag init` command to parse all your comments and generate the `docs` folder. 

Run this command in the root of your project:
```bash
swag init -g cmd/api_info.go --parseDependency --parseInternal
```

**Why these flags?**
- `-g cmd/main.go`: Tells swag where the General API Info (Step 2) is located.
- `--parseDependency`: Forces swag to parse external dependencies (like third-party JWT structs).
- `--parseInternal`: Forces swag to parse files inside your `internal/` folder.

If successful, you will see a `docs` folder created with `docs.go`, `swagger.json`, and `swagger.yaml`. 

You can now start your server and visit `http://localhost:5000/swagger/index.html`!
