package dto

import "encoding/json"

type CollectorScheduleDTO struct {
	DaysOfWeek      []int  `json:"daysOfWeek"`
	StartTime       string `json:"startTime"`
	EndTime         string `json:"endTime"`
	IntervalMinutes int    `json:"intervalMinutes"`
	Timezone        string `json:"timezone"`
}

type CollectorScheduleRaw struct {
	DaysOfWeek      []int  `json:"days_of_week"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	IntervalMinutes int    `json:"interval_minutes"`
	Timezone        string `json:"timezone"`
}

type CollectorAgentDetailDTO struct {
	ID              string               `json:"id"`
	Code            string               `json:"code"`
	CompanyID       string               `json:"companyId"`
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Context         string               `json:"context"`
	CollectorType   string               `json:"collectorType"`
	CollectorConfig json.RawMessage      `json:"collectorConfig"`
	Prompt          string               `json:"prompt"`
	Schedule        CollectorScheduleDTO `json:"schedule"`
	Enabled         bool                 `json:"enabled"`
	DataSourceID    string               `json:"dataSourceId,omitempty"`
	DataSourceSlug  string               `json:"dataSourceSlug,omitempty"`
	DataSourceName  string               `json:"dataSourceName,omitempty"`
	CreatedAt       string               `json:"createdAt"`
	UpdatedAt       string               `json:"updatedAt"`
}

type PaginatedCollectorAgents struct {
	Content       []CollectorAgentDetailDTO `json:"content"`
	Page          int                       `json:"page"`
	Size          int                       `json:"size"`
	TotalElements int64                     `json:"totalElements"`
	TotalPages    int                       `json:"totalPages"`
	First         bool                      `json:"first"`
	Last          bool                      `json:"last"`
	HasNext       bool                      `json:"hasNext"`
	HasPrevious   bool                      `json:"hasPrevious"`
}

type PaginatedCollectorAgentsRaw struct {
	Content       []CollectorAgentRaw `json:"content"`
	Page          int                 `json:"page"`
	Size          int                 `json:"size"`
	TotalElements int64               `json:"totalElements"`
	TotalPages    int                 `json:"totalPages"`
	First         bool                `json:"first"`
	Last          bool                `json:"last"`
	HasNext       bool                `json:"hasNext"`
	HasPrevious   bool                `json:"hasPrevious"`
}

type CollectorAgentCreateRequest struct {
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	Context         string               `json:"context,omitempty"`
	CollectorType   string               `json:"collectorType"`
	CollectorConfig json.RawMessage      `json:"collectorConfig"`
	Prompt          string               `json:"prompt,omitempty"`
	Schedule        CollectorScheduleDTO `json:"schedule"`
	Enabled         *bool                `json:"enabled,omitempty"`
}

type CollectorAgentUpdateRequest struct {
	Name            *string               `json:"name,omitempty"`
	Description     *string               `json:"description,omitempty"`
	Context         *string               `json:"context,omitempty"`
	CollectorConfig json.RawMessage       `json:"collectorConfig,omitempty"`
	Prompt          *string               `json:"prompt,omitempty"`
	Schedule        *CollectorScheduleDTO `json:"schedule,omitempty"`
}

type CollectorAgentWriteRaw struct {
	Name            string                `json:"name,omitempty"`
	Description     *string               `json:"description,omitempty"`
	Context         *string               `json:"context,omitempty"`
	CollectorType   string                `json:"collector_type,omitempty"`
	CollectorConfig json.RawMessage       `json:"collector_config,omitempty"`
	Prompt          *string               `json:"prompt,omitempty"`
	Schedule        *CollectorScheduleRaw `json:"schedule,omitempty"`
	Enabled         *bool                 `json:"enabled,omitempty"`
	DataSourceID    *string               `json:"data_source_id,omitempty"`
}

func MapCollectorAgentRaw(raw CollectorAgentRaw) CollectorAgentDetailDTO {
	cfg := raw.CollectorConfig
	if len(cfg) == 0 {
		cfg = json.RawMessage(`{}`)
	}
	context := raw.Context
	if context == "" {
		context = "geral"
	}
	return CollectorAgentDetailDTO{
		ID:              raw.ID,
		Code:            raw.Code,
		CompanyID:       raw.CompanyID,
		Name:            raw.Name,
		Description:     raw.Description,
		Context:         context,
		CollectorType:   raw.CollectorType,
		CollectorConfig: cfg,
		Prompt:          raw.Prompt,
		Schedule:        MapCollectorScheduleRaw(raw.Schedule),
		Enabled:         raw.Enabled,
		DataSourceID:    raw.DataSourceID,
		DataSourceSlug:  raw.DataSourceSlug,
		DataSourceName:  raw.DataSourceName,
		CreatedAt:       raw.CreatedAt,
		UpdatedAt:       raw.UpdatedAt,
	}
}

func MapCollectorScheduleRaw(raw json.RawMessage) CollectorScheduleDTO {
	if len(raw) == 0 {
		return CollectorScheduleDTO{}
	}
	var parsed CollectorScheduleRaw
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return CollectorScheduleDTO{}
	}
	return CollectorScheduleDTO{
		DaysOfWeek:      parsed.DaysOfWeek,
		StartTime:       parsed.StartTime,
		EndTime:         parsed.EndTime,
		IntervalMinutes: parsed.IntervalMinutes,
		Timezone:        parsed.Timezone,
	}
}

func MapCollectorScheduleDTO(dto CollectorScheduleDTO) CollectorScheduleRaw {
	return CollectorScheduleRaw{
		DaysOfWeek:      dto.DaysOfWeek,
		StartTime:       dto.StartTime,
		EndTime:         dto.EndTime,
		IntervalMinutes: dto.IntervalMinutes,
		Timezone:        dto.Timezone,
	}
}

func MapPaginatedCollectorAgents(raw PaginatedCollectorAgentsRaw) PaginatedCollectorAgents {
	content := make([]CollectorAgentDetailDTO, 0, len(raw.Content))
	for _, item := range raw.Content {
		content = append(content, MapCollectorAgentRaw(item))
	}
	return PaginatedCollectorAgents{
		Content:       content,
		Page:          raw.Page,
		Size:          raw.Size,
		TotalElements: raw.TotalElements,
		TotalPages:    raw.TotalPages,
		First:         raw.First,
		Last:          raw.Last,
		HasNext:       raw.HasNext,
		HasPrevious:   raw.HasPrevious,
	}
}

type CollectorAgentTestPreviewRaw struct {
	FileName         string `json:"file_name"`
	ContentType      string `json:"content_type"`
	SizeBytes        int    `json:"size_bytes"`
	PreviewTruncated bool   `json:"preview_truncated"`
	PreviewText      string `json:"preview_text,omitempty"`
}

type CollectorAgentTestResultRaw struct {
	Success        bool                           `json:"success"`
	AgentID        string                         `json:"agent_id"`
	CollectorType  string                         `json:"collector_type"`
	ItemsCollected int                            `json:"items_collected"`
	DurationMs     int64                          `json:"duration_ms"`
	Error          *string                        `json:"error"`
	Preview        []CollectorAgentTestPreviewRaw `json:"preview"`
}

type CollectorAgentTestPreviewDTO struct {
	FileName         string `json:"fileName"`
	ContentType      string `json:"contentType"`
	SizeBytes        int    `json:"sizeBytes"`
	PreviewTruncated bool   `json:"previewTruncated"`
	PreviewText      string `json:"previewText,omitempty"`
}

type CollectorAgentTestResultDTO struct {
	Success        bool                           `json:"success"`
	AgentID        string                         `json:"agentId"`
	CollectorType  string                         `json:"collectorType"`
	ItemsCollected int                            `json:"itemsCollected"`
	DurationMs     int64                          `json:"durationMs"`
	Error          *string                        `json:"error"`
	Preview        []CollectorAgentTestPreviewDTO `json:"preview"`
}

type CollectorAgentRunResultRaw struct {
	Status  string `json:"status"`
	AgentID string `json:"agent_id"`
}

type CollectorAgentRunResultDTO struct {
	Status  string `json:"status"`
	AgentID string `json:"agentId"`
}

func MapCollectorAgentRunResult(raw CollectorAgentRunResultRaw) CollectorAgentRunResultDTO {
	return CollectorAgentRunResultDTO{
		Status:  raw.Status,
		AgentID: raw.AgentID,
	}
}

type CollectorExecutionRaw struct {
	ID             string         `json:"id"`
	AgentID        string         `json:"agent_id"`
	StartedAt      string         `json:"started_at"`
	FinishedAt     *string        `json:"finished_at,omitempty"`
	Status         string         `json:"status"`
	ItemsCollected int            `json:"items_collected"`
	ItemsUploaded  int            `json:"items_uploaded"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

type CollectorExecutionDTO struct {
	ID             string         `json:"id"`
	AgentID        string         `json:"agentId"`
	StartedAt      string         `json:"startedAt"`
	FinishedAt     *string        `json:"finishedAt,omitempty"`
	Status         string         `json:"status"`
	ItemsCollected int            `json:"itemsCollected"`
	ItemsUploaded  int            `json:"itemsUploaded"`
	ErrorMessage   string         `json:"errorMessage,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

func MapCollectorExecution(raw CollectorExecutionRaw) CollectorExecutionDTO {
	return CollectorExecutionDTO{
		ID:             raw.ID,
		AgentID:        raw.AgentID,
		StartedAt:      raw.StartedAt,
		FinishedAt:     raw.FinishedAt,
		Status:         raw.Status,
		ItemsCollected: raw.ItemsCollected,
		ItemsUploaded:  raw.ItemsUploaded,
		ErrorMessage:   raw.ErrorMessage,
		Metadata:       raw.Metadata,
	}
}

func MapCollectorExecutions(raw []CollectorExecutionRaw) []CollectorExecutionDTO {
	out := make([]CollectorExecutionDTO, 0, len(raw))
	for _, item := range raw {
		out = append(out, MapCollectorExecution(item))
	}
	return out
}

func MapCollectorAgentTestResult(raw CollectorAgentTestResultRaw) CollectorAgentTestResultDTO {
	preview := make([]CollectorAgentTestPreviewDTO, 0, len(raw.Preview))
	for _, item := range raw.Preview {
		preview = append(preview, CollectorAgentTestPreviewDTO{
			FileName:         item.FileName,
			ContentType:      item.ContentType,
			SizeBytes:        item.SizeBytes,
			PreviewTruncated: item.PreviewTruncated,
			PreviewText:      item.PreviewText,
		})
	}
	return CollectorAgentTestResultDTO{
		Success:        raw.Success,
		AgentID:        raw.AgentID,
		CollectorType:  raw.CollectorType,
		ItemsCollected: raw.ItemsCollected,
		DurationMs:     raw.DurationMs,
		Error:          raw.Error,
		Preview:        preview,
	}
}

type CollectorDataSourceVariable struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
}

type CollectorDataSourceDTO struct {
	ID                  string               `json:"id"`
	Code                string               `json:"code"`
	CompanyID           string               `json:"companyId,omitempty"`
	Scope               string               `json:"scope"`
	Slug                string               `json:"slug"`
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	WebsiteURL          string               `json:"websiteUrl,omitempty"`
	CollectorType       string               `json:"collectorType"`
	NameTemplate        string               `json:"nameTemplate,omitempty"`
	DescriptionTemplate string               `json:"descriptionTemplate,omitempty"`
	PromptTemplate      string               `json:"promptTemplate,omitempty"`
	DefaultContext      string               `json:"defaultContext"`
	DefaultSchedule     CollectorScheduleDTO `json:"defaultSchedule"`
	ConfigTemplate      json.RawMessage      `json:"configTemplate"`
	Variables           json.RawMessage      `json:"variables"`
	Notes               string               `json:"notes,omitempty"`
	Enabled             bool                 `json:"enabled"`
	RateLimit           json.RawMessage      `json:"rateLimit,omitempty"`
	CreatedAt           string               `json:"createdAt,omitempty"`
	UpdatedAt           string               `json:"updatedAt,omitempty"`
}

type CollectorDataSourceRaw struct {
	ID                  string          `json:"id"`
	Code                string          `json:"code"`
	CompanyID           string          `json:"company_id,omitempty"`
	Scope               string          `json:"scope,omitempty"`
	Slug                string          `json:"slug"`
	Name                string          `json:"name"`
	Description         string          `json:"description"`
	WebsiteURL          string          `json:"website_url,omitempty"`
	CollectorType       string          `json:"collector_type"`
	NameTemplate        string          `json:"name_template,omitempty"`
	DescriptionTemplate string          `json:"description_template,omitempty"`
	PromptTemplate      string          `json:"prompt_template,omitempty"`
	DefaultContext      string          `json:"default_context"`
	DefaultSchedule     json.RawMessage `json:"default_schedule"`
	ConfigTemplate      json.RawMessage `json:"config_template"`
	Variables           json.RawMessage `json:"variables"`
	Notes               string          `json:"notes,omitempty"`
	Enabled             bool            `json:"enabled"`
	RateLimit           json.RawMessage `json:"rate_limit,omitempty"`
	CreatedAt           string          `json:"created_at,omitempty"`
	UpdatedAt           string          `json:"updated_at,omitempty"`
}

type CollectorDataSourceWriteRaw struct {
	Name                string          `json:"name,omitempty"`
	Slug                string          `json:"slug,omitempty"`
	Description         *string         `json:"description,omitempty"`
	WebsiteURL          *string         `json:"website_url,omitempty"`
	CollectorType       string          `json:"collector_type,omitempty"`
	NameTemplate        *string         `json:"name_template,omitempty"`
	DescriptionTemplate *string         `json:"description_template,omitempty"`
	PromptTemplate      *string         `json:"prompt_template,omitempty"`
	DefaultContext      *string         `json:"default_context,omitempty"`
	DefaultSchedule     json.RawMessage `json:"default_schedule,omitempty"`
	ConfigTemplate      json.RawMessage `json:"config_template,omitempty"`
	Variables           json.RawMessage `json:"variables,omitempty"`
	Notes               *string         `json:"notes,omitempty"`
	Enabled             *bool           `json:"enabled,omitempty"`
	RateLimit           json.RawMessage `json:"rate_limit,omitempty"`
}

func MapCollectorDataSourceRaw(raw CollectorDataSourceRaw) CollectorDataSourceDTO {
	scope := raw.Scope
	if scope == "" {
		scope = "company"
	}
	return CollectorDataSourceDTO{
		ID:                  raw.ID,
		Code:                raw.Code,
		CompanyID:           raw.CompanyID,
		Scope:               scope,
		Slug:                raw.Slug,
		Name:                raw.Name,
		Description:         raw.Description,
		WebsiteURL:          raw.WebsiteURL,
		CollectorType:       raw.CollectorType,
		NameTemplate:        raw.NameTemplate,
		DescriptionTemplate: raw.DescriptionTemplate,
		PromptTemplate:      raw.PromptTemplate,
		DefaultContext:      raw.DefaultContext,
		DefaultSchedule:     MapCollectorScheduleRaw(raw.DefaultSchedule),
		ConfigTemplate:      raw.ConfigTemplate,
		Variables:           raw.Variables,
		Notes:               raw.Notes,
		Enabled:             raw.Enabled,
		RateLimit:           raw.RateLimit,
		CreatedAt:           raw.CreatedAt,
		UpdatedAt:           raw.UpdatedAt,
	}
}

func MapCollectorDataSources(raw []CollectorDataSourceRaw) []CollectorDataSourceDTO {
	out := make([]CollectorDataSourceDTO, 0, len(raw))
	for _, item := range raw {
		out = append(out, MapCollectorDataSourceRaw(item))
	}
	return out
}

type PropagateDataSourceWriteRaw struct {
	Fields []string `json:"fields"`
	DryRun bool     `json:"dry_run"`
	Limit  int      `json:"limit,omitempty"`
}

type PropagateAgentPreviewRaw struct {
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	Ticker     string `json:"ticker"`
	BeforeURL  string `json:"before_url"`
	AfterURL   string `json:"after_url"`
	Changed    bool   `json:"changed"`
	SkipReason string `json:"skip_reason,omitempty"`
}

type PropagateDataSourceRaw struct {
	TotalLinked int                        `json:"total_linked"`
	Updated     int                        `json:"updated"`
	Skipped     int                        `json:"skipped"`
	Failed      int                        `json:"failed"`
	DryRun      bool                       `json:"dry_run"`
	Previews    []PropagateAgentPreviewRaw `json:"previews"`
	Errors      []string                   `json:"errors,omitempty"`
}

type PropagateAgentPreviewDTO struct {
	AgentID    string `json:"agentId"`
	AgentName  string `json:"agentName"`
	Ticker     string `json:"ticker"`
	BeforeURL  string `json:"beforeUrl"`
	AfterURL   string `json:"afterUrl"`
	Changed    bool   `json:"changed"`
	SkipReason string `json:"skipReason,omitempty"`
}

type PropagateDataSourceDTO struct {
	TotalLinked int                        `json:"totalLinked"`
	Updated     int                        `json:"updated"`
	Skipped     int                        `json:"skipped"`
	Failed      int                        `json:"failed"`
	DryRun      bool                       `json:"dryRun"`
	Previews    []PropagateAgentPreviewDTO `json:"previews"`
	Errors      []string                   `json:"errors,omitempty"`
}

func MapPropagateDataSourceRaw(raw PropagateDataSourceRaw) PropagateDataSourceDTO {
	previews := make([]PropagateAgentPreviewDTO, 0, len(raw.Previews))
	for _, item := range raw.Previews {
		previews = append(previews, PropagateAgentPreviewDTO{
			AgentID:    item.AgentID,
			AgentName:  item.AgentName,
			Ticker:     item.Ticker,
			BeforeURL:  item.BeforeURL,
			AfterURL:   item.AfterURL,
			Changed:    item.Changed,
			SkipReason: item.SkipReason,
		})
	}
	if raw.Errors == nil {
		raw.Errors = []string{}
	}
	return PropagateDataSourceDTO{
		TotalLinked: raw.TotalLinked,
		Updated:     raw.Updated,
		Skipped:     raw.Skipped,
		Failed:      raw.Failed,
		DryRun:      raw.DryRun,
		Previews:    previews,
		Errors:      raw.Errors,
	}
}
