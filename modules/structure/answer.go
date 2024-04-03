package structure

type WebsocketAnswer struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`

	RequestID int `json:"requestID"`

	DistributorID string      `json:"distributorID"`
	Residue       int         `json:"residue"`
	Data          interface{} `json:"data"`
}
