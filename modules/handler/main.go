package handler

import (
	"RediDB/modules/config"
	"RediDB/modules/structure"
	"log"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

var LengthOfID = 30
var App = fiber.New(fiber.Config{
	DisableStartupMessage: true,
	ReduceMemoryUsage:     false,
	UnescapePath:          true,
	JSONEncoder:           json.Marshal,
	JSONDecoder:           json.Unmarshal,
})

func init() {
	App.Use(func(ctx *fiber.Ctx) error {
		if ctx.Method() != "GET" {
			ctx.Request().Header.Set("Content-Type", "application/json")

			var auth structure.Auth
			if err := ctx.BodyParser(&auth); err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			if auth.Login != config.Get().Database.Login || auth.Password != config.Get().Database.Password {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "Authorization failed",
				})
			}
		}

		return ctx.Next()
	})

	App.Hooks().OnListen(func() error {
		println()
		log.Println("Served server on port " + strconv.Itoa(config.Get().Web.Port))
		return nil
	})

	handleSearch()
	handleCreate()
	handleSearchOrCreate()
	handleUpdate()
	handleDelete()
	HandleInfo()

	App.Get("*", func(ctx *fiber.Ctx) error {
		ctx.Context().SetStatusCode(400)
		return ctx.JSON(fiber.Map{
			"success": false,
			"message": "Bad Request",
		})
	})
}
