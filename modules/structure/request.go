package structure

type WebsocketRequest struct {
	Type string `json:"method"`

	Database   string `json:"database"`
	Collection string `json:"collection"`

	Filter map[string]interface{} `json:"filter"`
	Data   interface{}            `json:"data"`
}
