package model

import "time"

// TODO - Split into separate files

type User struct {
	ID          int    // Primary key
	Name        string // Unique user name, likely comes from OIDC provider
	DisplayName string // Display name of the user
}

type Time time.Time  // Timestamp respresentation
type TimeZone string // Timezone representation

type TimeSpan struct {
	ID       int      // Primary key
	Name     string   // Name of the time span
	StartUTC Time     // Start
	EndUTC   Time     // End
	TZ       TimeZone // Timezone, I.e. in the "America/New_York" format
	UserID   int      // Id of the user who created the time span
}

type Tag struct {
	ID     int    // Primary key
	Name   string // Name of the tag
	Color  string // Color of the tag
	UserID int    // Owner of the tag
}

type TimeSpanTag struct {
	ID         int // Primary key
	TimeSpanID int // ID of the time span
	TagID      int // ID of the tag
}

type Project struct {
	ID         int    // Primary key
	Name       string // Name of the project
	Color      string // Color of the project
	TimeBudget int    // Time budget for the project
	UserID     int    // Owner of the project
}

// Tags associated with a project will count towards the projects budget and
// total time spent
type ProjectTag struct {
	ID        int // Primary key
	ProjectID int // ID of the project
	TagID     int // ID of the tag
}
