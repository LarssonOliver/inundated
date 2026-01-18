package model

import "time"

type Uuid string

type Project struct {
	Id         Uuid
	Name       string
	Color      string
	TagIds     []string
	TimeBudget float64
}

type Tag struct {
	Id    Uuid
	Name  string
	Color string
}

type TimeSpan struct {
	Id        Uuid
	Name      string
	EndTime   time.Time
	StartTime time.Time
}
