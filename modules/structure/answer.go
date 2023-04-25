package structure

type WebsocketAnswer struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`

	Data interface{} `json:"data"`
}
