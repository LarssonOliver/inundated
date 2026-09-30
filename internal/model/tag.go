package model

import (
	"bytes"
	"cmp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	Id        uuid.UUID
	Name      string
	Color     string
	TotalTime *time.Duration
	UserId    *uuid.UUID
	Archived  bool
	// TaskId is set on task tags. A task tag's name, color and archived
	// state follow its task.
	TaskId *uuid.UUID
}

// CompareTagNames orders tags case-insensitively by name, then by name
// bytes, then by id: the order tags are listed in.
func CompareTagNames(a, b Tag) int {
	return cmp.Or(
		strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
		strings.Compare(a.Name, b.Name),
		bytes.Compare(a.Id[:], b.Id[:]),
	)
}

// TaskTagColor is the color of a task tag whose task carries labels as its
// regular tags: that of the first label by CompareTagNames, or
// DefaultTaskTagColor when there are none.
func TaskTagColor(labels []Tag) string {
	if len(labels) == 0 {
		return DefaultTaskTagColor
	}
	first := labels[0]
	for _, label := range labels[1:] {
		if CompareTagNames(label, first) < 0 {
			first = label
		}
	}
	return first.Color
}
