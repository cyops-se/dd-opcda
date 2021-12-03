package routes

import (
	"dd-opcda/db"
	"dd-opcda/engine"
	"dd-opcda/types"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cyops-se/opc"
	"github.com/gofiber/fiber/v2"
)

func RegisterOPCRoutes(api fiber.Router) {
	api.Get("/opc/server", GetAllServers)
	api.Get("/opc/server/:serverid", GetServerById)
	api.Get("/opc/server/:serverid/root", GetServerRoot)
	api.Get("/opc/server/:serverid/position", GetServerPosition)
	api.Get("/opc/server/:serverid/branches/:branch", GetServerBranches)
	api.Get("/opc/server/:serverid/leaves/:branch", GetServerLeaves)
	api.Get("/opc/server/:serverid/list/:branch", GetServerListBranches)

	api.Get("/opc/group", GetGroups)
	api.Post("/opc/group", NewGroup)
	api.Put("/opc/group", UpdateGroup)
	api.Get("/opc/group/:gid", GetGroup)
	api.Get("/opc/group/:gid/tags", GetGroupTags)
	api.Get("/opc/group/start/:gid", GetGroupStart)
	api.Get("/opc/group/stop/:gid", GetGroupStop)
	api.Post("/opc/group/:gid/tag", PostGroupAddTag)
	api.Delete("/opc/group/:gid/tag", PostGroupDelTag)

	api.Get("/opc/tag/names", GetTagNames)
	api.Post("/opc/tag/names", PostTagNames)
	api.Delete("/opc/tag/names", DeleteTagNames)

	api.Post("/opc/tag/changes", PostTagChanges)
}

func handlePanic(c *fiber.Ctx) {
	if r := recover(); r != nil {
		log.Println(r)
		// engine.Unlock()
		c.Status(500)
		c.JSON(r)
		return
	}
}

func handleError(c *fiber.Ctx, err error) error {
	return c.Status(400).JSON(err)
}

func GetAllServers(c *fiber.Ctx) error {
	return c.Status(200).JSON(engine.GetServers())
}

func GetServerById(c *fiber.Ctx) error {
	sid, _ := strconv.Atoi(c.Params("serverid"))
	server, err := engine.GetServer(sid)
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(200).JSON(server)
}

func GetServerRoot(c *fiber.Ctx) error {
	sid, _ := strconv.Atoi(c.Params("serverid"))
	browser, err := engine.GetBrowser(sid)
	if err != nil {
		return handleError(c, err)
	}

	engine.Lock()
	opc.MoveCursorHome(browser)
	branches := opc.CursorListBranches(browser)
	leaves := opc.CursorListLeaves(browser)
	position := fmt.Sprintf("root.%s", opc.CursorPosition(browser))
	server, _ := engine.GetServer(sid)
	engine.Unlock()

	c.Status(200)
	return c.JSON(&fiber.Map{
		"server":   server,
		"position": position,
		"branches": branches,
		"leaves":   leaves,
	})
}

func GetServerPosition(c *fiber.Ctx) error {
	sid, _ := strconv.Atoi(c.Params("serverid"))
	browser, err := engine.GetBrowser(sid)
	if err != nil {
		return handleError(c, err)
	}

	engine.Lock()
	position := opc.CursorPosition(browser)
	engine.Unlock()

	c.Status(200)
	return c.JSON(&fiber.Map{
		"position": position,
	})
}

func GetServerBranches(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	sid, _ := strconv.Atoi(c.Params("serverid"))
	encodedbranch := c.Params("branch")
	browser, err := engine.GetBrowser(sid)
	if err != nil {
		return handleError(c, err)
	}

	branch, err := url.QueryUnescape(encodedbranch)
	if err != nil {
		log.Println("Failed to decode branch", encodedbranch)
		return handleError(c, err)
	}

	engine.Lock()
	opc.MoveCursorTo(browser, branch)
	items := opc.CursorListBranches(browser)
	engine.Unlock()

	c.Status(200)
	return c.JSON(items)
}

