package routes

import (
	"dd-opcda/engine"

	"github.com/gofiber/fiber/v2"
)

type SystemInformation struct {
	GitVersion string           `json:"gitversion"`
	GitCommit  string           `json:"gitcommit"`
	CacheInfo  engine.CacheInfo `json:"cacheinfo"`
}

var SysInfo SystemInformation

func RegisterSystemRoutes(api fiber.Router) {
	api.Get("/system/info", GetSysInfo)
	api.Get("/system/sendhistory", SendFullCache)
	api.Post("/system/resend", ResendCacheItems)
}

func GetSysInfo(c *fiber.Ctx) error {
	SysInfo.CacheInfo = engine.GetCacheInfo()
	return c.Status(fiber.StatusOK).JSON(SysInfo)
}

func ResendCacheItems(c *fiber.Ctx) error {
	var items []engine.CacheItem
	if err := c.BodyParser(&items); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{"error": err.Error()})
	}

	count := engine.ResendCacheItems(items)
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{"count": count})
}

func SendFullCache(c *fiber.Ctx) error {
	if err := engine.SendFullCache(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{"status": "ok"})
}
