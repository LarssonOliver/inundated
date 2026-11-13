package caldav

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/stretchr/testify/require"
)

func id(last byte) uuid.UUID {
	return uuid.UUID{15: last}
}

func owned(tagId uuid.UUID, name string, kind model.TagOwnerKind) model.Tag {
	return model.Tag{Id: tagId, Name: name, Owner: &model.TagOwner{Kind: kind, Id: uuid.UUID{}}}
}

var (
	work    = model.Tag{Id: id(1), Name: "work"}
	alpha   = model.Tag{Id: id(2), Name: "alpha"}
	taskTag = owned(id(3), "Write report", model.TagOwnerTask)
	project = owned(id(4), "Apollo", model.TagOwnerProject)
	tags    = map[uuid.UUID]model.Tag{
		work.Id: work, alpha.Id: alpha, taskTag.Id: taskTag, project.Id: project,
	}

	start = time.Date(2024, 3, 5, 9, 0, 0, 0, time.UTC)
	end   = time.Date(2024, 3, 5, 10, 30, 0, 0, time.UTC)
)

func requireGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	want, err := os.ReadFile(filepath.Join("testdata", name+".ics"))
	require.NoError(t, err)
	require.Equal(t, string(want), string(got))
}

func TestTimespanEvent(t *testing.T) {
	stockholm, err := time.LoadLocation("Europe/Stockholm")
	require.NoError(t, err)

	tests := []struct {
		name     string
		timespan model.Timespan
		location *time.Location
	}{
		{"timespan_named", model.Timespan{
			Id: id(10), Name: "Standup", StartTime: start, EndTime: end,
			TagIds: []uuid.UUID{work.Id, taskTag.Id, alpha.Id},
		}, time.UTC},
		{"timespan_unnamed_owned_tags", model.Timespan{
			Id: id(11), StartTime: start, EndTime: end,
			TagIds: []uuid.UUID{taskTag.Id, project.Id},
		}, time.UTC},
		{"timespan_untitled", model.Timespan{
			Id: id(12), StartTime: start, EndTime: end,
			TagIds: []uuid.UUID{id(99)},
		}, time.UTC},
		{"timespan_local", model.Timespan{
			Id: id(13), Name: "Standup", StartTime: start, EndTime: end,
		}, stockholm},
		// Summer time began at 01:00 UTC that day.
		{"timespan_local_across_dst", model.Timespan{
			Id:        id(14),
			Name:      "Night shift",
			StartTime: time.Date(2024, 3, 31, 0, 30, 0, 0, time.UTC),
			EndTime:   time.Date(2024, 3, 31, 1, 30, 0, 0, time.UTC),
		}, stockholm},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _, err := Encode(TimespanEvent(test.timespan, tags, test.location))
			require.NoError(t, err)
			requireGolden(t, test.name, got)
		})
	}
}

func TestTaskTodo(t *testing.T) {
	due := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	closedAt := time.Date(2024, 3, 20, 15, 4, 5, 0, time.UTC)
	done := model.CloseReasonDone
	ignored := model.CloseReasonIgnored
	parent := id(20)

	tests := []struct {
		name string
		task model.Task
	}{
		{"task_open_subtask", model.Task{
			Id: id(21), Name: "Draft outline", ParentId: &parent, DueDate: &due,
			TagIds: []uuid.UUID{work.Id, project.Id},
		}},
		{"task_done", model.Task{
			Id: id(22), Name: "Ship it", CloseReason: &done, ClosedAt: &closedAt,
		}},
		{"task_ignored", model.Task{
			Id: id(23), Name: "Skip it", CloseReason: &ignored, ClosedAt: &closedAt,
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _, err := Encode(TaskTodo(test.task, tags))
			require.NoError(t, err)
			requireGolden(t, test.name, got)
		})
	}
}

func TestTimespanEvent_StableBytes(t *testing.T) {
	timespan := model.Timespan{Id: id(10), StartTime: start, EndTime: end}

	timespan.TagIds = []uuid.UUID{work.Id, alpha.Id, taskTag.Id, project.Id}
	first, firstETag, err := Encode(TimespanEvent(timespan, tags, time.UTC))
	require.NoError(t, err)

	timespan.TagIds = []uuid.UUID{project.Id, taskTag.Id, alpha.Id, work.Id}
	second, secondETag, err := Encode(TimespanEvent(timespan, tags, time.UTC))
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.Equal(t, firstETag, secondETag)
}
