package structure

type WebsocketRequest struct {
	Type      string `json:"method"`
	RequestID int    `json:"requestID"`

	Database   string `json:"database"`
	Collection string `json:"collection"`

	Filter        map[string]interface{} `json:"filter"`
	DistributorID string                 `json:"distributorID"`

	Data interface{} `json:"data"`
}
