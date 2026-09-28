package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/utils"
)

const taskColumnsSQL = `id, name, tag_id, parent_id, due_date, estimate, rank, close_reason, closed_at, user_id`

func scanTask(row pgx.Row) (model.Task, error) {
	var t model.Task
	var reason *string
	if err := row.Scan(&t.Id, &t.Name, &t.TagId, &t.ParentId, &t.DueDate, &t.Estimate, &t.Rank, &reason, &t.ClosedAt, &t.UserId); err != nil {
		return model.Task{}, err
	}
	if reason != nil {
		r := model.CloseReason(*reason)
		t.CloseReason = &r
	}
	return t, nil
}

func (r *PostgresStore) GetTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID) (model.Task, error) {
	if id == uuid.Nil {
		return model.Task{}, fmt.Errorf("GetTask: id: %w", model.ErrInvalidArgument)
	}
	return r.getTask(ctx, r.db, scope, id, false)
}

// getTask reads one live task in scope. forUpdate locks its row for the
// rest of the caller's transaction.
func (r *PostgresStore) getTask(ctx context.Context, q Querier, scope model.OwnerScope, id uuid.UUID, forUpdate bool) (model.Task, error) {
	ownerSQL, args := ownerPredicate("user_id", scope, []any{id})
	query := `
		SELECT ` + taskColumnsSQL + `
		FROM tasks
		WHERE id = $1 AND deleted_at IS NULL AND ` + ownerSQL
	if forUpdate {
		query += ` FOR UPDATE`
	}

	t, err := scanTask(q.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Task{}, fmt.Errorf("GetTask %s: %w", id, model.ErrNotFound)
	}
	if err != nil {
		return model.Task{}, fmt.Errorf("GetTask: %w", err)
	}

	tagIds, err := r.taskTagIds(ctx, q, []uuid.UUID{id})
	if err != nil {
		return model.Task{}, err
	}
	t.TagIds = tagIds[id]
	return t, nil
}

