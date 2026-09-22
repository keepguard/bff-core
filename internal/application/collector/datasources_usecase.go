package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/keepguard/bff-core/internal/pkg"
)

func (s *service) ListDataSources(ctx context.Context, query ListDataSourcesQuery) ([]appdto.CollectorDataSourceDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return nil, err
	}
	raw, err := s.collector.ListDataSources(ctx, companyID, query.CorrelationID, query.Filters)
	if err != nil {
		return nil, err
	}
	return appdto.MapCollectorDataSources(raw), nil
}

func (s *service) GetDataSource(ctx context.Context, query GetDataSourceQuery) (appdto.CollectorDataSourceDTO, error) {
	companyID, err := s.requireCollector(ctx, query.CompanyFromCtx, query.TenantID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	raw, err := s.collector.GetDataSource(ctx, companyID, query.ID, query.CorrelationID)
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	return appdto.MapCollectorDataSourceRaw(raw), nil
}

func (s *service) CreateDataSource(ctx context.Context, cmd CreateDataSourceCommand) (appdto.CollectorDataSourceDTO, error) {
	if strings.TrimSpace(cmd.Name) == "" {
		return appdto.CollectorDataSourceDTO{}, pkg.NewAppError("MISSING_NAME", "name é obrigatório", http.StatusBadRequest)
	}
	if strings.TrimSpace(cmd.Slug) == "" {
		return appdto.CollectorDataSourceDTO{}, pkg.NewAppError("MISSING_SLUG", "slug é obrigatório", http.StatusBadRequest)
	}
	if strings.TrimSpace(cmd.CollectorType) == "" {
		return appdto.CollectorDataSourceDTO{}, pkg.NewAppError("MISSING_COLLECTOR_TYPE", "collectorType é obrigatório", http.StatusBadRequest)
	}
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	schedule, err := json.Marshal(appdto.MapCollectorScheduleDTO(cmd.DefaultSchedule))
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, pkg.NewAppError("INVALID_SCHEDULE", "defaultSchedule inválido", http.StatusBadRequest)
	}
	raw, err := s.collector.CreateDataSource(ctx, companyID, cmd.CorrelationID, appdto.CollectorDataSourceWriteRaw{
		Name:                cmd.Name,
		Slug:                cmd.Slug,
		Description:         optionalString(cmd.Description),
		WebsiteURL:          optionalString(cmd.WebsiteURL),
		CollectorType:       cmd.CollectorType,
		NameTemplate:        optionalString(cmd.NameTemplate),
		DescriptionTemplate: optionalString(cmd.DescriptionTemplate),
		PromptTemplate:      optionalString(cmd.PromptTemplate),
		DefaultContext:      optionalString(cmd.DefaultContext),
		DefaultSchedule:     schedule,
		ConfigTemplate:      cmd.ConfigTemplate,
		Variables:           cmd.Variables,
		Notes:               optionalString(cmd.Notes),
		Enabled:             cmd.Enabled,
		RateLimit:           cmd.RateLimit,
	})
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	return appdto.MapCollectorDataSourceRaw(raw), nil
}

func (s *service) UpdateDataSource(ctx context.Context, cmd UpdateDataSourceCommand) (appdto.CollectorDataSourceDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	write := appdto.CollectorDataSourceWriteRaw{
		Description:         cmd.Description,
		WebsiteURL:          cmd.WebsiteURL,
		NameTemplate:        cmd.NameTemplate,
		DescriptionTemplate: cmd.DescriptionTemplate,
		PromptTemplate:      cmd.PromptTemplate,
		DefaultContext:      cmd.DefaultContext,
		ConfigTemplate:      cmd.ConfigTemplate,
		Variables:           cmd.Variables,
		Notes:               cmd.Notes,
		RateLimit:           cmd.RateLimit,
	}
	if cmd.Name != nil {
		write.Name = *cmd.Name
	}
	if cmd.Slug != nil {
		write.Slug = *cmd.Slug
	}
	if cmd.DefaultSchedule != nil {
		schedule, marshalErr := json.Marshal(appdto.MapCollectorScheduleDTO(*cmd.DefaultSchedule))
		if marshalErr != nil {
			return appdto.CollectorDataSourceDTO{}, pkg.NewAppError("INVALID_SCHEDULE", "defaultSchedule inválido", http.StatusBadRequest)
		}
		write.DefaultSchedule = schedule
	}
	raw, err := s.collector.UpdateDataSource(ctx, companyID, cmd.ID, cmd.CorrelationID, write)
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	return appdto.MapCollectorDataSourceRaw(raw), nil
}

func (s *service) EnableDataSource(ctx context.Context, cmd DataSourceIDCommand) (appdto.CollectorDataSourceDTO, error) {
	return s.toggleDataSource(ctx, cmd, true)
}

func (s *service) DisableDataSource(ctx context.Context, cmd DataSourceIDCommand) (appdto.CollectorDataSourceDTO, error) {
	return s.toggleDataSource(ctx, cmd, false)
}

func (s *service) toggleDataSource(ctx context.Context, cmd DataSourceIDCommand, enable bool) (appdto.CollectorDataSourceDTO, error) {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	var raw appdto.CollectorDataSourceRaw
	if enable {
		raw, err = s.collector.EnableDataSource(ctx, companyID, cmd.ID, cmd.CorrelationID)
	} else {
		raw, err = s.collector.DisableDataSource(ctx, companyID, cmd.ID, cmd.CorrelationID)
	}
	if err != nil {
		return appdto.CollectorDataSourceDTO{}, err
	}
	return appdto.MapCollectorDataSourceRaw(raw), nil
}

func (s *service) DeleteDataSource(ctx context.Context, cmd DataSourceIDCommand) error {
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return err
	}
	return s.collector.DeleteDataSource(ctx, companyID, cmd.ID, cmd.CorrelationID)
}

func (s *service) PropagateDataSource(ctx context.Context, cmd PropagateDataSourceCommand) (appdto.PropagateDataSourceDTO, error) {
	if len(cmd.Fields) == 0 {
		return appdto.PropagateDataSourceDTO{}, pkg.NewAppError("MISSING_FIELDS", "fields é obrigatório", http.StatusBadRequest)
	}
	companyID, err := s.requireCollector(ctx, cmd.CompanyFromCtx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return appdto.PropagateDataSourceDTO{}, err
	}
	raw, err := s.collector.PropagateDataSource(ctx, companyID, cmd.ID, cmd.CorrelationID, appdto.PropagateDataSourceWriteRaw{
		Fields: cmd.Fields,
		DryRun: cmd.DryRun,
		Limit:  cmd.Limit,
	})
	if err != nil {
		return appdto.PropagateDataSourceDTO{}, err
	}
	return appdto.MapPropagateDataSourceRaw(raw), nil
}
