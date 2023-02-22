package config

import (
	"RediDB/modules/structure"
)

func Get() structure.Config {
	if !cache.GetInit() {
		load()
	}

	return cache
}
