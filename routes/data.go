package routes

import (
	"dd-opcda/db"
	"dd-opcda/types"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
)

func RegisterDataRoutes(api fiber.Router) {
	api.Get("/data/:type", GetAllOfType)
	api.Get("/data/:type/id/:id", GetDataByID)
	api.Get("/data/:type/field/:field/:value", GetDataByField)
	api.Post("/data/:type", NewData)
	api.Put("/data/:type", UpdateData)
	api.Delete("/data/:type/:id", DeleteDataByID)
	api.Delete("/data/:type/field/:field/:value", DeleteDataByField)
}

func GetAllOfType(c *fiber.Ctx) error {
	table := c.Params("type")
	items := types.CreateSlice(table)
	if items == nil {
		db.Log("error", "Failed to map provided data to type", table)
		return c.Status(503).SendString(table)
	}

	db.DB.Table(table).Preload(clause.Associations).Find(items)

	c.Status(200)
	return c.JSON(items)
}

func GetDataByID(c *fiber.Ctx) (err error) {
	id := c.Params("id")
	table := c.Params("type")
	item := types.CreateType(table)

	if err = db.DB.Take(item, id).Error; err != nil {
		db.Log("error", "Failed to find item", fmt.Sprintf("Item type '%s', id: %s, database error: %s", table, id, err.Error()))
		return c.Status(503).SendString(err.Error())
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"item": item})
}

func GetDataByField(c *fiber.Ctx) error {
	field := c.Params("field")
	value := c.Params("value")
	table := c.Params("type")
	items := types.CreateSlice(table)

	conditions := map[string]interface{}{field: value}
	if result := db.DB.Find(items, conditions); result.Error != nil {
		msg := db.Log("error", "Failed to find items", fmt.Sprintf("Item type '%s', field: %s, value: %s, database error: %s", table, field, value, result.Error))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"msg": msg})
	} else if result.RowsAffected == 0 {
		msg := db.Log("error", "No items found", fmt.Sprintf("Item type '%s', field: %s, value: %s", table, field, value))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"msg": msg})
	}

	return c.Status(http.StatusOK).JSON(items)
}

func NewData(c *fiber.Ctx) error {
	table := c.Params("type")
	item := types.CreateType(table)
	if err := c.BodyParser(&item); err != nil {
		db.Log("error", "Failed to map provided data to type", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	if err := db.DB.Create(item).Error; err != nil {
		msg := db.Log("error", "Failed to create item", fmt.Sprintf("Type: %s, data: %#v, error: %s", table, item, err.Error()))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"msg": msg})
	}

	db.Log("trace", "Item created", fmt.Sprintf("Type: %s, item: %#v", table, item))

	return c.Status(http.StatusOK).JSON(item)
}

func UpdateData(c *fiber.Ctx) error {
	table := c.Params("type")
	item := types.CreateType(table)
	if err := c.BodyParser(&item); err != nil {
		db.Log("error", "Failed to map provided data to type", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	db.DB.Save(item)
	db.Log("trace", "Item updated", fmt.Sprintf("Type: %s, item: %#v", table, item))

	c.Status(200)
	return c.JSON(item)
}

func DeleteDataByID(c *fiber.Ctx) error {
	id := c.Params("id")
	table := c.Params("type")
	item := types.CreateType(table)

	if err := db.DB.Delete(item, id).Error; err != nil {
		fmt.Printf("DELETE FAIL: %#v (%d)\n", item, id)
		db.Log("error", "Failed to delete item", err.Error())
		return c.Status(503).SendString(err.Error())
	}

	db.Log("trace", "Item deleted", fmt.Sprintf("Type: %s, ID: %s", table, id))

	c.Status(200)
	return c.JSON(item)
}

func DeleteDataByField(c *fiber.Ctx) error {
	field := c.Params("field")
	value, _ := url.QueryUnescape(c.Params("value"))
	table := c.Params("type")
	item := types.CreateSlice(table)

	conditions := map[string]interface{}{field: value}
	if result := db.DB.Delete(item, conditions); result.Error != nil {
		msg := db.Log("error", "Failed to delete item", fmt.Sprintf("Item type '%s', field: %s, value: %s, database error: %s", table, field, value, result.Error))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"msg": msg})
	} else if result.RowsAffected == 0 {
		msg := db.Log("error", "No items found", fmt.Sprintf("Item type '%s', field: %s, value: %s", table, field, value))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"msg": msg})
	}

	return c.Status(http.StatusOK).JSON(item)
}
