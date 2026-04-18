package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	RecurrenceType   *string           `json:"recurrence_type,omitempty"`
    RecurrenceValue  *string           `json:"recurrence_value,omitempty"`
    RecurrenceEndDate *string          `json:"recurrence_end_date,omitempty"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	RecurrenceType   *string           `json:"recurrence_type,omitempty"`
    RecurrenceValue  *string           `json:"recurrence_value,omitempty"`
    RecurrenceEndDate *time.Time        `json:"recurrence_end_date,omitempty"`
    ParentTaskID     *string            `json:"parent_task_id,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		RecurrenceType:    (*string)(task.RecurrenceType),
        RecurrenceValue:   task.RecurrenceValue,
        RecurrenceEndDate: task.RecurrenceEndDate,
        ParentTaskID:      task.ParentTaskID,
	}
}
