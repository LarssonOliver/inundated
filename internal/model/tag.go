package model

import (
	"bytes"
	"cmp"
	"slices"
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
	// Owner is set on owned tags, naming the item that owns the tag (see
	// TagOwnerKind). Regular tags have none.
	Owner *TagOwner
}

// TagOwnerKind is a kind of item that owns a tag of its own, such as a task
// and its task tag. An owned tag's name and archived state follow its
// owner, and it can't be edited or deleted through the tag API.
//
// A new kind of owner is declared here, in tagOwnerKinds and tagHolderRules,
// in the tag_owners database view and the memory store's tag lookup, and in
// the frontend's owner registry.
type TagOwnerKind string

const (
	TagOwnerTask    TagOwnerKind = "task"
	TagOwnerProject TagOwnerKind = "project"
)

// tagOwnerKind holds what differs between kinds of owners.
type tagOwnerKind struct {
	// derivesColor marks kinds whose tags take their color from source
	// tags (see DerivedTagColor) rather than keeping one of their own.
	derivesColor bool
}

var tagOwnerKinds = map[TagOwnerKind]tagOwnerKind{
	TagOwnerTask: {derivesColor: true},
	// A project tag keeps its project's color, stored on the tag.
	TagOwnerProject: {},
}

// Valid reports whether k is a known owner kind.
func (k TagOwnerKind) Valid() bool {
	_, ok := tagOwnerKinds[k]
	return ok
}

// TagOwner names the item that owns an owned tag.
type TagOwner struct {
	Kind TagOwnerKind
	Id   uuid.UUID
}

// OwnedBy reports whether the tag is owned by an item of kind k.
func (t Tag) OwnedBy(k TagOwnerKind) bool {
	return t.Owner != nil && t.Owner.Kind == k
}

// TagHolder is a kind of item that carries tags.
type TagHolder string

const (
	TagHolderTask     TagHolder = "task"
	TagHolderProject  TagHolder = "project"
	TagHolderTimespan TagHolder = "timespan"
)

// tagHolderRules lists the owned tag kinds each holder may carry; regular
// tags can go on any of them. Leaving a kind off its own holder keeps
// links between owners from forming loops: a task can't carry task tags
// (subtasks nest through their parent instead), and a project can't carry
// project tags.
var tagHolderRules = map[TagHolder][]TagOwnerKind{
	TagHolderTask:     {TagOwnerProject},
	TagHolderProject:  {TagOwnerTask},
	TagHolderTimespan: {TagOwnerTask, TagOwnerProject},
}

// CarriedOwnerKinds returns the owned tag kinds h may carry.
func (h TagHolder) CarriedOwnerKinds() []TagOwnerKind {
	return tagHolderRules[h]
}

// MayCarry reports whether h may carry tag: any regular tag, or an owned
// tag of a kind h allows.
func (h TagHolder) MayCarry(tag Tag) bool {
	return tag.Owner == nil || slices.Contains(tagHolderRules[h], tag.Owner.Kind)
}

// DefaultDerivedTagColor is the color a derived tag reports when it has no
// source tag to borrow a color from (Nord's nord10).
const DefaultDerivedTagColor = "#5e81ac"

// IsDerived reports whether the tag's color is derived from other tags (see
// DerivedTagColor) rather than set on the tag itself. Task tags are derived:
// they take their color from their task's regular tags.
func (t Tag) IsDerived() bool {
	return t.Owner != nil && tagOwnerKinds[t.Owner.Kind].derivesColor
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
