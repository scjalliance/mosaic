package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Task represents a task in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type Task struct {
	// ID is the unique identifier for the task.
	ID int `json:"id,omitempty"`

	// TaskID is used in PUT requests to identify the task to update.
	TaskID int `json:"task_id,omitempty"`

	// TaskName is the display name of the task.
	TaskName string `json:"task_name"`

	// Note is the task description or note.
	Note string `json:"note,omitempty"`

	// ProjectID is the ID of the associated project.
	ProjectID int `json:"project_id"`

	// PhaseID is the ID of the associated phase.
	PhaseID int `json:"phase_id,omitempty"`

	// ActivityPhaseID is the ID of the associated activity phase.
	ActivityPhaseID int `json:"activity_phase_id,omitempty"`

	// TaskListID is the ID of the task list this task belongs to.
	TaskListID int `json:"task_list_id,omitempty"`

	// AssigneeIDs is the list of member IDs assigned to this task.
	AssigneeIDs []int `json:"assignee_ids,omitempty"`

	// UnassigneeIDs is the list of member IDs to unassign (used in PUT).
	UnassigneeIDs []int `json:"unassignee_ids,omitempty"`

	// Status is the current task status.
	Status string `json:"status,omitempty"`

	// EstimatedHours is the estimated hours for the task.
	EstimatedHours float64 `json:"estimated_hours,omitempty"`

	// ScheduleStart is the scheduled start date.
	ScheduleStart *Date `json:"schedule_start,omitempty"`

	// ScheduleEnd is the scheduled end date.
	ScheduleEnd *Date `json:"schedule_end,omitempty"`

	// CompletedAt is the datetime when the task was completed.
	CompletedAt string `json:"completed_at,omitempty"`

	// StandardWorkCategoryID is the work category ID.
	StandardWorkCategoryID int `json:"standard_work_category_id,omitempty"`

	// Position is the display position of the task (used in PUT).
	Position int `json:"position,omitempty"`
}

// TaskFilter contains filter parameters for listing tasks.
type TaskFilter struct {
	// MemberID filters tasks assigned to a specific member.
	MemberID int

	// ProjectID filters tasks by project.
	ProjectID int
}

// toQuery converts TaskFilter into URL query parameter values.
func (f TaskFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.MemberID > 0 {
		q.Set("member_id", strconv.Itoa(f.MemberID))
	}
	if f.ProjectID > 0 {
		q.Set("project_id", strconv.Itoa(f.ProjectID))
	}
	return q
}

// ListTasks retrieves a list of tasks matching the given filter.
// Uses GET /api/{team_id}/task.
func (c *Client) ListTasks(ctx context.Context, filter TaskFilter) ([]Task, error) {
	var resp []Task
	path := c.apiPath("task")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	return resp, nil
}

// CreateTask creates a new task and returns the created task.
// Uses POST /api/{team_id}/task.
func (c *Client) CreateTask(ctx context.Context, task *Task) (*Task, error) {
	var created Task
	path := c.apiPath("task")
	if err := c.post(ctx, path, task, &created); err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}
	return &created, nil
}

// UpdateTask updates an existing task and returns the updated task.
// The task must have TaskID set. Uses PUT /api/{team_id}/task (body, not path param).
func (c *Client) UpdateTask(ctx context.Context, task *Task) (*Task, error) {
	var updated Task
	path := c.apiPath("task")
	if err := c.put(ctx, path, task, &updated); err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}
	return &updated, nil
}

// DeleteTask deletes a task by ID.
// Uses DELETE /api/{team_id}/task/{task_id}.
func (c *Client) DeleteTask(ctx context.Context, taskID int) error {
	path := c.apiPath("task", strconv.Itoa(taskID))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting task %d: %w", taskID, err)
	}
	return nil
}

// TaskList represents a task list (grouping of tasks) in Mosaic.
type TaskList struct {
	// ID is the unique identifier for the task list.
	ID int `json:"id,omitempty"`

	// Name is the display name of the task list.
	Name string `json:"name,omitempty"`

	// ProjectID is the ID of the associated project.
	ProjectID int `json:"project_id,omitempty"`

	// Index is the display position of the task list.
	Index int `json:"index,omitempty"`
}

// TaskListFilter contains filter parameters for listing task lists.
type TaskListFilter struct {
	// ProjectID filters task lists by project.
	ProjectID int

	// ProjectIDs filters task lists by multiple projects (index endpoint).
	ProjectIDs []int

	// TaskListIDs filters to specific task list IDs (index endpoint).
	TaskListIDs []int
}

// toQuery converts TaskListFilter into URL query parameter values.
func (f TaskListFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.ProjectID > 0 {
		q.Set("project_id", strconv.Itoa(f.ProjectID))
	}
	for _, id := range f.ProjectIDs {
		q.Add("project_ids[]", strconv.Itoa(id))
	}
	for _, id := range f.TaskListIDs {
		q.Add("task_list_ids[]", strconv.Itoa(id))
	}
	return q
}

// ListTaskLists retrieves task lists for a project.
// Uses GET /api/{team_id}/task_list.
func (c *Client) ListTaskLists(ctx context.Context, filter TaskListFilter) ([]TaskList, error) {
	var resp []TaskList
	path := c.apiPath("task_list")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing task lists: %w", err)
	}
	return resp, nil
}

// ListAllTaskLists retrieves all task lists using the index endpoint.
// Uses GET /api/{team_id}/task_list/index.
func (c *Client) ListAllTaskLists(ctx context.Context, filter TaskListFilter) ([]TaskList, error) {
	var resp []TaskList
	path := c.apiPath("task_list", "index")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing all task lists: %w", err)
	}
	return resp, nil
}

// CreateTaskList creates a new task list and returns the created task list.
// Uses POST /api/{team_id}/task_list.
func (c *Client) CreateTaskList(ctx context.Context, taskList *TaskList) (*TaskList, error) {
	var created TaskList
	path := c.apiPath("task_list")
	if err := c.post(ctx, path, taskList, &created); err != nil {
		return nil, fmt.Errorf("creating task list: %w", err)
	}
	return &created, nil
}

// UpdateTaskList updates an existing task list.
// Uses PUT /api/{team_id}/task_list/{task_list_id}.
func (c *Client) UpdateTaskList(ctx context.Context, taskListID int, taskList *TaskList) (*TaskList, error) {
	var updated TaskList
	path := c.apiPath("task_list", strconv.Itoa(taskListID))
	if err := c.put(ctx, path, taskList, &updated); err != nil {
		return nil, fmt.Errorf("updating task list %d: %w", taskListID, err)
	}
	return &updated, nil
}

// DeleteTaskList deletes a task list by ID.
// Uses DELETE /api/{team_id}/task_list/{task_list_id}.
func (c *Client) DeleteTaskList(ctx context.Context, taskListID int) error {
	path := c.apiPath("task_list", strconv.Itoa(taskListID))
	if err := c.deleteNoBody(ctx, path); err != nil {
		return fmt.Errorf("deleting task list %d: %w", taskListID, err)
	}
	return nil
}
