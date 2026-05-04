package main

import (
	"github.com/k1v4/drip_mate/docs"
	"github.com/k1v4/drip_mate/internal/app"
)

// @title           Drip Mate API
// @version         1.0
// @description     Outfit recommendation service
// @host
// @BasePath        /api/v1
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
func main() {
	docs.SwaggerInfo.BasePath = "/api/v1"

	app.Run()
}
