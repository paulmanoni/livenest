package core

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Context wraps gin.Context with additional LiveNest features
type Context struct {
	*gin.Context
	app *App
}

// NewContext creates a new LiveNest context from gin.Context
func NewContext(c *gin.Context, app *App) *Context {
	return &Context{
		Context: c,
		app:     app,
	}
}

// DB returns the GORM database instance
func (c *Context) DB() *gorm.DB {
	return c.app.DB
}

// App returns the LiveNest app instance
func (c *Context) App() *App {
	return c.app
}

// WithDB creates a middleware that injects LiveNest context
func (a *App) WithDB() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := NewContext(c, a)
		c.Set("livenest", ctx)
		c.Next()
	}
}

// GetContext retrieves the LiveNest context from gin.Context
func GetContext(c *gin.Context) *Context {
	if ctx, exists := c.Get("livenest"); exists {
		return ctx.(*Context)
	}
	return nil
}