package model

type TimeSpan struct {
	ID       int      // Primary key
	Name     string   // Name of the time span
	StartUTC Time     // Start
	EndUTC   Time     // End
	TimeZone TimeZone // Timezone, in the "America/New_York" format
	UserID   int      // Id of the user who created the time span
}

type TimeSpanTag struct {
	ID         int // Primary key
	TimeSpanID int // ID of the time span
	TagID      int // ID of the tag
}
