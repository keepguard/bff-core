package collector

import (
	"encoding/json"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
)

type ListAgentsQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	Filters        map[string]string
}

type GetAgentQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ID             string
}

type CreateAgentCommand struct {
	CompanyFromCtx  string
	TenantID        string
	CorrelationID   string
	Name            string
	Description     string
	Context         string
	CollectorType   string
	CollectorConfig json.RawMessage
	Prompt          string
	Schedule        appdto.CollectorScheduleDTO
	Enabled         *bool
	DataSourceID    string
}

type UpdateAgentCommand struct {
	CompanyFromCtx  string
	TenantID        string
	CorrelationID   string
	ID              string
	Name            *string
	Description     *string
	Context         *string
	CollectorConfig json.RawMessage
	Prompt          *string
	Schedule        *appdto.CollectorScheduleDTO
	DataSourceID    *string
}

type AgentIDCommand struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ID             string
}

type BulkAgentsCommand struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	Action         string
	IDs            []string
}

type GetBulkQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ID             string
}

type GetActiveBulkQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
}

type ListExecutionsQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	AgentID        string
	Limit          int
}

type GetExecutionPayloadsQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ExecutionID    string
}

type ListDataSourcesQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	Filters        map[string]string
}

type GetDataSourceQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ID             string
}

type CreateDataSourceCommand struct {
	CompanyFromCtx      string
	TenantID            string
	CorrelationID       string
	Name                string
	Slug                string
	Description         string
	WebsiteURL          string
	CollectorType       string
	NameTemplate        string
	DescriptionTemplate string
	PromptTemplate      string
	DefaultContext      string
	DefaultSchedule     appdto.CollectorScheduleDTO
	ConfigTemplate      json.RawMessage
	Variables           json.RawMessage
	Notes               string
	Enabled             *bool
	RateLimit           json.RawMessage
}

type UpdateDataSourceCommand struct {
	CompanyFromCtx      string
	TenantID            string
	CorrelationID       string
	ID                  string
	Name                *string
	Slug                *string
	Description         *string
	WebsiteURL          *string
	NameTemplate        *string
	DescriptionTemplate *string
	PromptTemplate      *string
	DefaultContext      *string
	DefaultSchedule     *appdto.CollectorScheduleDTO
	ConfigTemplate      json.RawMessage
	Variables           json.RawMessage
	Notes               *string
	RateLimit           json.RawMessage
}

type DataSourceIDCommand struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ID             string
}

type PropagateDataSourceCommand struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ID             string
	Fields         []string
	DryRun         bool
	Limit          int
}

type ListIncidentsQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	Filters        map[string]string
}

type ListAgentIncidentsQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	AgentID        string
}

type MutateIncidentCommand struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ActorUserID    string
	ID             string
}

type GetIncidentSuggestionQuery struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ID             string
}

type ApplyIncidentSuccessorCommand struct {
	CompanyFromCtx string
	TenantID       string
	CorrelationID  string
	ActorUserID    string
	ID             string
	Confirmed      bool
}