func GetServerLeaves(c *fiber.Ctx) error {
	defer handlePanic(c)

	sid, _ := strconv.Atoi(c.Params("serverid"))
	encodedbranch := c.Params("branch")
	browser, err := engine.GetBrowser(sid)
	if err != nil {
		return handleError(c, err)
	}

	branch, err := url.QueryUnescape(encodedbranch)
	if err != nil {
		log.Println("Failed to decode branch", encodedbranch)
		return handleError(c, err)
	}

	engine.Lock()
	opc.MoveCursorTo(browser, branch)
	items := opc.CursorListLeaves(browser)
	engine.Unlock()

	c.Status(200)
	return c.JSON(items)
}

func GetServerListBranches(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	sid, _ := strconv.Atoi(c.Params("serverid"))
	encodedbranch := c.Params("branch")
	browser, err := engine.GetBrowser(sid)
	if err != nil {
		return handleError(c, err)
	}

	branch, err := url.QueryUnescape(encodedbranch)
	if err != nil {
		log.Println("Failed to decode branch", encodedbranch)
		return handleError(c, err)
	}

	engine.Lock()
	opc.MoveCursorTo(browser, branch)
	branches := opc.CursorListBranches(browser)
	leaves := opc.CursorListLeaves(browser)
	position := fmt.Sprintf("root.%s", opc.CursorPosition(browser))
	engine.Unlock()

	c.Status(200)
	return c.JSON(&fiber.Map{
		"position": position,
		"branches": branches,
		"leaves":   leaves,
	})
}

// GROUPS

func GetGroups(c *fiber.Ctx) (err error) {
	groups, _ := engine.GetGroups()

	// make sure at least one group is the default group
	found := false
	for _, group := range groups {
		if group.DefaultGroup {
			found = true
			break
		}
	}

	if !found && len(groups) > 0 {
		groups[0].DefaultGroup = true
		db.DB.Save(&groups[0])
	}

	return c.Status(http.StatusOK).JSON(groups)
}

func NewGroup(c *fiber.Ctx) (err error) {
	var group types.OPCGroup
	if err := c.BodyParser(&group); err != nil {
		db.Log("error", "UpdateGroup failed (bind)", fmt.Sprintf("%v", err))
		return c.Status(503).SendString(err.Error())
	}

	if group.DefaultGroup {
		if err := db.DB.Exec("update opc_groups set 'default_group' = false").Error; err != nil {
			db.Log("error", "UpdateGroup failed to reset default group flag", fmt.Sprintf("%v", err))
		}
	}

	if err := db.DB.Save(&group).Error; err != nil {
		db.Log("error", "UpdateGroup failed (save)", fmt.Sprintf("%v", err))
		return c.Status(503).SendString(err.Error())
	}

	// return c.Status(http.StatusOK).JSON(group)
	return GetGroups(c)
}

func UpdateGroup(c *fiber.Ctx) (err error) {
	var data types.OPCGroup
	if err := c.BodyParser(&data); err != nil {
		db.Log("error", "UpdateGroup failed (bind)", fmt.Sprintf("%v", err))
		return c.Status(503).SendString(err.Error())
	}

	var group types.OPCGroup
	if err := db.DB.First(&group, "ID = ?", data.ID).Error; err != nil {
		db.Log("error", "UpdateGroup failed (first)", fmt.Sprintf("%v", err))
		return c.Status(503).SendString(err.Error())
	}

	if data.DefaultGroup {
		if err := db.DB.Exec("update opc_groups set 'default_group' = false").Error; err != nil {
			db.Log("error", "UpdateGroup failed to reset default group flag", fmt.Sprintf("%v", err))
		}
	}

	group.Name = data.Name
	group.Description = data.Description
	group.DefaultGroup = data.DefaultGroup
	group.DiodeProxyID = data.DiodeProxy.ID
	group.Interval = data.Interval
	group.RunAtStart = data.RunAtStart

	if err := db.DB.Save(&group).Error; err != nil {
		db.Log("error", "UpdateGroup failed (save)", fmt.Sprintf("%v", err))
		return c.Status(503).SendString(err.Error())
	}

	// return c.Status(http.StatusOK).JSON(group)
	return GetGroups(c)
}

