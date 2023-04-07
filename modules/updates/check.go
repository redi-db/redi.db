package updates

import (
	"io"
	"net/http"
	"strings"
)

func Check() (string, bool, error) {
	resp, err := http.Get(VERSION_PATH)
	if err != nil {
		return "", false, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}

	latestVersion := strings.TrimSpace(string(body))
	if VERSION != latestVersion {
		return latestVersion, true, nil
	}

	return VERSION, false, nil
}
