package handler

import (
	"RediDB/modules/memcache"
	"RediDB/modules/path"
	"RediDB/modules/structure"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func handleCreate() {
	App.Post("/create", func(ctx *fiber.Ctx) error {
		var data struct {
			Database   string `json:"database"`
			Collection string `json:"collection"`

			Create []map[string]interface{} `json:"data"`
		}

		if err := ctx.BodyParser(&data); err != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if len(data.Create) == 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": "Nothing to create",
			})
		}

		path.Create()
		var created []map[string]interface{}
		for _, create := range data.Create {
			if len(create) == 0 {
				created = append(created, map[string]interface{}{
					"created": false,
					"skipped": true,
					"reason":  "Object is empty",
				})
				continue
			}

			if create["_id"] != nil {
				created = append(created, map[string]interface{}{
					"created": false,
					"skipped": true,
					"reason":  "ID is locked property",
				})
				continue
			}

			if create["$order"] != nil {
				created = append(created, map[string]interface{}{
					"created": false,
					"skipped": true,
					"reason":  "$order is locked property",
				})
				continue
			}

			if create["$max"] != nil {
				created = append(created, map[string]interface{}{
					"created": false,
					"skipped": true,
					"reason":  "$max is locked property",
				})
				continue
			}

			if create["$or"] != nil {
				created = append(created, map[string]interface{}{
					"created": false,
					"skipped": true,
					"reason":  "$or is locked property",
				})
				continue
			}

			id := generateID(LengthOfID)
			create["_id"] = id

			if _, err := os.Stat(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection)); os.IsNotExist(err) {
				err := os.MkdirAll(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection), os.ModePerm)
				if err != nil {
					created = append(created, map[string]interface{}{
						"_id":     id,
						"created": false,
						"reason":  err.Error(),
					})
					continue
				}
			}

			if _, err := os.Stat(fmt.Sprintf("./data/%s/%s/%s.db", data.Database, data.Collection, id)); os.IsNotExist(err) {
				file, err := os.Create(fmt.Sprintf("./data/%s/%s/%s.db", data.Database, data.Collection, id))
				if err != nil {
					created = append(created, map[string]interface{}{
						"_id":     id,
						"created": false,
						"reason":  err.Error(),
					})
					continue
				}

				encoded, err := json.Marshal(create)
				if err != nil {
					created = append(created, map[string]interface{}{
						"_id":     id,
						"created": false,
						"reason":  err.Error(),
					})
					continue
				}

				if _, err := file.WriteString(string(encoded)); err != nil {
					created = append(created, map[string]interface{}{
						"_id":     id,
						"created": false,
						"reason":  err.Error(),
					})
					continue
				}

				_ = file.Close()

				memcache.Cache.Lock()
				memcache.CacheSet(data.Database, data.Collection, id, create)
				created = append(created, map[string]interface{}{
					"_id":     id,
					"created": true,
				})

				memcache.Cache.Unlock()
			} else {
				created = append(created, map[string]interface{}{
					"_id":     id,
					"created": false,
					"reason":  "This id is already in usage",
				})
			}
		}

		return ctx.JSON(created)
	})
}

func WSHandleCreate(ws *websocket.Conn, request structure.WebsocketRequest) {
	if len(request.Data.([]interface{})) == 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: "Nothing to create",
		})
		return
	}

	path.Create()
	var created []map[string]interface{}
	for _, createData := range request.Data.([]interface{}) {
		create := createData.(map[string]interface{})

		if len(create) == 0 {
			created = append(created, map[string]interface{}{
				"created": false,
				"skipped": true,
				"reason":  "Object is empty",
			})
			continue
		}

		if create["_id"] != nil {
			created = append(created, map[string]interface{}{
				"created": false,
				"skipped": true,
				"reason":  "ID is locked property",
			})
			continue
		}

		if create["$order"] != nil {
			created = append(created, map[string]interface{}{
				"created": false,
				"skipped": true,
				"reason":  "$order is locked property",
			})
			continue
		}

		if create["$max"] != nil {
			created = append(created, map[string]interface{}{
				"created": false,
				"skipped": true,
				"reason":  "$max is locked property",
			})
			continue
		}

		if create["$or"] != nil {
			created = append(created, map[string]interface{}{
				"created": false,
				"skipped": true,
				"reason":  "$or is locked property",
			})
			continue
		}

		id := generateID(LengthOfID)
		create["_id"] = id

		if _, err := os.Stat(fmt.Sprintf("./data/%s/%s", request.Database, request.Collection)); os.IsNotExist(err) {
			err := os.MkdirAll(fmt.Sprintf("./data/%s/%s", request.Database, request.Collection), os.ModePerm)
			if err != nil {
				created = append(created, map[string]interface{}{
					"_id":     id,
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}
		}

		if _, err := os.Stat(fmt.Sprintf("./data/%s/%s/%s.db", request.Database, request.Collection, id)); os.IsNotExist(err) {
			file, err := os.Create(fmt.Sprintf("./data/%s/%s/%s.db", request.Database, request.Collection, id))
			if err != nil {
				created = append(created, map[string]interface{}{
					"_id":     id,
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}

			encoded, err := json.Marshal(create)
			if err != nil {
				created = append(created, map[string]interface{}{
					"_id":     id,
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}

			if _, err := file.WriteString(string(encoded)); err != nil {
				created = append(created, map[string]interface{}{
					"_id":     id,
					"created": false,
					"reason":  err.Error(),
				})
				continue
			}

			_ = file.Close()

			memcache.Cache.Lock()
			memcache.CacheSet(request.Database, request.Collection, id, create)
			created = append(created, map[string]interface{}{
				"_id":     id,
				"created": true,
			})

			memcache.Cache.Unlock()
		} else {
			created = append(created, map[string]interface{}{
				"_id":     id,
				"created": false,
				"reason":  "This id is already in usage",
			})
		}
	}

	ws.WriteJSON(structure.WebsocketAnswer{
		Data: created,
	})
}

func generateID(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
