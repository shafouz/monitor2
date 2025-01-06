package models

import "time"

type Endpoint struct {
	Id                   int
	Url                  string
	StatusCode           int
	ResponseBody         []byte
	PreviousResponseBody []byte
	ScheduleHours        int
	Selector             string
	Profile              string
	Deleted              bool
	UpdatedAt            time.Time
}

type Repository struct {
	Id            int
	Url           string
	Directory     string
	WatchedFiles  []byte
	Remote        string
	ScheduleHours int
	Deleted       bool
	UpdatedAt     time.Time
}

type Diff struct {
	Id        string
	Body      string
	Url       string
  Commit    string
	CreatedAt time.Time
}
