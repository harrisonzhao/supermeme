package main

import (
	"github.com/labstack/echo"
)

const (
	publicDir = "public"
)

func main() {
	e := echo.New()
	e.Static("/", publicDir)
	e.File("/", publicDir + "/index.html")
	//e.Start(":3000")
	//e.StartTLS(":443", "keys/www.catchupbot.com/fullchain.pem", "keys/www.catchupbot.com/privkey.pem")
}
