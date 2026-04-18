package task

import (
	"context"
	"fmt"
	"strings"
	"time"
	"strconv"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// Генерирует даты для периодической задачи
func generateRecurrenceDates(recurrenceType taskdomain.RecurrenceType, recurrenceValue string, startDate time.Time, limit int) []time.Time {
    dates := make([]time.Time, 0)
    
    switch recurrenceType {
		case taskdomain.RecurrenceDaily:
			// recurrenceValue = "2" → каждые 2 дня
			days, err := strconv.Atoi(recurrenceValue)
			if err != nil {
				days = 1 // значение по умолчанию
			}
			for i := 1; i <= limit; i++ {
				dates = append(dates, startDate.AddDate(0, 0, days*i))
			}
			
		case taskdomain.RecurrenceMonthly:
			// recurrenceValue = "15" → 15-е число каждого месяца
			day, err := strconv.Atoi(recurrenceValue)
			if err != nil || day < 1 || day > 31 {
				day = 1 // значение по умолчанию
			}
			for i := 1; i <= limit/30; i++ {
				// Переходим на следующий месяц
				nextMonth := startDate.AddDate(0, i, 0)
				// Создаём дату с нужным числом
				date := time.Date(nextMonth.Year(), nextMonth.Month(), day, 0, 0, 0, 0, time.UTC)
				dates = append(dates, date)
			}
			
		case taskdomain.RecurrenceParity:
			// "even" → чётные дни, "odd" → нечётные
			for i := 1; i <= limit; i++ {
				date := startDate.AddDate(0, 0, i)
				if recurrenceValue == "even" && date.Day()%2 == 0 {
					dates = append(dates, date)
				} else if recurrenceValue == "odd" && date.Day()%2 == 1 {
					dates = append(dates, date)
				}
			}
			
		case taskdomain.RecurrenceSpecific:
		// recurrenceValue = "2026-05-01,2026-05-15,2026-06-01"
		dateStrings := strings.Split(recurrenceValue, ",")
		for _, dateStr := range dateStrings {
			dateStr = strings.TrimSpace(dateStr)
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				// Пропускаем некорректные даты
				continue
			}
			// Добавляем только будущие даты (от startDate и позже)
			if date.After(startDate) || date.Equal(startDate) {
				dates = append(dates, date)
			}
		}  
	}
	return dates
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		RecurrenceType:    input.RecurrenceType,
		RecurrenceValue:   input.RecurrenceValue,
		RecurrenceEndDate: input.RecurrenceEndDate,
		ParentTaskID:      nil,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	if err := s.createRecurringTasks(ctx, created); err != nil {
        // Логируем ошибку, но не отменяем создание основной задачи
        fmt.Printf("failed to create recurring tasks: %v\n", err)
    }

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

// Создаёт дочерние задачи для периодической задачи
func (s *Service) createRecurringTasks(ctx context.Context, parentTask *taskdomain.Task) error {
    // Если нет типа периодичности — выходим
    if parentTask.RecurrenceType == nil {
        return nil
    }
    
    // Если нет значения — выходим
    if parentTask.RecurrenceValue == nil {
        return nil
    }
    
    // Создаём задачи на 30 дней вперёд
    const recurrenceLimit = 30
    
    // Генерируем даты
    dates := generateRecurrenceDates(
        *parentTask.RecurrenceType,
        *parentTask.RecurrenceValue,
        parentTask.CreatedAt,
        recurrenceLimit,
    )

	// Конвертируем ID в строку для ParentTaskID
    parentIDStr := fmt.Sprintf("%d", parentTask.ID)
    
    // Для каждой даты создаём задачу
    for _, date := range dates {
        childTask := &taskdomain.Task{
            Title:       parentTask.Title,
            Description: parentTask.Description,
            Status:      taskdomain.StatusNew,
            CreatedAt:   date,
            UpdatedAt:   date,
            ParentTaskID: &parentIDStr,
        }
        
        if _, err := s.repo.Create(ctx, childTask); err != nil {
            // Логируем ошибку, но продолжаем
            fmt.Printf("failed to create recurring task for date %v: %v\n", date, err)
        }
    }
    
    return nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}
