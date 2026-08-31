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
	CollectorType   string               `json:"collectorType"`
	CollectorConfig json.RawMessage      `json:"collectorConfig"`
	Prompt          string               `json:"prompt"`
	Schedule        CollectorScheduleDTO `json:"schedule"`
	Enabled         bool                 `json:"enabled"`
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
	CollectorType   string               `json:"collectorType"`
	CollectorConfig json.RawMessage      `json:"collectorConfig"`
	Prompt          string               `json:"prompt,omitempty"`
	Schedule        CollectorScheduleDTO `json:"schedule"`
	Enabled         *bool                `json:"enabled,omitempty"`
}

type CollectorAgentUpdateRequest struct {
	Name            *string               `json:"name,omitempty"`
	Description     *string               `json:"description,omitempty"`
	CollectorConfig json.RawMessage       `json:"collectorConfig,omitempty"`
	Prompt          *string               `json:"prompt,omitempty"`
	Schedule        *CollectorScheduleDTO `json:"schedule,omitempty"`
}

type CollectorAgentWriteRaw struct {
	Name            string                `json:"name,omitempty"`
	Description     *string               `json:"description,omitempty"`
	CollectorType   string                `json:"collector_type,omitempty"`
	CollectorConfig json.RawMessage       `json:"collector_config,omitempty"`
	Prompt          *string               `json:"prompt,omitempty"`
	Schedule        *CollectorScheduleRaw `json:"schedule,omitempty"`
	Enabled         *bool                 `json:"enabled,omitempty"`
}

func MapCollectorAgentRaw(raw CollectorAgentRaw) CollectorAgentDetailDTO {
	cfg := raw.CollectorConfig
	if len(cfg) == 0 {
		cfg = json.RawMessage(`{}`)
	}
	return CollectorAgentDetailDTO{
		ID:              raw.ID,
		Code:            raw.Code,
		CompanyID:       raw.CompanyID,
		Name:            raw.Name,
		Description:     raw.Description,
		CollectorType:   raw.CollectorType,
		CollectorConfig: cfg,
		Prompt:          raw.Prompt,
		Schedule:        MapCollectorScheduleRaw(raw.Schedule),
		Enabled:         raw.Enabled,
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
