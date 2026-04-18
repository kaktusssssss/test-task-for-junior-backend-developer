package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string  // Переодичность

const (
    RecurrenceDaily    RecurrenceType = "daily"     // Ежедневно, каждый N-й день
    RecurrenceMonthly  RecurrenceType = "monthly"   // Ежемесячно
    RecurrenceSpecific RecurrenceType = "specific"  // Конкретные даты
    RecurrenceParity   RecurrenceType = "parity"    // Чётные/нечётные дни
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	RecurrenceType    *RecurrenceType `json:"recurrence_type,omitempty"`
    RecurrenceValue   *string         `json:"recurrence_value,omitempty"`
    RecurrenceEndDate *time.Time      `json:"recurrence_end_date,omitempty"`
    ParentTaskID      *string         `json:"parent_task_id,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

