package config

import (
	"os"
)

func isExits() bool {
	if _, err := os.Stat(configName); os.IsNotExist(err) {
		return false
	}

	return true
}
