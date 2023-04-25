package handler

import (
	"RediDB/modules/config"
	"RediDB/modules/structure"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func HandleWS() {
	config := config.Get()
	App.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			if c.Query("login") != config.Database.Login || c.Query("password") != config.Database.Password {
				return fiber.ErrUnauthorized
			}

			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	App.Get("/ws", websocket.New(func(ws *websocket.Conn) {
		var (
			msg []byte
			err error
		)

		for {
			if _, msg, err = ws.ReadMessage(); err != nil {
				// Error on reading message, disconnect client
				break
			}

			var request structure.WebsocketRequest
			if err := json.Unmarshal(msg, &request); err != nil {
				ws.WriteJSON(structure.WebsocketAnswer{
					Error:   true,
					Message: err.Error(),
				})
			} else {
				if len(request.Database) == 0 {
					ws.WriteJSON(structure.WebsocketAnswer{
						Error:   true,
						Message: "<database> cannot be empty",
					})
				}

				if len(request.Collection) == 0 {
					ws.WriteJSON(structure.WebsocketAnswer{
						Error:   true,
						Message: "<collection> cannot be empty",
					})
				}

				switch request.Type {
				case "create":
					WSHandleCreate(ws, request)

				case "delete":
					WSHandleDelete(ws, request)

				case "update":
					WSHandleUpdate(ws, request)

				case "instantUpdate":
					WSHandleInstantUpdate(ws, request)

				case "search":
					WSHandleSearch(ws, request)

				case "searchOrCreate":
					WSHandleSearchOrCreate(ws, request)

				default:
					ws.WriteJSON(structure.WebsocketAnswer{
						Error:   true,
						Message: "Invalid request type",
					})
				}
			}
		}
	}))
}
