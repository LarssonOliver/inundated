package model

// CaldavRef holds how a CalDAV client identified an object it created, so
// the object is served back the same way. Both fields are nil for objects
// created any other way.
type CaldavRef struct {
	// Name is the object name in the client's href.
	Name *string
	// Uid is the iCalendar UID the client gave the object.
	Uid *string
}
