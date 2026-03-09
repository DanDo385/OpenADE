package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	cron "github.com/robfig/cron/v3"
	"openade/internal/db"
	"openade/internal/llm"
	"openade/internal/model"
)

type SchedulerService struct {
	DB         *sql.DB
	Tasks      *TaskService
	Runs       *RunService
	Providers  *ProviderService
	NewAdapter func(cfg *model.ProviderConfig) llm.Adapter

	stopOnce sync.Once
	stopCh   chan struct{}
}

func NewSchedulerService(
	database *sql.DB,
	taskSvc *TaskService,
	runSvc *RunService,
	providerSvc *ProviderService,
	newAdapter func(cfg *model.ProviderConfig) llm.Adapter,
) *SchedulerService {
	return &SchedulerService{
		DB:         database,
		Tasks:      taskSvc,
		Runs:       runSvc,
		Providers:  providerSvc,
		NewAdapter: newAdapter,
		stopCh:     make(chan struct{}),
	}
}

func (s *SchedulerService) Start(ctx context.Context) {
	go s.runLoop(ctx)
}

func (s *SchedulerService) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}

func (s *SchedulerService) List(ctx context.Context, taskID string) ([]model.Schedule, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if taskID != "" {
		rows, err = s.DB.QueryContext(ctx, `
			SELECT id, task_id, cron_expr, timezone, enabled, last_run_at, next_run_at, created_at, updated_at
			FROM scheduled_jobs
			WHERE task_id = ?
			ORDER BY created_at DESC
		`, taskID)
	} else {
		rows, err = s.DB.QueryContext(ctx, `
			SELECT id, task_id, cron_expr, timezone, enabled, last_run_at, next_run_at, created_at, updated_at
			FROM scheduled_jobs
			ORDER BY created_at DESC
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("listing schedules: %w", err)
	}
	defer rows.Close()

	var schedules []model.Schedule
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, *schedule)
	}
	return schedules, rows.Err()
}

func (s *SchedulerService) Get(ctx context.Context, id string) (*model.Schedule, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, task_id, cron_expr, timezone, enabled, last_run_at, next_run_at, created_at, updated_at
		FROM scheduled_jobs
		WHERE id = ?
	`, id)
	return scanScheduleRow(row)
}

func (s *SchedulerService) GetByTaskID(ctx context.Context, taskID string) (*model.Schedule, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, task_id, cron_expr, timezone, enabled, last_run_at, next_run_at, created_at, updated_at
		FROM scheduled_jobs
		WHERE task_id = ?
	`, taskID)
	return scanScheduleRow(row)
}

func (s *SchedulerService) Create(ctx context.Context, req model.CreateScheduleRequest) (*model.Schedule, error) {
	if strings.TrimSpace(req.TaskID) == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	payload, err := s.normalizeCreate(req)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	nowString := db.FormatTime(now)
	var nextRun any
	if payload.NextRunAt != nil {
		nextRun = db.FormatTime(*payload.NextRunAt)
	}

	_, err = s.DB.ExecContext(ctx, `
		INSERT INTO scheduled_jobs (id, task_id, cron_expr, timezone, enabled, next_run_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, id, payload.TaskID, payload.CronExpr, payload.Timezone, boolToInt(payload.Enabled), nextRun, nowString, nowString)
	if err != nil {
		return nil, fmt.Errorf("creating schedule: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *SchedulerService) Update(ctx context.Context, id string, req model.UpdateScheduleRequest) (*model.Schedule, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("schedule not found")
	}

	payload, err := s.normalizeUpdate(*current, req)
	if err != nil {
		return nil, err
	}

	nowString := db.FormatTime(time.Now().UTC())
	var nextRun any
	if payload.NextRunAt != nil {
		nextRun = db.FormatTime(*payload.NextRunAt)
	}

	_, err = s.DB.ExecContext(ctx, `
		UPDATE scheduled_jobs
		SET cron_expr = ?, timezone = ?, enabled = ?, next_run_at = ?, updated_at = ?
		WHERE id = ?
	`, payload.CronExpr, payload.Timezone, boolToInt(payload.Enabled), nextRun, nowString, id)
	if err != nil {
		return nil, fmt.Errorf("updating schedule: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *SchedulerService) UpsertForTask(ctx context.Context, taskID string, req model.UpdateScheduleRequest) (*model.Schedule, error) {
	existing, err := s.GetByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		createReq := model.CreateScheduleRequest{
			TaskID:   taskID,
			Enabled:  req.Enabled,
			Timezone: "",
		}
		if req.CronExpr != nil {
			createReq.CronExpr = *req.CronExpr
		}
		if req.Timezone != nil {
			createReq.Timezone = *req.Timezone
		}
		return s.Create(ctx, createReq)
	}
	return s.Update(ctx, existing.ID, req)
}

func (s *SchedulerService) Delete(ctx context.Context, id string) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM scheduled_jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("deleting schedule: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("schedule not found")
	}
	return nil
}

func (s *SchedulerService) DeleteByTaskID(ctx context.Context, taskID string) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM scheduled_jobs WHERE task_id = ?`, taskID)
	if err != nil {
		return fmt.Errorf("deleting schedule: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("schedule not found")
	}
	return nil
}

func (s *SchedulerService) runLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	if err := s.processDueJobs(context.Background()); err != nil {
		log.Printf("scheduler initial check failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.processDueJobs(context.Background()); err != nil {
				log.Printf("scheduler check failed: %v", err)
			}
		}
	}
}

func (s *SchedulerService) processDueJobs(ctx context.Context) error {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, task_id, cron_expr, timezone, enabled, last_run_at, next_run_at, created_at, updated_at
		FROM scheduled_jobs
		WHERE enabled = 1 AND next_run_at IS NOT NULL AND next_run_at <= ?
		ORDER BY next_run_at ASC
	`, db.FormatTime(time.Now().UTC()))
	if err != nil {
		return fmt.Errorf("querying due schedules: %w", err)
	}
	defer rows.Close()

	var due []model.Schedule
	for rows.Next() {
		schedule, err := scanSchedule(rows)
		if err != nil {
			return err
		}
		due = append(due, *schedule)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, schedule := range due {
		schedule := schedule
		go s.executeSchedule(schedule)
	}
	return nil
}

func (s *SchedulerService) executeSchedule(schedule model.Schedule) {
	now := time.Now().UTC()
	nextRunAt, err := nextRunTime(schedule.CronExpr, schedule.Timezone, now)
	if err != nil {
		log.Printf("scheduler invalid cron for task %s: %v", schedule.TaskID, err)
		return
	}

	_, err = s.DB.ExecContext(context.Background(), `
		UPDATE scheduled_jobs
		SET last_run_at = ?, next_run_at = ?, updated_at = ?
		WHERE id = ? AND enabled = 1
	`, db.FormatTime(now), db.FormatTime(nextRunAt), db.FormatTime(now), schedule.ID)
	if err != nil {
		log.Printf("scheduler failed to claim schedule %s: %v", schedule.ID, err)
		return
	}

	task, err := s.Tasks.Get(context.Background(), schedule.TaskID)
	if err != nil {
		log.Printf("scheduler failed to load task %s: %v", schedule.TaskID, err)
		return
	}
	if task == nil {
		log.Printf("scheduler task not found for schedule %s", schedule.ID)
		return
	}

	provider, err := s.Providers.GetDefault(context.Background())
	if err != nil {
		log.Printf("scheduler failed to load provider: %v", err)
		return
	}
	if provider == nil {
		log.Printf("scheduler skipped task %s: no LLM provider configured", task.ID)
		return
	}

	inputs := defaultInputs(task)
	if _, err := s.Runs.Execute(context.Background(), task, inputs, s.NewAdapter(provider), provider.DefaultModel); err != nil {
		log.Printf("scheduler failed to run task %s: %v", task.ID, err)
	}
}

type schedulePayload struct {
	TaskID    string
	CronExpr  string
	Timezone  string
	Enabled   bool
	NextRunAt *time.Time
}

func (s *SchedulerService) normalizeCreate(req model.CreateScheduleRequest) (*schedulePayload, error) {
	payload := &schedulePayload{
		TaskID:   strings.TrimSpace(req.TaskID),
		CronExpr: strings.TrimSpace(req.CronExpr),
		Timezone: strings.TrimSpace(req.Timezone),
		Enabled:  true,
	}
	if req.Enabled != nil {
		payload.Enabled = *req.Enabled
	}
	return s.normalizePayload(payload)
}

func (s *SchedulerService) normalizeUpdate(current model.Schedule, req model.UpdateScheduleRequest) (*schedulePayload, error) {
	payload := &schedulePayload{
		TaskID:   current.TaskID,
		CronExpr: current.CronExpr,
		Timezone: current.Timezone,
		Enabled:  current.Enabled,
	}
	if req.CronExpr != nil {
		payload.CronExpr = strings.TrimSpace(*req.CronExpr)
	}
	if req.Timezone != nil {
		payload.Timezone = strings.TrimSpace(*req.Timezone)
	}
	if req.Enabled != nil {
		payload.Enabled = *req.Enabled
	}
	return s.normalizePayload(payload)
}

func (s *SchedulerService) normalizePayload(payload *schedulePayload) (*schedulePayload, error) {
	if payload.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if payload.CronExpr == "" {
		return nil, fmt.Errorf("cron_expr is required")
	}
	if task, err := s.Tasks.Get(context.Background(), payload.TaskID); err != nil {
		return nil, fmt.Errorf("loading task: %w", err)
	} else if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	if !payload.Enabled {
		payload.NextRunAt = nil
		return payload, nil
	}

	nextRunAt, err := nextRunTime(payload.CronExpr, payload.Timezone, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	payload.NextRunAt = &nextRunAt
	return payload, nil
}

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(rows *sql.Rows) (*model.Schedule, error) {
	return scanScheduleFrom(rows)
}

func scanScheduleRow(row *sql.Row) (*model.Schedule, error) {
	return scanScheduleFrom(row)
}

func scanScheduleFrom(scanner scheduleScanner) (*model.Schedule, error) {
	var schedule model.Schedule
	var enabled int
	var lastRunAt, nextRunAt, createdAt, updatedAt sql.NullString
	err := scanner.Scan(
		&schedule.ID,
		&schedule.TaskID,
		&schedule.CronExpr,
		&schedule.Timezone,
		&enabled,
		&lastRunAt,
		&nextRunAt,
		&createdAt,
		&updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning schedule: %w", err)
	}

	schedule.Enabled = enabled != 0
	schedule.LastRunAt = parseNullableTime(lastRunAt)
	schedule.NextRunAt = parseNullableTime(nextRunAt)
	if createdAt.Valid {
		schedule.CreatedAt = db.ParseTime(createdAt.String)
	}
	if updatedAt.Valid {
		schedule.UpdatedAt = db.ParseTime(updatedAt.String)
	}
	return &schedule, nil
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid || strings.TrimSpace(value.String) == "" {
		return nil
	}
	parsed := db.ParseTime(value.String)
	if parsed.IsZero() {
		return nil
	}
	return &parsed
}

func nextRunTime(cronExpr, timezone string, from time.Time) (time.Time, error) {
	spec := strings.TrimSpace(cronExpr)
	if spec == "" {
		return time.Time{}, fmt.Errorf("cron_expr is required")
	}
	if timezone != "" {
		if _, err := time.LoadLocation(timezone); err != nil {
			return time.Time{}, fmt.Errorf("invalid timezone: %w", err)
		}
		spec = "CRON_TZ=" + timezone + " " + spec
	}
	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	next := schedule.Next(from)
	if next.IsZero() {
		return time.Time{}, fmt.Errorf("could not determine next run time")
	}
	return next.UTC(), nil
}

func defaultInputs(task *model.Task) map[string]interface{} {
	inputs := make(map[string]interface{}, len(task.InputSchema))
	for _, field := range task.InputSchema {
		if field.Default == "" {
			continue
		}
		switch field.Type {
		case "number":
			inputs[field.Key] = field.Default
		case "boolean":
			inputs[field.Key] = strings.EqualFold(field.Default, "true")
		default:
			inputs[field.Key] = field.Default
		}
	}
	return inputs
}
