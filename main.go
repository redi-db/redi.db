package main

import (
	"RediDB/modules/config"
	"RediDB/modules/handler"
	"RediDB/modules/memcache"
	"RediDB/modules/path"
	"log"
	"runtime/debug"
	"strconv"
)

func init() {
	log.Println("Preparing to start...")
	config := config.Get()

	threads := config.Settings.MaxThreads
	if threads < 10000 {
		log.Panicln("Minimum count of settings.max_threads is 10000")
	}

	data := config.Settings.MaxData
	if data < 1 {
		log.Panicln("Minimum count of settings.max_data is 1")
	}

	garbage := config.Garbage
	if garbage.Enabled && garbage.Interval < 1 {
		log.Panicln("Minimum count of garbage.interval is 1")
	}

	debug.SetMaxThreads(threads)
	path.Create()
	memcache.Load()
}

func main() {
	if err := handler.App.Listen(":" + strconv.Itoa(config.Get().Web.Port)); err != nil {
		log.Fatalln("Failed to listen server: ", err)
	}
}
