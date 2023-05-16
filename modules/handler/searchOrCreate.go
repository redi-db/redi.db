package handler

import (
	"RediDB/modules/memcache"
	"RediDB/modules/structure"
	"fmt"
	"os"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func handleSearchOrCreate() {
	App.Post("/searchOrCreate", func(ctx *fiber.Ctx) error {
		var data struct {
			Database   string `json:"database"`
			Collection string `json:"collection"`

			Filter     map[string]interface{} `json:"filter"`
			CreateData map[string]interface{} `json:"data"`
		}

		if err := ctx.BodyParser(&data); err != nil {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		if data.Filter == nil || len(data.Filter) == 0 || data.CreateData == nil || len(data.CreateData) == 0 {
			return ctx.JSON(fiber.Map{
				"success": false,
				"message": structure.EMPTY_DATA,
			})
		}

		if data.CreateData != nil {
			delete(data.CreateData, "$or")
			delete(data.CreateData, "$order")
			delete(data.CreateData, "$max")
			delete(data.CreateData, "$only")
			delete(data.CreateData, "$omit")
		}

		if data.Filter != nil {
			delete(data.Filter, "$order")
			delete(data.Filter, "$max")
			delete(data.Filter, "$only")
			delete(data.Filter, "$omit")
		}

		filter, err := handleHttpFilter(data.Filter)
		if len(err) > 0 {
			return ctx.JSON(err)
		}

		found := memcache.Get(data.Database, data.Collection, filter, 1)
		result := map[string]interface{}{
			"created": false,
		}

		if found == nil {
			result["created"] = true
			id := generateID(LengthOfID)
			document := data.CreateData
			document["_id"] = id

			if _, err := os.Stat(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection)); os.IsNotExist(err) {
				err := os.MkdirAll(fmt.Sprintf("./data/%s/%s", data.Database, data.Collection), os.ModePerm)
				if err != nil {
					return ctx.JSON(fiber.Map{
						"success": false,
						"message": err.Error(),
					})
				}
			}

			file, err := os.Create(fmt.Sprintf("./data/%s/%s/%s.db", data.Database, data.Collection, id))
			if err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			encoded, err := json.Marshal(document)
			if err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			if _, err := file.WriteString(string(encoded)); err != nil {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": err.Error(),
				})
			}

			_ = file.Close()

			memcache.Cache.Lock()
			memcache.CacheSet(data.Database, data.Collection, id, document)
			memcache.Cache.Unlock()
			result["data"] = document

			return ctx.JSON(result)
		}

		result["data"] = found[0]
		return ctx.JSON(result)
	})
}

func WSHandleSearchOrCreate(ws *websocket.Conn, request structure.WebsocketRequest) {
	if request.Filter == nil || len(request.Filter) == 0 || request.Data == nil || len(request.Data.([]interface{})) == 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: structure.EMPTY_DATA,
		})
		return
	}

	if reflect.TypeOf(request.Data).String() != "[]interface {}" || request.Data.([]interface{})[0] == nil || reflect.TypeOf(request.Data.([]interface{})[0]).String() != "map[string]interface {}" {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: structure.INVALID_STRUCTURE,
		})

		return
	}

	createData := request.Data.([]interface{})[0].(map[string]interface{})
	if len(createData) == 0 {
		ws.WriteJSON(structure.WebsocketAnswer{
			Error:   true,
			Message: structure.INVALID_STRUCTURE,
		})

		return
	}

	if request.Data.([]interface{})[0] != nil {
		delete(createData, "$or")
		delete(createData, "$order")
		delete(createData, "$max")
		delete(createData, "$only")
		delete(createData, "$omit")
	}

	filter, err := handleWSFilter(request.Filter)
	if err.Error {
		ws.WriteJSON(err)
		return
	}

	if filter != nil {
		delete(filter, "$order")
		delete(filter, "$max")
	}

	found := memcache.Get(request.Database, request.Collection, filter, 1)
	result := map[string]interface{}{
		"created": false,
	}

	if found == nil {
		result["created"] = true
		id := generateID(LengthOfID)
		document := request.Data.([]interface{})[0].(map[string]interface{})
		document["_id"] = id

		if _, err := os.Stat(fmt.Sprintf("./data/%s/%s", request.Database, request.Collection)); os.IsNotExist(err) {
			err := os.MkdirAll(fmt.Sprintf("./data/%s/%s", request.Database, request.Collection), os.ModePerm)
			if err != nil {
				ws.WriteJSON(structure.WebsocketAnswer{
					Error:   true,
					Message: err.Error(),
				})
				return
			}
		}

		file, err := os.Create(fmt.Sprintf("./data/%s/%s/%s.db", request.Database, request.Collection, id))
		if err != nil {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: err.Error(),
			})
			return
		}

		encoded, err := json.Marshal(document)
		if err != nil {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: err.Error(),
			})
			return
		}

		if _, err := file.WriteString(string(encoded)); err != nil {
			ws.WriteJSON(structure.WebsocketAnswer{
				Error:   true,
				Message: err.Error(),
			})
			return
		}

		_ = file.Close()

		memcache.Cache.Lock()
		memcache.CacheSet(request.Database, request.Collection, id, document)
		memcache.Cache.Unlock()
		result["data"] = document

		ws.WriteJSON(structure.WebsocketAnswer{
			Data: result,
		})

		return
	}

	result["data"] = found[0]
	ws.WriteJSON(structure.WebsocketAnswer{
		Data: result,
	})
}
