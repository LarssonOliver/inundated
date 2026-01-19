package memory_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository/memory"
)

func TestTagStore_CreateTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     model.Tag
		want    model.Tag
		wantErr bool
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
		},
		{
			name:    "Test CreateTag with invalid color",
			tag:     model.Tag{Name: "InvalidColor", Color: "NotAColor"},
			want:    model.Tag{},
			wantErr: true,
		},
		{
			name:    "Test CreateTag with empty color",
			tag:     model.Tag{Name: "NoColor", Color: ""},
			want:    model.Tag{},
			wantErr: true,
		},
		{
			name:    "Test CreateTag with nil tag",
			tag:     model.Tag{},
			want:    model.Tag{},
			wantErr: true,
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
			ta := memory.NewTagStore()
			got, gotErr := ta.CreateTag(context.Background(), tt.tag)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTag() succeeded unexpectedly")
			}
			if got.Name != tt.want.Name || got.Color != tt.want.Color || got.Id == tt.want.Id {
				t.Errorf("CreateTag() = %v, want %v", got, tt.want)
			}
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
		},
		{
			name:      "Test GetTag with empty UUID",
			createTag: model.Tag{Name: "Tag3", Color: "#0000FF"},
			getId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.Nil
			},
			want:    model.Tag{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTagStore()
			tag, _ := ta.CreateTag(context.Background(), tt.createTag)
			getId := tt.getId(&tag)

			got, gotErr := ta.GetTag(context.Background(), getId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTag() succeeded unexpectedly")
			}
			if tag.Name != tt.want.Name || tag.Color != tt.want.Color || tag.Id != got.Id {
				t.Errorf("GetTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagStore_ListTags(t *testing.T) {
	tests := []struct {
		name       string // description of this test case
		insertTags []model.Tag
		wantErr    bool
	}{
		{
			name: "Test ListTags with multiple tags",
			insertTags: []model.Tag{
				{Name: "Tag1", Color: "#FF0000"},
				{Name: "Tag2", Color: "#00FF00"},
				{Name: "Tag3", Color: "#0000FF"},
			},
			wantErr: false,
		},
		{
			name:       "Test ListTags with no tags",
			insertTags: []model.Tag{},
			wantErr:    false,
		},
		{
			name: "Test ListTags with one tag",
			insertTags: []model.Tag{
				{Name: "OnlyTag", Color: "#123456"},
			},
		},
		{
			name:       "Test ListTags with duplicate tags",
			insertTags: []model.Tag{{Name: "DupTag", Color: "#654321"}, {Name: "DupTag", Color: "#654321"}},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTagStore()

			for i, tag := range tt.insertTags {
				createdTag, _ := ta.CreateTag(context.Background(), tag)
				tt.insertTags[i].Id = createdTag.Id
			}

			got, gotErr := ta.ListTags(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTags() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTags() succeeded unexpectedly")
			}

			if len(got) != len(tt.insertTags) {
				t.Errorf("ListTags() = %v, want %v", got, tt.insertTags)
			}

			for _, tag := range tt.insertTags {
				found := false
				for _, gotTag := range got {
					if gotTag.Id == tag.Id && gotTag.Name == tag.Name && gotTag.Color == tag.Color {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ListTags() missing expected tag: %v", tag)
				}
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
			ta := memory.NewTagStore()

			insertedTag, _ := ta.CreateTag(context.Background(), tt.tag)
			editId := tt.editTagId(&insertedTag)

			tt.editTag.Id = editId
			tt.want.Id = editId

			got, gotErr := ta.UpdateTag(context.Background(), tt.editTag)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTag() succeeded unexpectedly")
			}
			if tt.want.Name != got.Name || tt.want.Color != got.Color || tt.want.Id != got.Id {
				t.Errorf("UpdateTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagStore_DeleteTag(t *testing.T) {
	tests := []struct {
		name      string
		insertTag model.Tag
		deleteId  func(createdTag *model.Tag) uuid.UUID
		wantErr   bool
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
		},
		{
			name:      "Test DeleteTag with empty UUID",
			insertTag: model.Tag{Name: "Tag3", Color: "#0000FF"},
			deleteId: func(createdTag *model.Tag) uuid.UUID {
				return uuid.Nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ta := memory.NewTagStore()

			tag, _ := ta.CreateTag(context.Background(), tt.insertTag)
			deleteId := tt.deleteId(&tag)

			gotErr := ta.DeleteTag(context.Background(), deleteId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteTag() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteTag() succeeded unexpectedly")
			}
			tag, err := ta.GetTag(context.Background(), deleteId)
			if err == nil {
				t.Errorf("Tag with ID %v was not deleted, still exists: %v", deleteId, tag)
			}
		})
	}
}
