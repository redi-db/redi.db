package handler

import (
	"RediDB/modules/config"
	"RediDB/modules/structure"
	"RediDB/modules/updates"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

var LengthOfID = 18
var IgnoreFilters = []string{"$only", "$order", "$max", "$omit"}
var LockedFilters = []string{"_id", "$or", "$order", "$gt", "$arr", "$lt", "$max", "$omit", "$only", "$regex", "$and"}

var App = fiber.New(fiber.Config{
	DisableStartupMessage: true,
	BodyLimit:             config.Get().Settings.MaxData * 1024 * 1024,
	ReduceMemoryUsage:     false,
	UnescapePath:          true,
	JSONEncoder:           json.Marshal,
	JSONDecoder:           json.Unmarshal,
})

var _config = config.Get()

func init() {
	if _config.Web.WebSocketAllowed {
		HandleWS()
	}

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

			if auth.Login != _config.Database.Login || auth.Password != _config.Database.Password {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "Authorization failed",
				})
			}

			if len(auth.Database) == 0 {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "Database required field",
				})
			} else if len(auth.Collection) == 0 {
				return ctx.JSON(fiber.Map{
					"success": false,
					"message": "Collection required field",
				})
			}
		}

		return ctx.Next()
	})

	App.Hooks().OnListen(func() error {
		println()
		log.Println("Served server on port " + strconv.Itoa(_config.Web.Port))

		if _config.Settings.CheckUpdates {
			version, updateRequired, err := updates.Check()
			if err != nil {
				log.Printf("Failed to check updates: %s", err.Error())
				return nil
			}

			if updateRequired {
				log.Printf("New version is available: v%s (Current v%s)", version, updates.VERSION)
			}
		}

		if _config.Garbage.Enabled {
			ticker := time.NewTicker(time.Duration(_config.Garbage.Interval) * time.Minute)
			go func() {
				for range ticker.C {
					runtime.GC()
				}
			}()
		}

		return nil
	})

	handleSearch()
	handleCreate()
	handleSearchOrCreate()
	handleInstantUpdate()
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