func GetGroup(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	gid, _ := strconv.Atoi(c.Params("gid"))
	group, err := engine.GetGroup(uint(gid))
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(200).JSON(group)
}

func GetGroupTags(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	gid, _ := strconv.Atoi(c.Params("gid"))
	tags, err := engine.GetGroupTags(uint(gid))
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(200).JSON(tags)
}

func GetGroupStart(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	gid, _ := strconv.Atoi(c.Params("gid"))
	group, err := engine.GetGroup(uint(gid))
	if err != nil {
		return handleError(c, err)
	}

	if err = engine.Start(group); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(&fiber.Map{"group": group, "error": err.Error()})
	}

	return c.Status(200).JSON(&fiber.Map{"group": group})
}

func GetGroupStop(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	gid, _ := strconv.Atoi(c.Params("gid"))
	group, err := engine.GetGroup(uint(gid))
	if err != nil {
		return handleError(c, err)
	}

	if err = engine.Stop(group); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(&fiber.Map{"group": group, "error": err.Error()})
	}

	return c.Status(200).JSON(&fiber.Map{"group": group})
}

func PostGroupAddTag(c *fiber.Ctx) (err error) {
	gid, _ := strconv.Atoi(c.Params("gid"))
	group, err := engine.GetGroup(uint(gid))
	if err != nil {
		return handleError(c, err)
	}

	var items []string
	if err = c.BodyParser(&items); err == nil {
		for _, tagname := range items {
			var tag types.OPCTag
			err = db.DB.Take(&tag, "name = ?", tagname).Error
			if err == nil {
				tag.GroupID = group.ID
				db.DB.Save(tag)
			}
		}

		return c.Status(200).JSON(&fiber.Map{"group": group})
	}

	return
}

func PostGroupDelTag(c *fiber.Ctx) (err error) {
	return nil
}

func GetTagNames(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	tags, err := engine.GetTagNames()
	if err != nil {
		return handleError(c, err)
	}

	return c.Status(200).JSON(tags)
}

func PostTagNames(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	tagnames, err := engine.GetTagNames()
	if err != nil {
		return handleError(c, err)
	}

	defaultgroup, err := engine.GetDefaultGroup()
	if err != nil {
		return handleError(c, err)
	}

	var items []string
	savedcount := 0
	failedcount := 0
	skippedcount := 0
	if err = c.BodyParser(&items); err == nil {
		for _, tagname := range items {
			found := false
			for _, name := range tagnames {
				if tagname == name {
					found = true
					skippedcount++
					break
				}
			}

			if !found {
				tag := &types.OPCTag{Name: tagname, GroupID: defaultgroup.ID}
				if err = db.DB.Create(&tag).Error; err == nil {
					savedcount++
				} else {
					failedcount++
				}
			}
		}
	}

	return c.Status(200).JSON(&fiber.Map{"saved": savedcount, "failed": failedcount, "skipped": skippedcount, "total": len(items)})
}

func DeleteTagNames(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	var items []string
	if err = c.BodyParser(&items); err == nil {
		if err = db.DB.Where("Name in ?", items).Delete(&types.OPCTag{}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(&fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(200).JSON(&fiber.Map{"deleteitems": items})
}

func PostTagChanges(c *fiber.Ctx) (err error) {
	defer handlePanic(c)

	tagnames, err := engine.GetTagNames()
	if err != nil {
		return handleError(c, err)
	}

	defaultgroup, err := engine.GetDefaultGroup()
	if err != nil {
		return handleError(c, err)
	}

	var items []types.OPCTag
	newcount := 0
	failedcount := 0
	updatedcount := 0
	if err = c.BodyParser(&items); err == nil {
		for _, item := range items {
			found := false
			for _, name := range tagnames {
				if item.Name == name {
					if err = db.DB.Save(&item).Error; err == nil {
						found = true
						updatedcount++
						break
					}
				}
			}

			if !found {
				item.GroupID = defaultgroup.ID
				if err = db.DB.Create(&item).Error; err == nil {
					newcount++
				} else {
					failedcount++
				}
			}
		}
	}

	return c.Status(200).JSON(&fiber.Map{"new": newcount, "updated": updatedcount, "failed": failedcount, "total": len(items)})
}
