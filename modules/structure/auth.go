package structure

type Auth struct {
	Database   string `json:"database"`
	Collection string `json:"collection"`

	Login    string `json:"login"`
	Password string `json:"password"`
}
