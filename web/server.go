package web

import (
	ddlog "dd-opcda/logger"
	"dd-opcda/routes"
	"dd-opcda/types"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/gofiber/websocket/v2"
)

//go:embed static/index.html
var admin string

//go:embed static/*
var static embed.FS

func handlePanic() {
	if r := recover(); r != nil {
		// ddlog.Error("RunWeb", "Panic, recovery: %#v", r)
		log.Printf("Servers panic, recovery: %#v", r)
		return
	}
}

func RunWeb(args types.Context) {
	defer handlePanic()

	// http.FS can be used to create a http Filesystem
	subFS2, _ := fs.Sub(static, "static")
	var staticFS = http.FS(subFS2)

	// Set a file transfer limit to 50MB
	app := fiber.New(fiber.Config{StrictRouting: true, BodyLimit: 50 * 1024 * 1024})
	if args.Trace {
		app.Use(logger.New())
	}

	app.Use("/", filesystem.New(filesystem.Config{
		Root:   staticFS,
		Browse: false,
	}))

	app.Get("/ui/*", func(ctx *fiber.Ctx) error {
		ctx.Status(200)
		ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		// return ctx.Send([]byte(admin))
		return ctx.SendString(admin)
	})

	app.Use("/static", filesystem.New(filesystem.Config{
		Root:   staticFS,
		Browse: false,
	}))

	// WebSocket registration
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		ddlog.RegisterWebsocket(c)
		for c.Conn != nil {
			time.Sleep(1)
		}
	}))

	api := app.Group("/api")
	routes.RegisterAuthRoutes(api)
	api.Get("/system/info", routes.GetSysInfo)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("897puihjöknawerthgfp7<yvalknp98h"),
	}))

	routes.RegisterUserRoutes(api)
	routes.RegisterDataRoutes(api)
	routes.RegisterOPCRoutes(api)
	routes.RegisterSystemRoutes(api)
	routes.RegisterFileTransferRoutes(api)

	app.Listen(":3000")

	select {}
}
