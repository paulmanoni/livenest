package core

import (
	"log"

	"livenest/liveview"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App is the main application structure wrapping Gin and GORM
type App struct {
	Router        *gin.Engine
	DB            *gorm.DB
	config        *Config
	lvHandler     *liveview.Handler
	webComponents map[string]liveview.WebComponentConfig
}

// New creates a new LiveNest application
func New(config *Config) *App {
	if config == nil {
		config = DefaultConfig()
	}

	// Set Gin mode
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	app := &App{
		Router: gin.Default(),
		config: config,
	}

	// Serve LiveNest static files
	app.setupLiveNestStatic()

	return app
}

// setupLiveNestStatic serves the LiveView JavaScript files
func (a *App) setupLiveNestStatic() {
	// Ensure LiveView handler exists
	if a.lvHandler == nil {
		a.lvHandler = liveview.NewHandler()
	}

	// Serve embedded LiveView JavaScript (includes component tag)
	a.Router.GET("/livenest/liveview.js", func(c *gin.Context) {
		c.Header("Content-Type", "application/javascript")
		c.String(200, liveview.GetLiveViewJS())
	})

	// Serve web components JavaScript
	a.Router.GET("/livenest/components.js", func(c *gin.Context) {
		c.Header("Content-Type", "application/javascript")
		c.String(200, a.GetWebComponentsJS())
	})

	// Handle component tag requests
	a.Router.GET("/livenest/component/:name", a.lvHandler.HandleComponentTag)
}

// ConnectDB connects to the database using GORM
func (a *App) ConnectDB(dialector gorm.Dialector, opts ...gorm.Option) error {
	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return err
	}

	a.DB = db
	return nil
}

// Use adds middleware to the Gin router
func (a *App) Use(middleware ...gin.HandlerFunc) {
	a.Router.Use(middleware...)
}

// Group creates a new router group
func (a *App) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return a.Router.Group(relativePath, handlers...)
}

// GET is a shortcut for router.Handle("GET", path, handlers)
func (a *App) GET(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return a.Router.GET(path, handlers...)
}

// POST is a shortcut for router.Handle("POST", path, handlers)
func (a *App) POST(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return a.Router.POST(path, handlers...)
}

// PUT is a shortcut for router.Handle("PUT", path, handlers)
func (a *App) PUT(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return a.Router.PUT(path, handlers...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handlers)
func (a *App) DELETE(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return a.Router.DELETE(path, handlers...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handlers)
func (a *App) PATCH(path string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return a.Router.PATCH(path, handlers...)
}

// Run starts the HTTP server
func (a *App) Run(addr ...string) error {
	address := ":8080"
	if len(addr) > 0 {
		address = addr[0]
	}

	log.Printf("LiveNest server starting on %s", address)
	return a.Router.Run(address)
}

// GetDB returns the GORM database instance
func (a *App) GetDB() *gorm.DB {
	return a.DB
}

// GetConfig returns the app configuration
func (a *App) GetConfig() *Config {
	return a.config
}

// RegisterComponent registers a LiveView component
func (a *App) RegisterComponent(name string, component liveview.Component) {
	if a.lvHandler == nil {
		a.lvHandler = liveview.NewHandler()
	}
	a.lvHandler.Register(name, component)
}
