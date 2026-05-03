package main

import "github.com/k1v4/drip_mate/internal/app"

// @title           Drip Mate API
// @version         1.0
// @description     Outfit recommendation service
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
func main() {
	app.Run()
}