func handleHttpFilter(filter map[string]interface{}) (map[string]interface{}, fiber.Map) {
	if filter == nil {
		return nil, nil
	}

	if filter["$max"] == nil {
		filter["$max"] = 0.0
	}

	if reflect.TypeOf(filter["$max"]).String() != "float64" && reflect.TypeOf(filter["$max"]).String() != "int" {
		return nil, fiber.Map{
			"success": false,
			"message": fmt.Sprintf(structure.MUST_BY, "$max", "integer"),
		}
	}

	if filter["$only"] != nil {
		if reflect.TypeOf(filter["$only"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$only", "array with strings"),
			}
		}

		for i, onlyValue := range filter["$only"].([]interface{}) {
			if onlyValue == nil || reflect.TypeOf(onlyValue).String() != "string" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$only with index %d", i), "string"),
				}
			}
		}
	}

	if filter["$arr"] != nil {
		if reflect.TypeOf(filter["$arr"]).String() != "map[string]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$arr", "object with \"by\" and \"type\""),
			}
		}

		arr := filter["$arr"].(map[string]interface{})
		if arr["by"] == nil || reflect.TypeOf(arr["by"]).String() != "string" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$arr", "object with \"by\" (string)"),
			}
		} else {
			value := arr["value"]
			if value == nil || reflect.TypeOf(value).String() != "map[string]interface {}" && reflect.TypeOf(value).String() != "string" && reflect.TypeOf(value).String() != "float64" && reflect.TypeOf(value).String() != "bool" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, "$arr", "object with \"value\" (boolean, number, string or object)"),
				}
			}
		}
	}

	if filter["$omit"] != nil {
		if reflect.TypeOf(filter["$omit"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$omit", "array with strings"),
			}
		}

		for i, omitValue := range filter["$omit"].([]interface{}) {
			if omitValue == nil || reflect.TypeOf(omitValue).String() != "string" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$omit with index %d", i), "string"),
				}
			}
		}
	}

	if filter["$order"] != nil {
		if reflect.TypeOf(filter["$order"]).String() != "map[string]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$order", "\"type\" and \"by\""),
			}
		}

		orderType, orderTypeOk := filter["$order"].(map[string]interface{})["type"]
		orderBy, orderByOk := filter["$order"].(map[string]interface{})["by"]

		if !orderTypeOk {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"type\""),
			}
		}

		if orderType != "desc" && orderType != "asc" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.REQUIRED_INVALID, "$order \"type\"", "\"desc\" and \"asc\""),
			}
		}

		if !orderByOk {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"by\""),
			}
		}

		if reflect.TypeOf(orderBy).String() != "string" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "\"by\"", "string"),
			}
		}
	}

	if filter["$or"] != nil {
		if reflect.TypeOf(filter["$or"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$or", "array"),
			}
		}

		if len(filter["$or"].([]interface{})) == 0 {
			return nil, fiber.Map{
				"success": false,
				"message": structure.EMPTY_DATA,
			}
		}

		for i, or := range filter["$or"].([]interface{}) {
			if or == nil || reflect.TypeOf(or).String() != "map[string]interface {}" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$or with index %d", i), "object"),
				}
			}

			ok, msg := handleHttpFilter(or.(map[string]interface{}))
			if ok == nil {
				return nil, msg
			}
		}
	}

	if filter["$and"] != nil {
		if reflect.TypeOf(filter["$and"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$and", "array"),
			}
		}

		if len(filter["$and"].([]interface{})) == 0 {
			return nil, fiber.Map{
				"success": false,
				"message": structure.EMPTY_DATA,
			}
		}

		for i, and := range filter["$and"].([]interface{}) {
			if and == nil || reflect.TypeOf(and).String() != "map[string]interface {}" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$and with index %d", i), "object"),
				}
			}

			ok, msg := handleHttpFilter(and.(map[string]interface{}))
			if ok == nil {
				return nil, msg
			}
		}
	}

	if filter["$regex"] != nil {
		if reflect.TypeOf(filter["$regex"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$regex", "array"),
			}
		}

		for i, rValue := range filter["$regex"].([]interface{}) {
			if rValue == nil || reflect.TypeOf(rValue).String() != "map[string]interface {}" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$regex with index %d", i), "object"),
				}
			}

			r := rValue.(map[string]interface{})
			if r["by"] == nil || reflect.TypeOf(r["by"]).String() != "string" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$regex with index %d", i), "object with \"by\" (string)"),
				}
			} else if r["value"] == nil || reflect.TypeOf(r["value"]).String() != "string" || r["value"].(string) == "" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$regex with index %d", i), "object with \"value\" (string)"),
				}
			}
		}
	}

	if filter["$ne"] != nil {
		if reflect.TypeOf(filter["$ne"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$ne", "array"),
			}
		}

		for i, neValue := range filter["$ne"].([]interface{}) {
			if neValue == nil || reflect.TypeOf(neValue).String() != "map[string]interface {}" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$ne with index %d", i), "object"),
				}
			}

			ne := neValue.(map[string]interface{})
			if ne["by"] == nil || reflect.TypeOf(ne["by"]).String() != "string" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$ne with index %d", i), "object with \"by\" (string)"),
				}
			}
		}
	}

	if filter["$gt"] != nil {
		if reflect.TypeOf(filter["$gt"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$gt", "array"),
			}
		}

		for i, gtValue := range filter["$gt"].([]interface{}) {
			if gtValue == nil || reflect.TypeOf(gtValue).String() != "map[string]interface {}" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$gt with index %d", i), "object"),
				}
			}

			gt := gtValue.(map[string]interface{})
			if gt["by"] == nil || reflect.TypeOf(gt["by"]).String() != "string" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$gt with index %d", i), "object with \"by\" (string)"),
				}
			} else if gt["value"] == nil || reflect.TypeOf(gt["value"]).String() != "float64" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$gt with index %d", i), "object with \"value\" (number)"),
				}
			}
		}
	}

	if filter["$lt"] != nil {
		if reflect.TypeOf(filter["$lt"]).String() != "[]interface {}" {
			return nil, fiber.Map{
				"success": false,
				"message": fmt.Sprintf(structure.MUST_BY, "$lt", "array"),
			}
		}

		for i, ltValue := range filter["$lt"].([]interface{}) {
			if ltValue == nil || reflect.TypeOf(ltValue).String() != "map[string]interface {}" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$lt with index %d", i), "object"),
				}
			}

			lt := ltValue.(map[string]interface{})
			if lt["by"] == nil || reflect.TypeOf(lt["by"]).String() != "string" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$lt with index %d", i), "object with \"by\" (string)"),
				}
			} else if lt["value"] == nil || reflect.TypeOf(lt["value"]).String() != "float64" {
				return nil, fiber.Map{
					"success": false,
					"message": fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$lt with index %d", i), "object with \"value\" (number)"),
				}
			}
		}
	}

	max := filter["$max"].(float64)
	if int(max) < 0 {
		return nil, fiber.Map{
			"success": false,
			"message": fmt.Sprintf(structure.MUST_BY, "$max", ">= 0"),
		}
	}

	filter["$max"] = int(max)
	return filter, nil
}

