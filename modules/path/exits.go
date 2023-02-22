package path

import "os"

func isExits() bool {
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		return false
	}

	return true
}
