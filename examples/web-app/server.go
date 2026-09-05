package main

import (
	"fmt"
	"net/http"

	"github.com/joysriramsarkar/alap-framework/web/router"
	"github.com/joysriramsarkar/alap-framework/web/server"
)

func main() {
	app := server.New("EnterpriseWebApp")

	app.Router.GET("/", func(ctx *router.Context) (interface{}, error) {
		return map[string]interface{}{
			"app":     "Alap Enterprise Web",
			"version": "1.0.0",
			"engine":  "Nilang Language",
			"status":  "active",
		}, nil
	})

	app.Router.GET("/api/v1/users/{id:int}", func(ctx *router.Context) (interface{}, error) {
		id, _ := ctx.ParamInt("id")
		return map[string]interface{}{
			"id":    id,
			"name":  fmt.Sprintf("Enterprise User %d", id),
			"email": fmt.Sprintf("user%d@nilang.org", id),
		}, nil
	})

	fmt.Println("🚀 [Alap Web] Server listening on http://localhost:8080")
	if err := app.Listen(":8080"); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Error: %v\n", err)
	}
}
