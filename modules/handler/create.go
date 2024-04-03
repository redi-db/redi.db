package handler

import (
	"RediDB/modules/distributor"
	"RediDB/modules/memcache"
	"RediDB/modules/path"
	"RediDB/modules/structure"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func handleCreate() {
	App.Post("/create", func(ctx *fiber.Ctx) error {
		var data struct {
			Database   string `json:"database"`
			Collection string `json:"collection"`

			DistributorID string                   `json:"distributorID"`
			Create        []map[string]interface{} `json:"data"`
		}

		if err := ctx.BodyParser(&data); err != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if len(data.DistributorID) > 0 {
			documents, size, err := distributor.GetData(data.DistributorID)
			if err != nil {
				ctx.Status(fiber.StatusNotFound)
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			return ctx.JSON(fiber.Map{
				"residue": size,
				"data":    documents,
			})
		}

		if len(data.Create) == 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": structure.NOTHING,
			})
		}

		path.Create()
		var created []map[string]interface{}
	MainLoop:
		for _, create := range data.Create {
			if len(create) == 0 {
				created = append(created, map[string]interface{}{
					"created": false,
					"skipped": true,
					"reason":  structure.EMPTY_DATA,
				})
				continue
			}

			for _, lockkey := range LockedFilters {
				if create[lockkey] != nil {
					created = append(created, map[string]interface{}{
						"created": false,
						"skipped": true,
						"reason":  fmt.Sprintf(structure.LOCK, lockkey),
					})
					continue MainLoop
				}
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
					"reason":  structure.ID_BANNED,
				})
			}
		}

		if len(created) > _config.Distribute.StartFrom {
			ctx.Status(fiber.StatusPartialContent)

			distributorID := distributor.Set(created)
			return ctx.JSON(fiber.Map{
				"distribute":    true,
				"distributorID": distributorID,
			})
		}

		return ctx.JSON(created)
	})
}

func WSHandleCreate(ws *websocket.Conn, request structure.WebsocketRequest) {
	distributorID := request.DistributorID
	if len(distributorID) > 0 {
		documents, size, err := distributor.GetData(distributorID)
		if err != nil {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:     true,
				RequestID: request.RequestID,
				Message:   err.Error(),
			})

			return
		}

		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID: request.RequestID,

			Residue: size,
			Data:    documents,
		})

		return
	}

	if len(request.Data.([]interface{})) == 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:     true,
			RequestID: request.RequestID,
			Message:   structure.NOTHING,
		})
		return
	}

	path.Create()
	var created []map[string]interface{}

MainLoop:
	for _, createData := range request.Data.([]interface{}) {
		if reflect.TypeOf(createData).String() != "map[string]interface {}" {
			created = append(created, map[string]interface{}{
				"created": false,
				"skipped": true,
				"reason":  structure.INVALID_STRUCTURE,
			})
			continue
		}

		create := createData.(map[string]interface{})

		if len(create) == 0 {
			created = append(created, map[string]interface{}{
				"created": false,
				"skipped": true,
				"reason":  structure.EMPTY_DATA,
			})
			continue
		}

		for _, lockkey := range LockedFilters {
			if create[lockkey] != nil {
				created = append(created, map[string]interface{}{
					"created": false,
					"skipped": true,
					"reason":  fmt.Sprintf(structure.LOCK, lockkey),
				})
				continue MainLoop
			}
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
				"reason":  structure.ID_BANNED,
			})
		}
	}

	if len(created) > _config.Distribute.StartFrom {
		distributorID = distributor.Set(created)
		ws.WriteJSON(structure.WebsocketAnswer{
			RequestID:     request.RequestID,
			DistributorID: distributorID,
		})

		return
	}

	ws.WriteJSON(structure.WebsocketAnswer{
		RequestID: request.RequestID,
		Data:      created,
	})
}

func generateID(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
