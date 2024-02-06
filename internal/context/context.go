package context

import (
	"github.com/labstack/echo/v4"

	"github.com/pluhe7/shortener/internal/app"
)

type Context struct {
	echo.Context

	Server        *app.Server
	SessionUserID string
}

func NewContext(c echo.Context, server *app.Server) *Context {
	cc := &Context{
		Context: c,
		Server:  server,
	}

	return cc
}
