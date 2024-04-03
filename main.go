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

const (
	MIN_THREADS  = 10000
	MIN_GARBAGE  = 1
	MIN_MAX_DATA = 1

	MIN_DISTRIBUTE_FROM = 2
	MIN_DISTRIBUTE_GIVE = 2

	MIN_WORKER_TASK = 1
	MAX_WORKER_TASK = 100000
)

func init() {
	log.Println("Preparing to start...")
	config := config.Get()

	threads := config.Settings.MaxThreads
	if threads < MIN_THREADS {
		log.Panicf("Minimum count of settings.max_threads is %v", MIN_THREADS)
	}

	workerTasks := config.Settings.TasksCount
	if workerTasks < MIN_WORKER_TASK {
		log.Panicf("Minimum count of settings.worker_tasks is %v", MIN_WORKER_TASK)
	} else if workerTasks > MAX_WORKER_TASK {
		log.Panicf("Maximum count of settings.worker_tasks is %v", MAX_WORKER_TASK)
	}

	garbage := config.Garbage
	if garbage.Enabled && garbage.Interval < MIN_GARBAGE {
		log.Panicf("Minimum count of garbage.interval is %v", MIN_GARBAGE)
	}

	if config.Settings.MaxData < MIN_MAX_DATA {
		log.Panicf("Minimum count of settings.max_data is %v", MIN_MAX_DATA)
	}

	distribute := config.Distribute
	if distribute.StartFrom < MIN_DISTRIBUTE_FROM {
		log.Panicf("Minimum count of distribute.from is %v", MIN_DISTRIBUTE_FROM)
	}

	if distribute.GiveMax < MIN_DISTRIBUTE_GIVE {
		log.Panicf("Minimum count of distribute.give_at_one_call is %v", MIN_DISTRIBUTE_GIVE)
	}

	debug.SetMaxThreads(threads)
	path.Create()
}

func main() {
	memcache.Load()

	if err := handler.App.Listen(":" + strconv.Itoa(config.Get().Web.Port)); err != nil {
		log.Fatalln("Failed to listen server: ", err)
	}
}