func (r *PostgresStore) ListTasks(ctx context.Context, scope model.OwnerScope, params model.TaskListParams) (model.Page[model.Task], error) {
	countBuilder := newSQLConditionBuilder(nil)
	countClosedSQL := addTaskFilters(countBuilder, params)
	countOwnerSQL, countArgs := ownerPredicate("user_id", scope, countBuilder.Args())
	countQ := `
		SELECT COUNT(*)
		FROM tasks
		WHERE deleted_at IS NULL` + countClosedSQL + countBuilder.SQL() + ` AND ` + countOwnerSQL

	var totalCount int
	if err := r.db.QueryRow(ctx, countQ, countArgs...).Scan(&totalCount); err != nil {
		return model.Page[model.Task]{}, fmt.Errorf("ListTasks count: %w", err)
	}

	dataBuilder := newSQLConditionBuilder([]any{params.Limit, params.Offset})
	dataClosedSQL := addTaskFilters(dataBuilder, params)
	dataOwnerSQL, args := ownerPredicate("user_id", scope, dataBuilder.Args())
	dataQ := `
		SELECT ` + taskColumnsSQL + `
		FROM tasks
		WHERE deleted_at IS NULL` + dataClosedSQL + dataBuilder.SQL() + ` AND ` + dataOwnerSQL + `
		ORDER BY parent_id NULLS FIRST, rank, id
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, dataQ, args...)
	if err != nil {
		return model.Page[model.Task]{}, fmt.Errorf("ListTasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return model.Page[model.Task]{}, fmt.Errorf("ListTasks scan: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return model.Page[model.Task]{}, fmt.Errorf("ListTasks rows: %w", err)
	}

	ids := make([]uuid.UUID, len(tasks))
	for i, t := range tasks {
		ids[i] = t.Id
	}
	tagIds, err := r.taskTagIds(ctx, r.db, ids)
	if err != nil {
		return model.Page[model.Task]{}, err
	}
	for i := range tasks {
		tasks[i].TagIds = tagIds[tasks[i].Id]
	}
	if tasks == nil {
		tasks = []model.Task{}
	}

	return model.Page[model.Task]{
		Data:       tasks,
		TotalCount: totalCount,
		Limit:      params.Limit,
		Offset:     params.Offset,
	}, nil
}

// addTaskFilters adds ListTasks's optional filters to b and returns the
// closed-task condition, which binds nothing.
func addTaskFilters(b *sqlConditionBuilder, params model.TaskListParams) string {
	if params.ParentId != nil {
		b.add("parent_id =", *params.ParentId)
	}
	if params.TagId != nil {
		// The placeholder sits inside a subquery, which b.add can't express.
		b.args = append(b.args, *params.TagId)
		fmt.Fprintf(&b.sql, " AND id IN (SELECT kt.task_id FROM task_tags kt JOIN tags t ON t.id = kt.tag_id AND t.deleted_at IS NULL WHERE kt.tag_id = $%d)", len(b.args))
	}
	if params.DueFrom != nil {
		b.add("due_date >=", *params.DueFrom)
	}
	if params.DueTo != nil {
		b.add("due_date <=", *params.DueTo)
	}
	if params.IncludeClosed {
		return ""
	}
	return " AND closed_at IS NULL"
}

func (r *PostgresStore) CreateTask(ctx context.Context, scope model.OwnerScope, task model.Task) (model.Task, error) {
	if task.Name == "" {
		return model.Task{}, fmt.Errorf("CreateTask: name must not be empty: %w", model.ErrInvalidArgument)
	}
	tagIds := utils.DedupeUUIDs(task.TagIds)

	var created model.Task
	err := r.withTx(ctx, func(q Querier) error {
		if err := lockTaskTree(ctx, q, scope); err != nil {
			return err
		}
		if err := r.checkTaskTags(ctx, q, scope, tagIds, noAssociatedTags); err != nil {
			return fmt.Errorf("CreateTask: %w", err)
		}

		if task.ParentId != nil {
			parent, err := r.getTask(ctx, q, scope, *task.ParentId, true)
			if errors.Is(err, model.ErrNotFound) {
				return fmt.Errorf("CreateTask: parent: %w", model.ErrInvalidReference)
			}
			if err != nil {
				return err
			}
			if parent.Closed() {
				return fmt.Errorf("CreateTask: parent is closed: %w", model.ErrInvalidArgument)
			}
		}

		siblings, err := r.siblingTasks(ctx, q, scope, task.ParentId, uuid.Nil)
		if err != nil {
			return err
		}
		last := ""
		if len(siblings) > 0 {
			last = siblings[len(siblings)-1].rank
		}
		rank, err := utils.RankBetween(last, "")
		if err != nil || len(rank) > utils.MaxRankLength {
			if err := r.respaceTasks(ctx, q, siblings); err != nil {
				return err
			}
			rank, err = utils.RankBetween(utils.EvenRanks(len(siblings))[len(siblings)-1], "")
			if err != nil {
				return err
			}
		}

		tagId := uuid.New()
		if _, err := q.Exec(ctx,
			`INSERT INTO tags (id, name, color, user_id) VALUES ($1, $2, $3, $4)`,
			tagId, task.Name, model.DefaultTaskTagColor, scope.UserID(),
		); err != nil {
			return fmt.Errorf("CreateTask tag: %w", err)
		}

		const insert = `
			INSERT INTO tasks (id, name, tag_id, parent_id, due_date, estimate, rank, user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING ` + taskColumnsSQL
		created, err = scanTask(q.QueryRow(ctx, insert,
			uuid.New(), task.Name, tagId, task.ParentId, task.DueDate, task.Estimate, rank, scope.UserID()))
		if err != nil {
			return fmt.Errorf("CreateTask: %w", err)
		}
		return r.setTaskTags(ctx, q, created.Id, tagIds)
	})
	if err != nil {
		return model.Task{}, err
	}
	created.TagIds = tagIds
	return created, nil
}

func (r *PostgresStore) UpdateTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID, patch model.TaskPatch) (model.Task, error) {
	if id == uuid.Nil {
		return model.Task{}, fmt.Errorf("UpdateTask: id: %w", model.ErrInvalidArgument)
	}

	var updated model.Task
	var tagIds []uuid.UUID
	err := r.withTx(ctx, func(q Querier) error {
		if err := lockTaskTree(ctx, q, scope); err != nil {
			return err
		}
		existing, err := r.getTask(ctx, q, scope, id, true)
		if err != nil {
			return err
		}
		task := patch.Apply(existing)
		if task.Name == "" {
			return fmt.Errorf("UpdateTask: name must not be empty: %w", model.ErrInvalidArgument)
		}
		if task.CloseReason != nil && !task.CloseReason.Valid() {
			return fmt.Errorf("UpdateTask: close reason: %w", model.ErrInvalidArgument)
		}
		tagIds = utils.DedupeUUIDs(task.TagIds)
		if err := r.checkTaskTags(ctx, q, scope, tagIds, func() ([]uuid.UUID, error) {
			return existing.TagIds, nil
		}); err != nil {
			return fmt.Errorf("UpdateTask: %w", err)
		}

		var reason *string
		if task.CloseReason != nil {
			s := string(*task.CloseReason)
			reason = &s
		}
		const update = `
			UPDATE tasks SET name = $2, due_date = $3, estimate = $4, close_reason = $5,
				closed_at = CASE WHEN $5::text IS NULL THEN NULL ELSE COALESCE(closed_at, now()) END
			WHERE id = $1
			RETURNING ` + taskColumnsSQL
		updated, err = scanTask(q.QueryRow(ctx, update, task.Id, task.Name, task.DueDate, task.Estimate, reason))
		if err != nil {
			return fmt.Errorf("UpdateTask: %w", err)
		}

		switch {
		case !existing.Closed() && updated.Closed():
			// Close the open part of the subtree, stamped like the task.
			const closeDescendants = `
				WITH RECURSIVE sub AS (
					SELECT id FROM tasks WHERE parent_id = $1 AND deleted_at IS NULL
					UNION
					SELECT c.id FROM tasks c JOIN sub ON c.parent_id = sub.id
					WHERE c.deleted_at IS NULL
				)
				UPDATE tasks SET close_reason = $2, closed_at = $3
				WHERE id IN (SELECT id FROM sub) AND closed_at IS NULL`
			if _, err := q.Exec(ctx, closeDescendants, updated.Id, reason, updated.ClosedAt); err != nil {
				return fmt.Errorf("UpdateTask close subtasks: %w", err)
			}
		case existing.Closed() && !updated.Closed():
			const reopenAncestors = `
				WITH RECURSIVE anc AS (
					SELECT parent_id AS id FROM tasks WHERE id = $1
					UNION
					SELECT t.parent_id FROM tasks t JOIN anc ON t.id = anc.id
					WHERE t.parent_id IS NOT NULL
				)
				UPDATE tasks SET close_reason = NULL, closed_at = NULL
				WHERE id IN (SELECT id FROM anc WHERE id IS NOT NULL)`
			if _, err := q.Exec(ctx, reopenAncestors, updated.Id); err != nil {
				return fmt.Errorf("UpdateTask reopen parents: %w", err)
			}
		}

		if _, err := q.Exec(ctx, `UPDATE tags SET name = $2 WHERE id = $1`, updated.TagId, updated.Name); err != nil {
			return fmt.Errorf("UpdateTask tag: %w", err)
		}
		return r.setTaskTags(ctx, q, updated.Id, tagIds)
	})
	if err != nil {
		return model.Task{}, err
	}
	updated.TagIds = tagIds
	return updated, nil
}

func (r *PostgresStore) MoveTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID, parentId *uuid.UUID, afterId *uuid.UUID) (model.Task, error) {
	if id == uuid.Nil {
		return model.Task{}, fmt.Errorf("MoveTask: id: %w", model.ErrInvalidArgument)
	}

	var moved model.Task
	err := r.withTx(ctx, func(q Querier) error {
		if err := lockTaskTree(ctx, q, scope); err != nil {
			return err
		}
		task, err := r.getTask(ctx, q, scope, id, true)
		if err != nil {
			return err
		}

		if parentId != nil {
			parent, err := r.getTask(ctx, q, scope, *parentId, true)
			if errors.Is(err, model.ErrNotFound) {
				return fmt.Errorf("MoveTask: parent: %w", model.ErrInvalidArgument)
			}
			if err != nil {
				return err
			}
			if parent.Closed() && !task.Closed() {
				return fmt.Errorf("MoveTask: parent is closed: %w", model.ErrInvalidArgument)
			}

			const isDescendant = `
				WITH RECURSIVE anc AS (
					SELECT id, parent_id FROM tasks WHERE id = $1
					UNION
					SELECT t.id, t.parent_id FROM tasks t JOIN anc ON t.id = anc.parent_id
				)
				SELECT EXISTS (SELECT 1 FROM anc WHERE id = $2)`
			var cycle bool
			if err := q.QueryRow(ctx, isDescendant, *parentId, id).Scan(&cycle); err != nil {
				return fmt.Errorf("MoveTask: %w", err)
			}
			if cycle {
				return fmt.Errorf("MoveTask: parent is the task or below it: %w", model.ErrInvalidArgument)
			}
		}

		siblings, err := r.siblingTasks(ctx, q, scope, parentId, id)
		if err != nil {
			return err
		}
		pos := 0
		if afterId != nil {
			pos = -1
			for i, s := range siblings {
				if s.id == *afterId {
					pos = i + 1
					break
				}
			}
			if pos == -1 {
				return fmt.Errorf("MoveTask: afterTaskId is not a sibling: %w", model.ErrInvalidArgument)
			}
		}

		lo, hi := "", ""
		if pos > 0 {
			lo = siblings[pos-1].rank
		}
		if pos < len(siblings) {
			hi = siblings[pos].rank
		}
		rank, rankErr := utils.RankBetween(lo, hi)

		if _, err := q.Exec(ctx, `UPDATE tasks SET parent_id = $2 WHERE id = $1`, id, parentId); err != nil {
			return fmt.Errorf("MoveTask: %w", err)
		}
		if rankErr != nil || len(rank) > utils.MaxRankLength {
			ordered := make([]rankedTask, 0, len(siblings)+1)
			ordered = append(ordered, siblings[:pos]...)
			ordered = append(ordered, rankedTask{id: id})
			ordered = append(ordered, siblings[pos:]...)
			if err := r.respaceTasks(ctx, q, ordered); err != nil {
				return err
			}
		} else if _, err := q.Exec(ctx, `UPDATE tasks SET rank = $2 WHERE id = $1`, id, rank); err != nil {
			return fmt.Errorf("MoveTask: %w", err)
		}

		moved, err = r.getTask(ctx, q, scope, id, false)
		return err
	})
	if err != nil {
		return model.Task{}, err
	}
	return moved, nil
}

func (r *PostgresStore) DeleteTask(ctx context.Context, scope model.OwnerScope, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DeleteTask: id: %w", model.ErrInvalidArgument)
	}

	return r.withTx(ctx, func(q Querier) error {
		if err := lockTaskTree(ctx, q, scope); err != nil {
			return err
		}
		if _, err := r.getTask(ctx, q, scope, id, true); err != nil {
			return err
		}

		const subtree = `
			WITH RECURSIVE sub AS (
				SELECT id, tag_id FROM tasks WHERE id = $1
				UNION
				SELECT c.id, c.tag_id FROM tasks c JOIN sub ON c.parent_id = sub.id
				WHERE c.deleted_at IS NULL
			)`

		var logged bool
		if err := q.QueryRow(ctx, subtree+`
			SELECT EXISTS (
				SELECT 1 FROM timespan_tags tt
				JOIN timespans ts ON ts.id = tt.timespan_id AND ts.deleted_at IS NULL
				WHERE tt.tag_id IN (SELECT tag_id FROM sub)
			)`, id).Scan(&logged); err != nil {
			return fmt.Errorf("DeleteTask: %w", err)
		}
		if logged {
			return fmt.Errorf("DeleteTask %s: has logged time: %w", id, model.ErrConflict)
		}

		if _, err := q.Exec(ctx, subtree+`,
			deleted_tasks AS (
				UPDATE tasks SET deleted_at = now() WHERE id IN (SELECT id FROM sub)
			),
			deleted_tags AS (
				UPDATE tags SET deleted_at = now() WHERE id IN (SELECT tag_id FROM sub)
			)
			DELETE FROM project_tags WHERE tag_id IN (SELECT tag_id FROM sub)`, id); err != nil {
			return fmt.Errorf("DeleteTask: %w", err)
		}
		return nil
	})
}

// checkTaskTags reports model.ErrInvalidReference unless every id is a
// usable regular tag in scope (see tagsInScope); a task's own tags can
// never include task tags.
func (r *PostgresStore) checkTaskTags(ctx context.Context, q Querier, scope model.OwnerScope, tagIds []uuid.UUID, alreadyAssociated func() ([]uuid.UUID, error)) error {
	ok, err := r.tagsInScope(ctx, q, scope, tagIds, true, alreadyAssociated)
	if err != nil {
		return err
	}
	if !ok {
		return model.ErrInvalidReference
	}
	return nil
}

// lockTaskTree serialises changes to scope's task tree until the
// transaction ends. Row locks alone can't: a move's cycle check reads the
// whole ancestor chain, a create's rank reads every sibling, and closing
// or deleting walks a whole subtree, so concurrent edits that lock
// disjoint rows could still leave a parent cycle, duplicate sibling ranks
// or live tasks under deleted ones. Callers take it before any row lock.
func lockTaskTree(ctx context.Context, q Querier, scope model.OwnerScope) error {
	key := ""
	if id := scope.UserID(); id != nil {
		key = id.String()
	}
	if _, err := q.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('tasks:' || $1, 0))`, key); err != nil {
		return fmt.Errorf("lockTaskTree: %w", err)
	}
	return nil
}

