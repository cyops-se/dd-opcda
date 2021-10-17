package main

import (
	"dd-opcda/engine"
	"dd-opcda/routes"
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

func RunWeb() {
	defer handlePanic()

	// http.FS can be used to create a http Filesystem
	subFS2, _ := fs.Sub(static, "static")
	var staticFS = http.FS(subFS2)

	app := fiber.New(fiber.Config{StrictRouting: true})
	app.Use(logger.New())

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
		engine.RegisterWebsocket(c)
		for c.Conn != nil {
			time.Sleep(1)
		}
		log.Println("Dropping websocket connection")
	}))

	api := app.Group("/api")
	routes.RegisterAuthRoutes(api)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte("897puihjöknawerthgfp7<yvalknp98h"),
	}))

	routes.RegisterUserRoutes(api)
	routes.RegisterDataRoutes(api)
	routes.RegisterOPCRoutes(api)
	routes.RegisterSystemRoutes(api)

	app.Listen(":3000")

	select {}
}
