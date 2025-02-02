package model

type Tag struct {
	ID     int    // Primary key
	Name   string // Name of the tag
	Color  string // Color of the tag
	UserID int    // Owner of the tag
}