type rankedTask struct {
	id   uuid.UUID
	rank string
}

// siblingTasks returns the live tasks under parentId (nil for the top
// level) in rank order, leaving out exclude.
func (r *PostgresStore) siblingTasks(ctx context.Context, q Querier, scope model.OwnerScope, parentId *uuid.UUID, exclude uuid.UUID) ([]rankedTask, error) {
	ownerSQL, args := ownerPredicate("user_id", scope, []any{parentId, exclude})
	query := `
		SELECT id, rank FROM tasks
		WHERE parent_id IS NOT DISTINCT FROM $1 AND id <> $2 AND deleted_at IS NULL AND ` + ownerSQL + `
		ORDER BY rank`
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("siblingTasks: %w", err)
	}
	defer rows.Close()

	var out []rankedTask
	for rows.Next() {
		var s rankedTask
		if err := rows.Scan(&s.id, &s.rank); err != nil {
			return nil, fmt.Errorf("siblingTasks scan: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// respaceTasks gives tasks fresh, evenly spread ranks in the order given.
func (r *PostgresStore) respaceTasks(ctx context.Context, q Querier, tasks []rankedTask) error {
	for i, rank := range utils.EvenRanks(len(tasks)) {
		if _, err := q.Exec(ctx, `UPDATE tasks SET rank = $2 WHERE id = $1`, tasks[i].id, rank); err != nil {
			return fmt.Errorf("respaceTasks: %w", err)
		}
	}
	return nil
}

// taskTagIds returns the live regular tag IDs on each of taskIds, keyed by
// task. Links to deleted tags are kept (tag deletion is soft) but never
// read back.
func (r *PostgresStore) taskTagIds(ctx context.Context, q Querier, taskIds []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	out := make(map[uuid.UUID][]uuid.UUID, len(taskIds))
	if len(taskIds) == 0 {
		return out, nil
	}
	const query = `
		SELECT kt.task_id, kt.tag_id
		FROM task_tags kt
		JOIN tags t ON t.id = kt.tag_id AND t.deleted_at IS NULL
		WHERE kt.task_id = ANY($1)
		ORDER BY kt.task_id, kt.tag_id`
	rows, err := q.Query(ctx, query, taskIds)
	if err != nil {
		return nil, fmt.Errorf("taskTagIds: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskId, tagId uuid.UUID
		if err := rows.Scan(&taskId, &tagId); err != nil {
			return nil, fmt.Errorf("taskTagIds scan: %w", err)
		}
		out[taskId] = append(out[taskId], tagId)
	}
	return out, rows.Err()
}

// setTaskTags replaces a task's regular tags.
func (r *PostgresStore) setTaskTags(ctx context.Context, q Querier, taskId uuid.UUID, tagIds []uuid.UUID) error {
	if _, err := q.Exec(ctx, `DELETE FROM task_tags WHERE task_id = $1`, taskId); err != nil {
		return fmt.Errorf("setTaskTags delete: %w", err)
	}
	for _, tagId := range tagIds {
		if _, err := q.Exec(ctx,
			`INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2)`,
			taskId, tagId,
		); err != nil {
			return fmt.Errorf("setTaskTags insert: %w: %w", model.ErrInvalidReference, err)
		}
	}
	return nil
}
