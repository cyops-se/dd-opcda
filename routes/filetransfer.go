package routes

import (
	"dd-opcda/db"
	"dd-opcda/engine"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
)

func RegisterFileTransferRoutes(api fiber.Router) {
	api.Get("/filetransfer", GetFileTransferInfo)
	api.Post("/filetransfer", PostFileTransferInfo)
	api.Post("/filetransfer/upload", UploadFilesToTransfer)
}

func GetFileTransferInfo(c *fiber.Ctx) error {
	SysInfo.CacheInfo = engine.GetCacheInfo()
	return c.Status(fiber.StatusOK).JSON(SysInfo)
}

func PostFileTransferInfo(c *fiber.Ctx) error {
	SysInfo.CacheInfo = engine.GetCacheInfo()
	return c.Status(fiber.StatusOK).JSON(SysInfo)
}

func UploadFilesToTransfer(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		db.Log("error", "No file provided", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	log.Printf("FILE TRANSFER: received file from upload:", file.Filename)

	// Make sure file transfer outging exists
	// TODO: use proper interface to filetransfer for this (now hardcoded to ./outgoing/new)
	if _, err := os.Stat("./outgoing/new"); os.IsNotExist(err) {
		os.MkdirAll("./outgoing/new", 0755)
	}

	filename := path.Join("./outgoing/new", file.Filename) //fmt.Sprintf("./outgoing/new/%s", file.Filename)
	if err := c.SaveFile(file, filename); err != nil {
		msg := fmt.Sprintf("failed to save file, name: '%s', size: %d, error: %s", file.Filename, file.Size, err.Error())
		db.Log("error", "Upload of file to transfer failed", msg)
		return c.Status(503).SendString(msg)
	} else {
		db.Log("trace", "Import meta request", fmt.Sprintf("name: '%s', size: %d", file.Filename, file.Size))
	}

	return c.Status(http.StatusOK).JSON(file)
}
