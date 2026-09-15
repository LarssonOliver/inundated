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
		return model.Project{}, fmt.Errorf("GetProject: id: %w", model.ErrInvalidArgument)
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	q := `
		SELECT id, name, color, time_budget, user_id, archived_at
		FROM projects
		WHERE id = $1 AND deleted_at IS NULL AND ` + ownerSQL

	var p model.Project
	var archivedAt *time.Time
	err := r.db.QueryRow(ctx, q, args...).Scan(&p.Id, &p.Name, &p.Color, &p.TimeBudget, &p.UserId, &archivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Project{}, fmt.Errorf("GetProject %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("GetProject: %w", err)
	}
	p.Archived = archivedAt != nil

	p.TagIds, err = r.projectTagIds(ctx, id)
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
		SELECT id, name, color, time_budget, user_id, archived_at
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
		if err := rows.Scan(&p.Id, &p.Name, &p.Color, &p.TimeBudget, &p.UserId, &archivedAt); err != nil {
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
		projects[i].TagIds, tagErr = r.projectTagIds(ctx, projects[i].Id)
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
		ok, err := r.tagsInScope(ctx, q, scope, project.TagIds, nil)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("CreateProject: %w", model.ErrInvalidReference)
		}

		const insert = `
			INSERT INTO projects (id, name, color, time_budget, user_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, color, time_budget, user_id`
		if err := q.QueryRow(ctx, insert, project.Id, project.Name, project.Color, project.TimeBudget, scope.UserID()).
			Scan(&created.Id, &created.Name, &created.Color, &created.TimeBudget, &created.UserId); err != nil {
			return fmt.Errorf("CreateProject: %w", err)
		}
		return r.setProjectTags(ctx, q, created.Id, project.TagIds)
	})
	if err != nil {
		return model.Project{}, err
	}
	created.TagIds = project.TagIds
	return created, nil
}

func (r *PostgresStore) UpdateProject(ctx context.Context, scope model.OwnerScope, project model.Project) (model.Project, error) {
	if project.Id == uuid.Nil {
		return model.Project{}, fmt.Errorf("UpdateProject: id: %w", model.ErrInvalidArgument)
	}
	if project.Name == "" {
		return model.Project{}, fmt.Errorf("UpdateProject: name must not be empty: %w", model.ErrInvalidArgument)
	}
	project.TagIds = utils.DedupeUUIDs(project.TagIds)

	existingTagIds, err := r.projectTagIds(ctx, project.Id)
	if err != nil {
		return model.Project{}, err
	}

	var updated model.Project
	err = r.withTx(ctx, func(q Querier) error {
		ok, err := r.tagsInScope(ctx, q, scope, project.TagIds, existingTagIds)
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
			RETURNING id, name, color, time_budget, user_id, archived_at`
		var archivedAt *time.Time
		err = q.QueryRow(ctx, update, args...).
			Scan(&updated.Id, &updated.Name, &updated.Color, &updated.TimeBudget, &updated.UserId, &archivedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("UpdateProject %s: %w", project.Id, model.ErrNotFound)
		}
		if err != nil {
			return fmt.Errorf("UpdateProject: %w", err)
		}
		updated.Archived = archivedAt != nil
		return r.setProjectTags(ctx, q, updated.Id, project.TagIds)
	})
	if err != nil {
		return model.Project{}, err
	}
	updated.TagIds = project.TagIds
	return updated, nil
}

func (r *PostgresStore) DeleteProject(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteProject: id: %w", model.ErrInvalidArgument)
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	q := `
		UPDATE projects
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND ` + ownerSQL

	res, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("DeleteProject: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeleteProject %s: %w", id, model.ErrNotFound)
	}
	return nil
}

// projectTagIds returns all tag IDs linked to a project.
func (r *PostgresStore) projectTagIds(ctx context.Context, projectId uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT tag_id 
		FROM project_tags 
		WHERE project_id = $1 
		ORDER BY tag_id`

	rows, err := r.db.Query(ctx, q, projectId)
	if err != nil {
		return nil, fmt.Errorf("projectTagIds: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("projectTagIds scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// tagsInScope reports whether every id refers to a live tag owned by scope.
// An archived tag is only acceptable if it's already in alreadyAssociated
// (i.e. it was attached to this project/timespan before this call) - that
// keeps existing associations with a since-archived tag intact across
// unrelated edits, while still blocking a fresh attachment of an archived
// tag that a picker would never surface.
func (r *PostgresStore) tagsInScope(ctx context.Context, q Querier, scope model.OwnerScope, tagIds []uuid.UUID, alreadyAssociated []uuid.UUID) (bool, error) {
	if len(tagIds) == 0 {
		return true, nil
	}

	allowedArchived := make(map[uuid.UUID]bool, len(alreadyAssociated))
	for _, id := range alreadyAssociated {
		allowedArchived[id] = true
	}

	ownerSQL, args := ownerPredicate("user_id", scope, []any{tagIds})
	query := `
		SELECT id, archived_at FROM tags
		WHERE id = ANY($1) AND deleted_at IS NULL AND ` + ownerSQL
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("tagsInScope: %w", err)
	}
	defer rows.Close()

	found := 0
	for rows.Next() {
		var id uuid.UUID
		var archivedAt *time.Time
		if err := rows.Scan(&id, &archivedAt); err != nil {
			return false, fmt.Errorf("tagsInScope scan: %w", err)
		}
		found++
		if archivedAt != nil && !allowedArchived[id] {
			return false, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("tagsInScope rows: %w", err)
	}

	return found == len(tagIds), nil
}

// setProjectTags replaces all tag associations for a project.
func (r *PostgresStore) setProjectTags(ctx context.Context, q Querier, projectId uuid.UUID, tagIds []uuid.UUID) error {
	if _, err := q.Exec(ctx, `DELETE FROM project_tags WHERE project_id = $1`, projectId); err != nil {
		return fmt.Errorf("setProjectTags delete: %w", err)
	}
	for _, tagId := range tagIds {
		if _, err := q.Exec(ctx,
			`INSERT INTO project_tags (project_id, tag_id) VALUES ($1, $2)`,
			projectId, tagId,
		); err != nil {
			return fmt.Errorf("setProjectTags insert: %w: %w", model.ErrInvalidReference, err)
		}
	}
	return nil
}
