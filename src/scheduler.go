package monitor2

import (
	"log"
	"monitor2/src/crawler"
	"monitor2/src/repositories"
	database "monitor2/src/db"
	"time"
)

func make_schedule(duration int) time.Duration {
	return time.Duration(duration) * time.Hour
}

func StartScheduler() {
	// time.Duration(1) * time.Second
	jobs := []int{
    8,
    24,
    24*7,
	}

	for _, schedule := range jobs {
		go run_at(crawler.RunBySchedule, schedule, "crawler", &database.DB)
		go run_at(repositories.RunBySchedule, schedule, "repos", &database.DB)
	}
}

type RowsAffected = int
type Task func(duration int, DB *database.Database) (RowsAffected, []error)

func run_at(fn Task, duration int, func_name string, DB *database.Database) {
	for {
    log.Printf("Running %+v, duration: %d\n", func_name, duration)
		fn(duration, DB)
		time.Sleep(make_schedule(duration))
	}
}
