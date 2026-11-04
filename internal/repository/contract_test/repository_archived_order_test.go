package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/test/testutils"
	"github.com/stretchr/testify/require"
)

// Lists that include archived items show active ones by name, then archived
// ones most recently archived first.
func TestArchivedOrderContract(t *testing.T) {
	ctx := context.Background()
	includeArchived := model.PaginationParams{Limit: 50, IncludeArchived: true}

	// tick keeps archive times apart, so their order is certain.
	tick := func() { time.Sleep(5 * time.Millisecond) }

	run := func(t *testing.T, repoName string, newRepo func(t *testing.T) repository.Repository) {
		t.Run(repoName+"Tags", func(t *testing.T) {
			repo := newRepo(t)

			tags := map[string]model.Tag{}
			for _, name := range []string{"a", "b", "c", "d"} {
				created, err := repo.CreateTag(ctx, testScope, model.Tag{Name: name, Color: "#ffffff"})
				require.NoError(t, err)
				require.Nil(t, created.ArchivedAt)
				tags[name] = created
			}
			archive := func(name string, archived bool) model.Tag {
				tag := tags[name]
				tag.Archived = archived
				updated, err := repo.UpdateTag(ctx, testScope, tag)
				require.NoError(t, err)
				return updated
			}

			b := archive("b", true)
			require.NotNil(t, b.ArchivedAt)
			tick()
			archive("a", true)
			tick()
			// Saving an archived tag again keeps when it was archived.
			require.Equal(t, b.ArchivedAt.UnixMicro(), archive("b", true).ArchivedAt.UnixMicro())
			// Unarchiving clears it.
			archive("d", true)
			require.Nil(t, archive("d", false).ArchivedAt)

			page, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: includeArchived})
			require.NoError(t, err)
			require.Equal(t, []string{"c", "d", "a", "b"}, tagNames(page.Data))

			got, err := repo.GetTag(ctx, testScope, tags["b"].Id)
			require.NoError(t, err)
			require.Equal(t, b.ArchivedAt.UnixMicro(), got.ArchivedAt.UnixMicro())
		})

		t.Run(repoName+"OwnedTagsTakeTurns", func(t *testing.T) {
			repo := newRepo(t)
			reason := model.CloseReasonDone

			for _, name := range []string{"x", "y", "z"} {
				_, err := repo.CreateTag(ctx, testScope, model.Tag{Name: name, Color: "#ffffff"})
				require.NoError(t, err)
			}
			closed := map[string]model.Task{}
			for _, name := range []string{"t1", "t2", "t3"} {
				task, err := repo.CreateTask(ctx, testScope, model.Task{Name: name})
				require.NoError(t, err)
				closed[name] = task
			}
			labels, err := repo.ListTags(ctx, testScope, model.TagListParams{PaginationParams: includeArchived})
			require.NoError(t, err)
			byName := map[string]model.Tag{}
			for _, tag := range labels.Data {
				byName[tag.Name] = tag
			}

			// Archive in the order y, t2, z, t3: the latest of each kind
			// comes first among the archived ones.
			archiveLabel := func(name string) {
				tag := byName[name]
				tag.Archived = true
				_, err := repo.UpdateTag(ctx, testScope, tag)
				require.NoError(t, err)
				tick()
			}
			closeTask := func(name string) model.Task {
				task, err := repo.UpdateTask(ctx, testScope, closed[name].Id, model.TaskPatch{CloseReason: &reason})
				require.NoError(t, err)
				tick()
				return task
			}
			archiveLabel("y")
			closeTask("t2")
			archiveLabel("z")
			t3 := closeTask("t3")

			page, err := repo.ListTags(ctx, testScope, model.TagListParams{
				PaginationParams: includeArchived,
				Kinds:            []model.TagKind{model.TagKindTask, model.TagKindLabel},
			})
			require.NoError(t, err)
			require.Equal(t, []string{"x", "t1", "z", "t3", "y", "t2"}, tagNames(page.Data))

			// A task tag was archived when its task was closed.
			got, err := repo.GetTag(ctx, testScope, t3.TagId)
			require.NoError(t, err)
			require.NotNil(t, got.ArchivedAt)
			require.Equal(t, t3.ClosedAt.UnixMicro(), got.ArchivedAt.UnixMicro())
		})

		t.Run(repoName+"Projects", func(t *testing.T) {
			repo := newRepo(t)

			projects := map[string]model.Project{}
			for _, name := range []string{"a", "b", "c", "d"} {
				created, err := repo.CreateProject(ctx, testScope, model.Project{Name: name, Color: "#ffffff"})
				require.NoError(t, err)
				require.Nil(t, created.ArchivedAt)
				projects[name] = created
			}
			archive := func(name string, archived bool) model.Project {
				project := projects[name]
				project.Archived = archived
				updated, err := repo.UpdateProject(ctx, testScope, project)
				require.NoError(t, err)
				return updated
			}

			b := archive("b", true)
			require.NotNil(t, b.ArchivedAt)
			tick()
			archive("a", true)
			require.Equal(t, b.ArchivedAt.UnixMicro(), archive("b", true).ArchivedAt.UnixMicro())
			archive("d", true)
			require.Nil(t, archive("d", false).ArchivedAt)

			page, err := repo.ListProjects(ctx, testScope, includeArchived)
			require.NoError(t, err)
			names := []string{}
			for _, project := range page.Data {
				names = append(names, project.Name)
			}
			require.Equal(t, []string{"c", "d", "a", "b"}, names)

			// The project tag is archived with its project.
			tag, err := repo.GetTag(ctx, testScope, b.TagId)
			require.NoError(t, err)
			require.NotNil(t, tag.ArchivedAt)
			require.Equal(t, b.ArchivedAt.UnixMicro(), tag.ArchivedAt.UnixMicro())
		})
	}

	run(t, "memory", func(t *testing.T) repository.Repository {
		return memory.NewMemoryStore()
	})

	run(t, "postgres", func(t *testing.T) repository.Repository {
		t.Parallel()
		pool := testutils.StartPostgresContainerWithMigrationsApplied(ctx, t)
		repo := postgres.NewPostgresStoreFromPool(pool)
		seedScopeUser(t, ctx, repo, testScope)
		return repo
	})
}

func tagNames(tags []model.Tag) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		out = append(out, tag.Name)
	}
	return out
}
