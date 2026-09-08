package memory_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/stretchr/testify/require"
)

var testScope = model.UserScope(uuid.MustParse("11111111-1111-1111-1111-111111111111"))

func TestMemoryStore_Tag_ScopeIsolation(t *testing.T) {
	ctx := context.Background()
	store := memory.NewMemoryStore()
	a := model.UserScope(uuid.New())
	b := model.UserScope(uuid.New())

	tag, err := store.CreateTag(ctx, a, model.Tag{Name: "x", Color: "#123456"})
	require.NoError(t, err)

	_, err = store.GetTag(ctx, b, tag.Id)
	require.ErrorIs(t, err, model.ErrNotFound)

	page, err := store.ListTags(ctx, b, model.DefaultPaginationParams())
	require.NoError(t, err)
	require.Empty(t, page.Data)
}

func TestTagStore_CreateTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     model.Tag
		want    model.Tag
		wantErr bool
		errType error
	}{
		{
			name:    "Test CreateTag with valid input",
			tag:     model.Tag{Name: "Urgent", Color: "#FF0000"},
			want:    model.Tag{Name: "Urgent", Color: "#FF0000"},
			wantErr: false,
		},
		{
			name:    "Test CreateTag with another valid input",
			tag:     model.Tag{Name: "Optional", Color: "#00FF00"},
			want:    model.Tag{Name: "Optional", Color: "#00FF00"},
			wantErr: false,
		},
		{
			name:    "Test CreateTag with empty name",
			tag:     model.Tag{Name: "", Color: "#0000FF"},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateTag with invalid color",
			tag:     model.Tag{Name: "InvalidColor", Color: "NotAColor"},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateTag with empty color",
			tag:     model.Tag{Name: "NoColor", Color: ""},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateTag with nil tag",
			tag:     model.Tag{},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test CreateTag with set ID (should be ignored)",
			tag:     model.Tag{Id: uuid.New(), Name: "WithID", Color: "#123456"},
			want:    model.Tag{Name: "WithID", Color: "#123456"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()
			got, gotErr := ta.CreateTag(context.Background(), testScope, tt.tag)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
			require.NotEqual(t, uuid.Nil, got.Id)
		})
	}
}

func TestTagStore_GetTag(t *testing.T) {
	tests := []struct {
		name      string
		createTag model.Tag
		getId     func(createdTag *model.Tag) uuid.UUID
		want      model.Tag
		wantErr   bool
		errType   error
	}{
		{
			name:      "Test GetTag with existing ID",
			createTag: model.Tag{Name: "Tag1", Color: "#FF0000"},
			getId: func(createdTag *model.Tag) uuid.UUID {
				return createdTag.Id
			},
			want:    model.Tag{Name: "Tag1", Color: "#FF0000"},
			wantErr: false,
		},
		{
			name:      "Test GetTag with non-existing ID",
			createTag: model.Tag{Name: "Tag2", Color: "#00FF00"},
			getId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.New()
			},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:      "Test GetTag with empty UUID",
			createTag: model.Tag{Name: "Tag3", Color: "#0000FF"},
			getId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.Nil
			},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()
			tag, _ := ta.CreateTag(context.Background(), testScope, tt.createTag)
			getId := tt.getId(&tag)

			got, gotErr := ta.GetTag(context.Background(), testScope, getId)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
			require.NotEqual(t, uuid.Nil, got.Id)
		})
	}
}

func TestTagStore_ListTags(t *testing.T) {
	tests := []struct {
		name       string
		insertTags []model.Tag
		params     model.PaginationParams
		wantLen    int
		wantTotal  int
		wantErr    bool
	}{
		{
			name: "Test ListTags with multiple tags default pagination",
			insertTags: []model.Tag{
				{Name: "Tag1", Color: "#FF0000"},
				{Name: "Tag2", Color: "#00FF00"},
				{Name: "Tag3", Color: "#0000FF"},
			},
			params:    model.DefaultPaginationParams(),
			wantLen:   3,
			wantTotal: 3,
		},
		{
			name:       "Test ListTags with no tags",
			insertTags: []model.Tag{},
			params:     model.DefaultPaginationParams(),
			wantLen:    0,
			wantTotal:  0,
		},
		{
			name:       "Test ListTags with one tag",
			insertTags: []model.Tag{{Name: "OnlyTag", Color: "#123456"}},
			params:     model.DefaultPaginationParams(),
			wantLen:    1,
			wantTotal:  1,
		},
		{
			name:       "Test ListTags with duplicate tags",
			insertTags: []model.Tag{{Name: "DupTag", Color: "#654321"}, {Name: "DupTag", Color: "#654321"}},
			params:     model.DefaultPaginationParams(),
			wantLen:    2,
			wantTotal:  2,
		},
		{
			name: "Test ListTags with limit",
			insertTags: []model.Tag{
				{Name: "Tag1", Color: "#FF0000"},
				{Name: "Tag2", Color: "#00FF00"},
				{Name: "Tag3", Color: "#0000FF"},
			},
			params:    model.PaginationParams{Limit: 2, Offset: 0},
			wantLen:   2,
			wantTotal: 3,
		},
		{
			name: "Test ListTags with offset",
			insertTags: []model.Tag{
				{Name: "Tag1", Color: "#FF0000"},
				{Name: "Tag2", Color: "#00FF00"},
				{Name: "Tag3", Color: "#0000FF"},
			},
			params:    model.PaginationParams{Limit: 10, Offset: 2},
			wantLen:   1,
			wantTotal: 3,
		},
		{
			name: "Test ListTags with offset beyond end",
			insertTags: []model.Tag{
				{Name: "Tag1", Color: "#FF0000"},
			},
			params:    model.PaginationParams{Limit: 10, Offset: 100},
			wantLen:   0,
			wantTotal: 1,
		},
		{
			name: "Test ListTags total count unaffected by limit",
			insertTags: []model.Tag{
				{Name: "Tag1", Color: "#FF0000"},
				{Name: "Tag2", Color: "#00FF00"},
				{Name: "Tag3", Color: "#0000FF"},
			},
			params:    model.PaginationParams{Limit: 1, Offset: 0},
			wantLen:   1,
			wantTotal: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			insertedIds := make(map[uuid.UUID]bool)
			for i, tag := range tt.insertTags {
				createdTag, _ := ta.CreateTag(context.Background(), testScope, tag)
				tt.insertTags[i].Id = createdTag.Id
				insertedIds[createdTag.Id] = true
			}

			page, gotErr := ta.ListTags(context.Background(), testScope, tt.params)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			require.Len(t, page.Data, tt.wantLen)
			require.Equal(t, tt.wantTotal, page.TotalCount)
			require.Equal(t, tt.params.Limit, page.Limit)
			require.Equal(t, tt.params.Offset, page.Offset)

			for _, tag := range page.Data {
				require.True(t, insertedIds[tag.Id], "returned tag not in inserted set")
			}
		})
	}
}

