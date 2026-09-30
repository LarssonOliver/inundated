package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

func (r *PostgresStore) GetProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Project, error) {
	if id == uuid.Nil {
		return model.Project{}, fmt.Errorf("GetProject: %w", errNilId)
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	q := `
		SELECT id, name, color, time_budget, user_id, archived_at, tag_id
		FROM projects
		WHERE id = $1 AND deleted_at IS NULL AND ` + ownerSQL

	var p model.Project
	var archivedAt *time.Time
	err := r.db.QueryRow(ctx, q, args...).Scan(&p.Id, &p.Name, &p.Color, &p.TimeBudget, &p.UserId, &archivedAt, &p.TagId)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Project{}, fmt.Errorf("GetProject %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("GetProject: %w", err)
	}
	p.Archived = archivedAt != nil

	p.TagIds, err = r.projectTagIds(ctx, r.db, id)
	if err != nil {
		return model.Project{}, err
	}
	return p, nil
}

func (r *PostgresStore) ListProjects(ctx context.Context, scope model.OwnerScope, params model.PaginationParams) (model.Page[model.Project], error) {
	archivedSQL := archivedFilter(params.IncludeArchived)

	countOwnerSQL, countArgs := ownerPredicate("user_id", scope, nil)
	countQ := `
		SELECT COUNT(*)
		FROM projects
		WHERE deleted_at IS NULL AND ` + archivedSQL + countOwnerSQL

	var totalCount int
	if err := r.db.QueryRow(ctx, countQ, countArgs...).Scan(&totalCount); err != nil {
		return model.Page[model.Project]{}, fmt.Errorf("ListProjects count: %w", err)
	}

	dataOwnerSQL, args := ownerPredicate("user_id", scope, []any{params.Limit, params.Offset})
	dataQ := `
		SELECT id, name, color, time_budget, user_id, archived_at, tag_id
		FROM projects
		WHERE deleted_at IS NULL AND ` + archivedSQL + dataOwnerSQL + `
		ORDER BY name
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, dataQ, args...)
	if err != nil {
		return model.Page[model.Project]{}, fmt.Errorf("ListProjects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project

	for rows.Next() {
		var p model.Project
		var archivedAt *time.Time
		if err := rows.Scan(&p.Id, &p.Name, &p.Color, &p.TimeBudget, &p.UserId, &archivedAt, &p.TagId); err != nil {
			return model.Page[model.Project]{}, fmt.Errorf("ListProjects scan: %w", err)
		}
		p.Archived = archivedAt != nil
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return model.Page[model.Project]{}, fmt.Errorf("ListProjects rows: %w", err)
	}

	var tagErr error
	for i := range projects {
		projects[i].TagIds, tagErr = r.projectTagIds(ctx, r.db, projects[i].Id)
		if tagErr != nil {
			return model.Page[model.Project]{}, tagErr
		}
	}

	if projects == nil {
		projects = []model.Project{}
	}

	return model.Page[model.Project]{
		Data:       projects,
		TotalCount: totalCount,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

func (r *PostgresStore) CreateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	if project.Name == "" {
		return model.Project{}, fmt.Errorf("CreateProject: name must not be empty: %w", model.ErrInvalidArgument)
	}
	if project.Id == uuid.Nil {
		project.Id = uuid.New()
	}
	project.TagIds = utils.DedupeUUIDs(project.TagIds)

	var created model.Project
	err := r.withTx(ctx, func(q Querier) error {
		ok, err := r.tagsInScope(ctx, q, scope, project.TagIds, model.TagHolderProject, noAssociatedTags)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("CreateProject: %w", model.ErrInvalidReference)
		}

		// The project tag's name and color follow the project's.
		tagId := uuid.New()
		if _, err := q.Exec(ctx,
			`INSERT INTO tags (id, name, color, user_id) VALUES ($1, $2, $3, $4)`,
			tagId, project.Name, project.Color, scope.UserID(),
		); err != nil {
			return fmt.Errorf("CreateProject tag: %w", err)
		}

		const insert = `
			INSERT INTO projects (id, name, color, time_budget, user_id, tag_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, name, color, time_budget, user_id, tag_id`
		if err := q.QueryRow(ctx, insert, project.Id, project.Name, project.Color, project.TimeBudget, scope.UserID(), tagId).
			Scan(&created.Id, &created.Name, &created.Color, &created.TimeBudget, &created.UserId, &created.TagId); err != nil {
			return fmt.Errorf("CreateProject: %w", err)
		}
		return setLinkedTags(ctx, q, "project_tags", "project_id", created.Id, project.TagIds)
	})
	if err != nil {
		return model.Project{}, err
	}
	created.TagIds = project.TagIds
	return created, nil
}

func (r *PostgresStore) UpdateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	if project.Id == uuid.Nil {
		return model.Project{}, fmt.Errorf("UpdateProject: %w", errNilId)
	}
	if project.Name == "" {
		return model.Project{}, fmt.Errorf("UpdateProject: name must not be empty: %w", model.ErrInvalidArgument)
	}
	project.TagIds = utils.DedupeUUIDs(project.TagIds)

	var updated model.Project
	err := r.withTx(ctx, func(q Querier) error {
		ok, err := r.tagsInScope(ctx, q, scope, project.TagIds, model.TagHolderProject, func() ([]uuid.UUID, error) {
			return r.projectTagIds(ctx, q, project.Id)
		})
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("UpdateProject: %w", model.ErrInvalidReference)
		}

		ownerSQL, args := ownerPredicate("user_id", scope, []any{project.Id, project.Name, project.Color, project.TimeBudget, project.Archived})
		update := `
			UPDATE projects SET name = $2, color = $3, time_budget = $4,
				archived_at = CASE WHEN $5 THEN COALESCE(archived_at, now()) ELSE NULL END
			WHERE id = $1 AND deleted_at IS NULL AND ` + ownerSQL + `
			RETURNING id, name, color, time_budget, user_id, archived_at, tag_id`
		var archivedAt *time.Time
		err = q.QueryRow(ctx, update, args...).
			Scan(&updated.Id, &updated.Name, &updated.Color, &updated.TimeBudget, &updated.UserId, &archivedAt, &updated.TagId)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("UpdateProject %s: %w", project.Id, model.ErrNotFound)
		}
		if err != nil {
			return fmt.Errorf("UpdateProject: %w", err)
		}
		updated.Archived = archivedAt != nil
		if _, err := q.Exec(ctx, `UPDATE tags SET name = $2, color = $3 WHERE id = $1`, updated.TagId, updated.Name, updated.Color); err != nil {
			return fmt.Errorf("UpdateProject tag: %w", err)
		}
		return setLinkedTags(ctx, q, "project_tags", "project_id", updated.Id, project.TagIds)
	})
	if err != nil {
		return model.Project{}, err
	}
	updated.TagIds = project.TagIds
	return updated, nil
}

func (r *PostgresStore) DeleteProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteProject: %w", errNilId)
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	lockProject := `
		SELECT tag_id FROM projects
		WHERE id = $1 AND deleted_at IS NULL AND ` + ownerSQL + `
		FOR UPDATE`

	return r.withTx(ctx, func(tx Querier) error {
		var tagId uuid.UUID
		err := tx.QueryRow(ctx, lockProject, args...).Scan(&tagId)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("DeleteProject %s: %w", id, model.ErrNotFound)
		}
		if err != nil {
			return fmt.Errorf("DeleteProject: %w", err)
		}

		// Lock the project's effective tags before checking for attributed
		// time: a timespan or task attaching one of them share-locks it (see
		// tagsInScope), so it either commits first and shows up below, or
		// waits and then finds the project gone.
		rows, err := tx.Query(ctx, `
			SELECT id FROM tags
			WHERE id IN (SELECT tag_id FROM project_effective_tags WHERE project_id = $1)
				AND deleted_at IS NULL
			ORDER BY id FOR UPDATE`, id)
		if err != nil {
			return fmt.Errorf("DeleteProject lock tags: %w", err)
		}
		tagIds, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
		if err != nil {
			return fmt.Errorf("DeleteProject lock tags: %w", err)
		}

		// A project with time attributed to it can only be archived:
		// deleting its project tag would leave that time counting toward
		// nothing.
		timespanOwnerSQL, checkArgs := ownerPredicate("t.user_id", scope, []any{tagIds})
		var attributed bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM timespans t
				WHERE t.deleted_at IS NULL
					AND `+timespanOwnerSQL+`
					AND `+timespanHasEffectiveTagSQL("t.id", "$1")+`
			)`, checkArgs...).Scan(&attributed); err != nil {
			return fmt.Errorf("DeleteProject: %w", err)
		}
		if attributed {
			return fmt.Errorf("DeleteProject %s: has attributed time: %w", id, model.ErrConflict)
		}

		if _, err := tx.Exec(ctx, `UPDATE projects SET deleted_at = now() WHERE id = $1`, id); err != nil {
			return fmt.Errorf("DeleteProject: %w", err)
		}
		// The project tag goes with its project.
		if _, err := tx.Exec(ctx, `UPDATE tags SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, tagId); err != nil {
			return fmt.Errorf("DeleteProject tag: %w", err)
		}
		return nil
	})
}

// projectTagIds returns the live tag IDs linked to a project (see
// linkedTagIds).
func (r *PostgresStore) projectTagIds(ctx context.Context, q Querier, projectId uuid.UUID) ([]uuid.UUID, error) {
	ids, err := linkedTagIds(ctx, q, "project_tags", "project_id", []uuid.UUID{projectId})
	return ids[projectId], err
}

// tagsInScope reports whether every id refers to a live tag owned by scope
// that holder may carry (see model.TagHolder.MayCarry).
// An archived tag is only acceptable if it's already in the set alreadyAssociated
// returns (i.e. it was attached to this project/timespan before this call) -
// that keeps existing associations with a since-archived tag intact across
// unrelated edits, while still blocking a fresh attachment of an archived
// tag that a picker would never surface. alreadyAssociated is called at most
// once, and only if an archived tag is actually encountered, since the vast
// majority of calls reference no archived tags at all.
func (r *PostgresStore) tagsInScope(ctx context.Context, q Querier, scope model.OwnerScope, tagIds []uuid.UUID, holder model.TagHolder, alreadyAssociated func() ([]uuid.UUID, error)) (bool, error) {
	if len(tagIds) == 0 {
		return true, nil
	}

	carried := []string{}
	for _, kind := range holder.CarriedOwnerKinds() {
		carried = append(carried, string(kind))
	}
	ownerSQL, args := ownerPredicate("t.user_id", scope, []any{tagIds, carried})
	query := `
		SELECT t.id, ` + tagArchivedAtSQL + ` FROM ` + tagFromSQL + `
		WHERE t.id = ANY($1) AND t.deleted_at IS NULL
			AND (o.tag_id IS NULL OR o.kind = ANY($2))
			AND ` + ownerSQL
	// Share-lock the tags until the caller's transaction ends, so a
	// concurrent delete (of the tag, or of the owner of an owned tag)
	// either waits for this write or has already happened and fails the
	// deleted_at check above. Locking in id order avoids deadlocking with
	// DeleteTask, which locks its task tags the same way.
	query += ` ORDER BY t.id FOR SHARE OF t`
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("tagsInScope: %w", err)
	}

	// Fully drain and close rows before calling alreadyAssociated() below -
	// issuing another query against q (the same connection, when q is a
	// transaction) while this result set is still open would fail.
	var archivedIds []uuid.UUID
	found := 0
	for rows.Next() {
		var id uuid.UUID
		var archivedAt *time.Time
		if err := rows.Scan(&id, &archivedAt); err != nil {
			rows.Close()
			return false, fmt.Errorf("tagsInScope scan: %w", err)
		}
		found++
		if archivedAt != nil {
			archivedIds = append(archivedIds, id)
		}
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return false, fmt.Errorf("tagsInScope rows: %w", rowsErr)
	}

	if found != len(tagIds) {
		return false, nil
	}
	if len(archivedIds) == 0 {
		return true, nil
	}

	associated, err := alreadyAssociated()
	if err != nil {
		return false, err
	}
	allowedArchived := make(map[uuid.UUID]bool, len(associated))
	for _, aid := range associated {
		allowedArchived[aid] = true
	}
	for _, id := range archivedIds {
		if !allowedArchived[id] {
			return false, nil
		}
	}

	return true, nil
}
