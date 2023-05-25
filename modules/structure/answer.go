package structure

type WebsocketAnswer struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`

	RequestID int         `json:"requestID"`
	Data      interface{} `json:"data"`
}
