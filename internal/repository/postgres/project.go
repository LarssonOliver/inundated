package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/larssonoliver/inundated/internal/model"
)

func (r *PostgresStore) GetProject(ctx context.Context, id uuid.UUID) (model.Project, error) {
	if id == uuid.Nil {
		return model.Project{}, fmt.Errorf("GetProject: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		SELECT id, name, color, time_budget 
		FROM projects 
		WHERE id = $1 AND deleted_at IS NULL`

	var p model.Project
	err := r.db.QueryRow(ctx, q, id).Scan(&p.Id, &p.Name, &p.Color, &p.TimeBudget)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Project{}, fmt.Errorf("GetProject %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("GetProject: %w", err)
	}

	p.TagIds, err = r.projectTagIds(ctx, id)
	if err != nil {
		return model.Project{}, err
	}
	return p, nil
}

func (r *PostgresStore) ListProjects(ctx context.Context, params model.PaginationParams) (model.Page[model.Project], error) {
	const q = `
		SELECT id, name, color, time_budget, COUNT(*) OVER() AS total_count
		FROM projects 
		WHERE deleted_at IS NULL
		ORDER BY name
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, q, params.Limit, params.Offset)
	if err != nil {
		return model.Page[model.Project]{}, fmt.Errorf("ListProjects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	totalCount := 0
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.Id, &p.Name, &p.Color, &p.TimeBudget, &totalCount); err != nil {
			return model.Page[model.Project]{}, fmt.Errorf("ListProjects scan: %w", err)
		}
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

func (r *PostgresStore) CreateProject(ctx context.Context, project model.Project) (model.Project, error) {
	if project.Name == "" {
		return model.Project{}, fmt.Errorf("CreateProject: name must not be empty: %w", model.ErrInvalidArgument)
	}
	if project.Id == uuid.Nil {
		project.Id = uuid.New()
	}

	const q = `
		INSERT INTO projects (id, name, color, time_budget)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, color, time_budget`

	var created model.Project
	err := r.db.QueryRow(ctx, q, project.Id, project.Name, project.Color, project.TimeBudget).
		Scan(&created.Id, &created.Name, &created.Color, &created.TimeBudget)
	if err != nil {
		return model.Project{}, fmt.Errorf("CreateProject: %w", err)
	}

	if err := r.setProjectTags(ctx, created.Id, project.TagIds); err != nil {
		return model.Project{}, err
	}
	created.TagIds = project.TagIds
	return created, nil
}

func (r *PostgresStore) UpdateProject(ctx context.Context, project model.Project) (model.Project, error) {
	if project.Id == uuid.Nil {
		return model.Project{}, fmt.Errorf("UpdateProject: id: %w", model.ErrInvalidArgument)
	}
	if project.Name == "" {
		return model.Project{}, fmt.Errorf("UpdateProject: name must not be empty: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE projects SET name = $2, color = $3, time_budget = $4
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, color, time_budget`

	var updated model.Project
	err := r.db.QueryRow(ctx, q, project.Id, project.Name, project.Color, project.TimeBudget).
		Scan(&updated.Id, &updated.Name, &updated.Color, &updated.TimeBudget)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Project{}, fmt.Errorf("UpdateProject %s: %w", project.Id, model.ErrNotFound)
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("UpdateProject: %w", err)
	}

	if err := r.setProjectTags(ctx, updated.Id, project.TagIds); err != nil {
		return model.Project{}, err
	}
	updated.TagIds = project.TagIds
	return updated, nil
}

func (r *PostgresStore) DeleteProject(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteProject: id: %w", model.ErrInvalidArgument)
	}

	const q = `
		UPDATE projects
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL`

	res, err := r.db.Exec(ctx, q, id)
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

// setProjectTags replaces all tag associations for a project.
func (r *PostgresStore) setProjectTags(ctx context.Context, projectId uuid.UUID, tagIds []uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM project_tags WHERE project_id = $1`, projectId); err != nil {
		return fmt.Errorf("setProjectTags delete: %w", err)
	}
	for _, tagId := range tagIds {
		if _, err := r.db.Exec(ctx,
			`INSERT INTO project_tags (project_id, tag_id) VALUES ($1, $2)`,
			projectId, tagId,
		); err != nil {
			return fmt.Errorf("setProjectTags insert: %w", model.ErrInvalidReference)
		}
	}
	return nil
}
