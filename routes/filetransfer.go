package routes

import (
	"dd-opcda/engine"
	"dd-opcda/logger"
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
		logger.Log("error", "No file provided", err.Error())
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
		// msg := fmt.Sprintf("failed to save file, name: '%s', size: %d, error: %s", file.Filename, file.Size, err.Error())
		e := logger.Error("Upload of file to transfer failed", "failed to save file, name: '%s', size: %d, error: %s", file.Filename, file.Size, err.Error())
		return c.Status(http.StatusInternalServerError).JSON(&fiber.Map{"error": e.Error()})
	} else {
		logger.Trace("File transfer requested", "File %s, size %d requested to be transferred by operator", file.Filename, file.Size) // fmt.Sprintf("name: '%s', size: %d", file.Filename, file.Size))
	}

	return c.Status(http.StatusOK).JSON(file)
}
