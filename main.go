package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

func main() {
	e := echo.New()
	e.Static("/", "public")
	e.File("/", "public/index.html")
	e.Run(standard.New(":3000"))
	//e.Run(standard.WithTLS(
	//	":443",
	//	"keys/www.catchupbot.com/fullchain.pem",
	//	"keys/www.catchupbot.com/privkey.pem"))
}
