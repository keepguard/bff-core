package dto

import (
	"encoding/json"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type CollectorAgentCreateRequestDTO struct {
	Name            string                      `json:"name"`
	Description     string                      `json:"description,omitempty"`
	Context         string                      `json:"context,omitempty"`
	CollectorType   string                      `json:"collectorType"`
	CollectorConfig json.RawMessage             `json:"collectorConfig"`
	Prompt          string                      `json:"prompt,omitempty"`
	Schedule        appdto.CollectorScheduleDTO `json:"schedule"`
	Enabled         *bool                       `json:"enabled,omitempty"`
	DataSourceID    string                      `json:"dataSourceId,omitempty"`
}

type CollectorAgentUpdateRequestDTO struct {
	Name            *string                      `json:"name,omitempty"`
	Description     *string                      `json:"description,omitempty"`
	Context         *string                      `json:"context,omitempty"`
	CollectorConfig json.RawMessage              `json:"collectorConfig,omitempty"`
	Prompt          *string                      `json:"prompt,omitempty"`
	Schedule        *appdto.CollectorScheduleDTO `json:"schedule,omitempty"`
	DataSourceID    *string                      `json:"dataSourceId,omitempty"`
}

type CollectorDataSourceCreateRequestDTO struct {
	Name                string                      `json:"name"`
	Slug                string                      `json:"slug"`
	Description         string                      `json:"description,omitempty"`
	WebsiteURL          string                      `json:"websiteUrl,omitempty"`
	CollectorType       string                      `json:"collectorType"`
	NameTemplate        string                      `json:"nameTemplate,omitempty"`
	DescriptionTemplate string                      `json:"descriptionTemplate,omitempty"`
	PromptTemplate      string                      `json:"promptTemplate,omitempty"`
	DefaultContext      string                      `json:"defaultContext,omitempty"`
	DefaultSchedule     appdto.CollectorScheduleDTO `json:"defaultSchedule"`
	ConfigTemplate      json.RawMessage             `json:"configTemplate"`
	Variables           json.RawMessage             `json:"variables"`
	Notes               string                      `json:"notes,omitempty"`
	Enabled             *bool                       `json:"enabled,omitempty"`
	RateLimit           json.RawMessage             `json:"rateLimit,omitempty"`
}

type CollectorDataSourceUpdateRequestDTO struct {
	Name                *string                      `json:"name,omitempty"`
	Slug                *string                      `json:"slug,omitempty"`
	Description         *string                      `json:"description,omitempty"`
	WebsiteURL          *string                      `json:"websiteUrl,omitempty"`
	NameTemplate        *string                      `json:"nameTemplate,omitempty"`
	DescriptionTemplate *string                      `json:"descriptionTemplate,omitempty"`
	PromptTemplate      *string                      `json:"promptTemplate,omitempty"`
	DefaultContext      *string                      `json:"defaultContext,omitempty"`
	DefaultSchedule     *appdto.CollectorScheduleDTO `json:"defaultSchedule,omitempty"`
	ConfigTemplate      json.RawMessage              `json:"configTemplate,omitempty"`
	Variables           json.RawMessage              `json:"variables,omitempty"`
	Notes               *string                      `json:"notes,omitempty"`
	RateLimit           json.RawMessage              `json:"rateLimit,omitempty"`
}

type CollectorPropagateRequestDTO struct {
	Fields []string `json:"fields"`
	DryRun bool     `json:"dryRun"`
	Limit  int      `json:"limit,omitempty"`
}

type CollectorBulkRequestDTO struct {
	Action string   `json:"action"`
	IDs    []string `json:"ids"`
}

type CollectorApplySuccessorRequestDTO struct {
	Confirmed bool `json:"confirmed"`
}
