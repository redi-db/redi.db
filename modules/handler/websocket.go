package handler

import (
	"RediDB/modules/config"
	"RediDB/modules/structure"
	"reflect"

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
				if request.RequestID < 1 {
					ws.WriteJSON(structure.WebsocketAnswer{
						Error:   true,
						Message: structure.INVALID_REQUEST_ID,
					})
					return
				}

				if len(request.Database) == 0 {
					ws.WriteJSON(structure.WebsocketAnswer{
						Error:     true,
						RequestID: request.RequestID,
						Message:   structure.INVALID_DATABASE,
					})
					return
				}

				if len(request.Collection) == 0 {
					ws.WriteJSON(structure.WebsocketAnswer{
						Error:     true,
						RequestID: request.RequestID,
						Message:   structure.INVALID_COLLECTION,
					})
					return
				}

				switch request.Type {
				case "create":
					if reflect.TypeOf(request.Data).String() != "[]interface {}" {
						ws.WriteJSON(structure.WebsocketAnswer{
							Error:     true,
							RequestID: request.RequestID,
							Message:   structure.METHOD_NOT_ALLOWED,
						})
						return
					}

					WSHandleCreate(ws, request)

				case "delete":
					WSHandleDelete(ws, request)

				case "update":
					if reflect.TypeOf(request.Data).String() != "[]interface {}" {
						ws.WriteJSON(structure.WebsocketAnswer{
							Error:     true,
							RequestID: request.RequestID,
							Message:   structure.METHOD_NOT_ALLOWED,
						})
						return
					}

					WSHandleUpdate(ws, request)

				case "instantUpdate":
					if reflect.TypeOf(request.Data).String() != "[]interface {}" {
						ws.WriteJSON(structure.WebsocketAnswer{
							Error:     true,
							RequestID: request.RequestID,
							Message:   structure.METHOD_NOT_ALLOWED,
						})
						return
					}

					WSHandleInstantUpdate(ws, request)

				case "search":
					WSHandleSearch(ws, request)

				case "searchOrCreate":
					if reflect.TypeOf(request.Data).String() != "[]interface {}" {
						ws.WriteJSON(structure.WebsocketAnswer{
							Error:     true,
							RequestID: request.RequestID,
							Message:   structure.METHOD_NOT_ALLOWED,
						})
						return
					}

					WSHandleSearchOrCreate(ws, request)

				default:
					ws.WriteJSON(structure.WebsocketAnswer{
						Error:     true,
						RequestID: request.RequestID,
						Message:   "Invalid request type",
					})
				}
			}
		}
	}))
}