func TestTagStore_UpdateTag(t *testing.T) {
	tests := []struct {
		name      string
		tag       model.Tag
		editTag   model.Tag
		editTagId func(createdTag *model.Tag) uuid.UUID
		want      model.Tag
		wantErr   bool
		errType   error
	}{
		{
			name:    "Test UpdateTag with valid input",
			tag:     model.Tag{Name: "OldName", Color: "#FF0000"},
			editTag: model.Tag{Name: "NewName", Color: "#00FF00"},
			editTagId: func(createdTag *model.Tag) uuid.UUID {
				return createdTag.Id
			},
			want:    model.Tag{Name: "NewName", Color: "#00FF00"},
			wantErr: false,
		},
		{
			name:    "Test UpdateTag with non-existing ID",
			tag:     model.Tag{Name: "Tag1", Color: "#FF0000"},
			editTag: model.Tag{Name: "ShouldFail", Color: "#0000FF"},
			editTagId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.New()
			},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:    "Test UpdateTag with empty name",
			tag:     model.Tag{Name: "Tag2", Color: "#00FF00"},
			editTag: model.Tag{Name: "", Color: "#0000FF"},
			editTagId: func(createdTag *model.Tag) uuid.UUID {
				return createdTag.Id
			},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test UpdateTag with invalid color",
			tag:     model.Tag{Name: "Tag3", Color: "#0000FF"},
			editTag: model.Tag{Name: "Tag3", Color: "InvalidColor"},
			editTagId: func(createdTag *model.Tag) uuid.UUID {
				return createdTag.Id
			},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrInvalidArgument,
		},
		{
			name:    "Test UpdateTag with empty ID",
			tag:     model.Tag{Name: "Tag4", Color: "#123456"},
			editTag: model.Tag{Name: "Tag4Updated", Color: "#654321"},
			editTagId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.Nil
			},
			want:    model.Tag{},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:    "Test UpdateTag with same name and color",
			tag:     model.Tag{Name: "Tag5", Color: "#ABCDEF"},
			editTag: model.Tag{Name: "Tag5", Color: "#ABCDEF"},
			editTagId: func(createdTag *model.Tag) uuid.UUID {
				return createdTag.Id
			},
			want:    model.Tag{Name: "Tag5", Color: "#ABCDEF"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			insertedTag, _ := ta.CreateTag(context.Background(), testScope, tt.tag)
			editId := tt.editTagId(&insertedTag)

			tt.editTag.Id = editId
			tt.want.Id = editId

			got, gotErr := ta.UpdateTag(context.Background(), testScope, tt.editTag)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			require.Equal(t, tt.want.Name, got.Name)
			require.Equal(t, tt.want.Color, got.Color)
			require.NotEqual(t, uuid.Nil, got.Id)
		})
	}
}

func TestTagStore_DeleteTag(t *testing.T) {
	tests := []struct {
		name      string
		insertTag model.Tag
		deleteId  func(createdTag *model.Tag) uuid.UUID
		wantErr   bool
		errType   error
	}{
		{
			name:      "Test DeleteTag with existing ID",
			insertTag: model.Tag{Name: "Tag1", Color: "#FF0000"},
			deleteId: func(createdTag *model.Tag) uuid.UUID {
				return createdTag.Id
			},
			wantErr: false,
		},
		{
			name:      "Test DeleteTag with non-existing ID",
			insertTag: model.Tag{},
			deleteId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
			errType: model.ErrNotFound,
		},
		{
			name:      "Test DeleteTag with empty UUID",
			insertTag: model.Tag{Name: "Tag3", Color: "#0000FF"},
			deleteId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.Nil
			},
			wantErr: true,
			errType: model.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewMemoryStore()

			tag, _ := ta.CreateTag(context.Background(), testScope, tt.insertTag)
			deleteId := tt.deleteId(&tag)

			gotErr := ta.DeleteTag(context.Background(), testScope, deleteId)
			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.errType != nil {
					require.ErrorIs(t, gotErr, tt.errType)
				}
				return
			}

			require.NoError(t, gotErr)
			_, err := ta.GetTag(context.Background(), testScope, deleteId)
			require.ErrorIs(t, err, model.ErrNotFound)
		})
	}
}
