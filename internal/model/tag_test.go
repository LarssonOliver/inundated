package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDerivedTagColor(t *testing.T) {
	low := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	high := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	tests := []struct {
		name    string
		sources []Tag
		want    string
	}{
		{"no sources", nil, DefaultDerivedTagColor},
		{"first by name ignoring case", []Tag{{Name: "banana", Color: "#222222"}, {Name: "Apple", Color: "#111111"}}, "#111111"},
		{"same name folded, then by bytes", []Tag{{Name: "apple", Color: "#222222"}, {Name: "Apple", Color: "#111111"}}, "#111111"},
		{"same name, then by id", []Tag{{Id: high, Name: "a", Color: "#222222"}, {Id: low, Name: "a", Color: "#111111"}}, "#111111"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, DerivedTagColor(tt.sources))
		})
	}
}

func TestTagHolderMayCarry(t *testing.T) {
	label := Tag{Id: uuid.New()}
	taskTag := Tag{Id: uuid.New(), Owner: &TagOwner{Kind: TagOwnerTask, Id: uuid.New()}}
	projectTag := Tag{Id: uuid.New(), Owner: &TagOwner{Kind: TagOwnerProject, Id: uuid.New()}}

	tests := []struct {
		holder TagHolder
		tag    Tag
		want   bool
	}{
		{TagHolderTask, label, true},
		{TagHolderTask, taskTag, false},
		{TagHolderTask, projectTag, true},
		{TagHolderProject, projectTag, false},
		{TagHolderTimespan, projectTag, true},
		{TagHolderProject, label, true},
		{TagHolderProject, taskTag, true},
		{TagHolderTimespan, label, true},
		{TagHolderTimespan, taskTag, true},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, tt.holder.MayCarry(tt.tag), "%s carrying owner %v", tt.holder, tt.tag.Owner)
	}
}

func TestTagIsDerived(t *testing.T) {
	require.False(t, Tag{}.IsDerived())
	require.True(t, Tag{Owner: &TagOwner{Kind: TagOwnerTask}}.IsDerived())
	require.False(t, Tag{Owner: &TagOwner{Kind: TagOwnerProject}}.IsDerived())
}

func TestTagKindOwnerKind(t *testing.T) {
	kind, ok := TagKindTask.OwnerKind()
	require.True(t, ok)
	require.Equal(t, TagOwnerTask, kind)

	_, ok = TagKindLabel.OwnerKind()
	require.False(t, ok)
	_, ok = TagKindAll.OwnerKind()
	require.False(t, ok)
}

func TestTagKindValid(t *testing.T) {
	for _, k := range []TagKind{TagKindLabel, TagKindTask, TagKindProject, TagKindAll} {
		require.True(t, k.Valid(), k)
	}
	require.False(t, TagKind("").Valid())
	require.False(t, TagKind("tag").Valid())
}

func TestTagListParamsSelectsKind(t *testing.T) {
	task := &TagOwner{Kind: TagOwnerTask}
	project := &TagOwner{Kind: TagOwnerProject}

	byDefault := TagListParams{}
	require.True(t, byDefault.SelectsKind(nil))
	require.False(t, byDefault.SelectsKind(task))

	some := TagListParams{Kinds: []TagKind{TagKindLabel, TagKindTask}}
	require.True(t, some.SelectsKind(nil))
	require.True(t, some.SelectsKind(task))
	require.False(t, some.SelectsKind(project))

	owned := TagListParams{Kinds: []TagKind{TagKindProject}}
	require.False(t, owned.SelectsKind(nil))
	require.True(t, owned.SelectsKind(project))

	all := TagListParams{Kinds: []TagKind{TagKindAll}}
	require.True(t, all.SelectsKind(nil))
	require.True(t, all.SelectsKind(task))
	require.True(t, all.SelectsKind(project))
}
