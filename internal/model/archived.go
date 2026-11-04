package model

import "time"

// ArchivedAtAfter returns when an item archived since prev (nil if it
// wasn't) is archived once its archived flag is set to archived at now:
// an item that stays archived keeps its original time.
func ArchivedAtAfter(prev *time.Time, archived bool, now time.Time) *time.Time {
	switch {
	case !archived:
		return nil
	case prev != nil:
		return prev
	default:
		return &now
	}
}

// CompareArchived orders active items before archived ones, and archived
// ones most recently archived first. Items it ranks equal are left for the
// caller to order, e.g. by name.
func CompareArchived(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1
	case b == nil:
		return 1
	default:
		return b.Compare(*a)
	}
}
