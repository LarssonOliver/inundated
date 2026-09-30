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

// DefaultDerivedTagColor is the color a derived tag reports when it has no
// source tag to borrow a color from (Nord's nord10).
const DefaultDerivedTagColor = "#5e81ac"

// IsDerived reports whether the tag's color is derived from other tags (see
// DerivedTagColor) rather than set on the tag itself. Task tags are derived:
// they take their color from their task's regular tags.
func (t Tag) IsDerived() bool {
	return t.TaskId != nil
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

// DerivedTagColor is the color of a derived tag whose color comes from
// sources: that of the first source by CompareTagNames, or
// DefaultDerivedTagColor when there are none.
func DerivedTagColor(sources []Tag) string {
	if len(sources) == 0 {
		return DefaultDerivedTagColor
	}
	first := sources[0]
	for _, source := range sources[1:] {
		if CompareTagNames(source, first) < 0 {
			first = source
		}
	}
	return first.Color
}