func handleWSFilter(filter map[string]interface{}, requestID int) (map[string]interface{}, structure.WebsocketAnswer) {
	if filter["$max"] == nil {
		filter["$max"] = 0.0
	}

	if reflect.TypeOf(filter["$max"]).String() != "float64" && reflect.TypeOf(filter["$max"]).String() != "int" {
		return nil, structure.WebsocketAnswer{
			Error:     true,
			RequestID: requestID,
			Message:   fmt.Sprintf(structure.MUST_BY, "$max", "integer"),
		}
	}

	if filter["$only"] != nil {
		if reflect.TypeOf(filter["$only"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$only", "array with strings"),
			}
		}

		for i, onlyValue := range filter["$only"].([]interface{}) {
			if onlyValue == nil || reflect.TypeOf(onlyValue).String() != "string" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$only with index %d", i), "string"),
				}
			}
		}
	}

	if filter["$arr"] != nil {
		if reflect.TypeOf(filter["$arr"]).String() != "map[string]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$arr", "object with \"by\" and \"type\""),
			}
		}

		arr := filter["$arr"].(map[string]interface{})
		if arr["by"] == nil || reflect.TypeOf(arr["by"]).String() != "string" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$arr", "object with \"by\" (string)"),
			}
		} else {
			value := arr["value"]
			if value == nil || reflect.TypeOf(value).String() != "map[string]interface {}" && reflect.TypeOf(value).String() != "string" && reflect.TypeOf(value).String() != "float64" && reflect.TypeOf(value).String() != "bool" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, "$arr", "object with \"value\" (boolean, number, string or object)"),
				}
			}
		}
	}

	if filter["$omit"] != nil {
		if reflect.TypeOf(filter["$omit"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$omit", "array with strings"),
			}
		}

		for i, omitValue := range filter["$omit"].([]interface{}) {
			if omitValue == nil || reflect.TypeOf(omitValue).String() != "string" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$omit with index %d", i), "string"),
				}
			}
		}
	}

	if filter["$order"] != nil {
		if reflect.TypeOf(filter["$order"]).String() != "map[string]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$order", "\"type\" and \"by\""),
			}
		}

		orderType, orderTypeOk := filter["$order"].(map[string]interface{})["type"]
		orderBy, orderByOk := filter["$order"].(map[string]interface{})["by"]

		if !orderTypeOk {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"type\""),
			}
		}

		if orderType != "desc" && orderType != "asc" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.REQUIRED_INVALID, "$order \"type\"", "\"desc\" and \"asc\""),
			}
		}

		if !orderByOk {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.REQUIRED_FIELD, "$order \"by\""),
			}
		}

		if reflect.TypeOf(orderBy).String() != "string" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "\"by\"", "string"),
			}
		}
	}

	if filter["$or"] != nil {
		if reflect.TypeOf(filter["$or"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$or", "array"),
			}
		}

		if len(filter["$or"].([]interface{})) == 0 {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   structure.EMPTY_DATA,
			}
		}

		for i, or := range filter["$or"].([]interface{}) {
			if or == nil || reflect.TypeOf(or).String() != "map[string]interface {}" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$or with index %d", i), "object"),
				}
			}

			ok, msg := handleWSFilter(or.(map[string]interface{}), requestID)
			if ok == nil {
				return nil, msg
			}
		}
	}

	if filter["$and"] != nil {
		if reflect.TypeOf(filter["$and"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$and", "array"),
			}
		}

		if len(filter["$and"].([]interface{})) == 0 {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   structure.EMPTY_DATA,
			}
		}

		for i, and := range filter["$and"].([]interface{}) {
			if and == nil || reflect.TypeOf(and).String() != "map[string]interface {}" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$and with index %d", i), "object"),
				}
			}

			ok, msg := handleWSFilter(and.(map[string]interface{}), requestID)
			if ok == nil {
				return nil, msg
			}
		}
	}

	if filter["$regex"] != nil {
		if reflect.TypeOf(filter["$regex"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$regex", "array"),
			}
		}

		for i, rValue := range filter["$regex"].([]interface{}) {
			if rValue == nil || reflect.TypeOf(rValue).String() != "map[string]interface {}" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$regex with index %d", i), "object"),
				}
			}

			r := rValue.(map[string]interface{})
			if r["by"] == nil || reflect.TypeOf(r["by"]).String() != "string" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$regex with index %d", i), "object with \"by\" (string)"),
				}
			} else if r["value"] == nil || reflect.TypeOf(r["value"]).String() != "string" || r["value"].(string) == "" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$regex with index %d", i), "object with \"value\" (string)"),
				}
			}
		}
	}

	if filter["$ne"] != nil {
		if reflect.TypeOf(filter["$ne"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$ne", "array"),
			}
		}

		for i, neValue := range filter["$ne"].([]interface{}) {
			if neValue == nil || reflect.TypeOf(neValue).String() != "map[string]interface {}" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$ne with index %d", i), "object"),
				}
			}

			ne := neValue.(map[string]interface{})
			if ne["by"] == nil || reflect.TypeOf(ne["by"]).String() != "string" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$ne with index %d", i), "object with \"by\" (string)"),
				}
			}
		}
	}

	if filter["$gt"] != nil {
		if reflect.TypeOf(filter["$gt"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$gt", "array"),
			}
		}

		for i, gtValue := range filter["$gt"].([]interface{}) {
			if gtValue == nil || reflect.TypeOf(gtValue).String() != "map[string]interface {}" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$gt with index %d", i), "object"),
				}
			}

			gt := gtValue.(map[string]interface{})
			if gt["by"] == nil || reflect.TypeOf(gt["by"]).String() != "string" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$gt with index %d", i), "object with \"by\" (string)"),
				}
			} else if gt["value"] == nil || reflect.TypeOf(gt["value"]).String() != "float64" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$gt with index %d", i), "object with \"value\" (number)"),
				}
			}
		}
	}

	if filter["$lt"] != nil {
		if reflect.TypeOf(filter["$lt"]).String() != "[]interface {}" {
			return nil, structure.WebsocketAnswer{
				Error:     true,
				RequestID: requestID,
				Message:   fmt.Sprintf(structure.MUST_BY, "$lt", "array"),
			}
		}

		for i, ltValue := range filter["$lt"].([]interface{}) {
			if ltValue == nil || reflect.TypeOf(ltValue).String() != "map[string]interface {}" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$lt with index %d", i), "object"),
				}
			}

			lt := ltValue.(map[string]interface{})
			if lt["by"] == nil || reflect.TypeOf(lt["by"]).String() != "string" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$lt with index %d", i), "object with \"by\" (string)"),
				}
			} else if lt["value"] == nil || reflect.TypeOf(lt["value"]).String() != "float64" {
				return nil, structure.WebsocketAnswer{
					Error:     true,
					RequestID: requestID,
					Message:   fmt.Sprintf(structure.MUST_BY, fmt.Sprintf("$lt with index %d", i), "object with \"value\" (number)"),
				}
			}
		}
	}

	max := filter["$max"].(float64)
	if int(max) < 0 {
		return nil, structure.WebsocketAnswer{
			Error:     true,
			RequestID: requestID,
			Message:   fmt.Sprintf(structure.MUST_BY, "$max", ">= 0"),
		}
	}

	filter["$max"] = int(filter["$max"].(float64))
	return filter, structure.WebsocketAnswer{
		Error:     false,
		RequestID: requestID,
	}
}
