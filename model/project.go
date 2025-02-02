package model

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
