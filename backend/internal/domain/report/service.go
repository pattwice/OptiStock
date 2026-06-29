package report

import (
	"context"
	"strings"

	"optistock/pkg/export"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) StockOnHand(ctx context.Context, f StockOnHandFilters) (*PageResult, error) {
	rows, total, err := s.repo.StockOnHand(ctx, f)
	if err != nil {
		return nil, err
	}
	return &PageResult{Headers: HeadersStockOnHand(), Rows: rows, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Service) MovementLedger(ctx context.Context, f MovementLedgerFilters) (*PageResult, error) {
	rows, total, err := s.repo.MovementLedger(ctx, f)
	if err != nil {
		return nil, err
	}
	return &PageResult{Headers: HeadersMovementLedger(), Rows: rows, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Service) WOSummary(ctx context.Context, f WOSummaryFilters) (*PageResult, error) {
	rows, total, err := s.repo.WOSummary(ctx, f)
	if err != nil {
		return nil, err
	}
	return &PageResult{Headers: HeadersWOSummary(), Rows: rows, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Service) ShortageDamage(ctx context.Context, f ShortageDamageFilters) (*PageResult, error) {
	rows, total, err := s.repo.ShortageDamage(ctx, f)
	if err != nil {
		return nil, err
	}
	return &PageResult{Headers: HeadersShortageDamage(), Rows: rows, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Service) AuditTrail(ctx context.Context, f AuditTrailFilters) (*PageResult, error) {
	rows, total, err := s.repo.AuditTrail(ctx, f)
	if err != nil {
		return nil, err
	}
	return &PageResult{Headers: HeadersAuditTrail(), Rows: rows, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Service) PartialCompletion(ctx context.Context, f PartialCompletionFilters) (*PageResult, error) {
	rows, total, err := s.repo.PartialCompletion(ctx, f)
	if err != nil {
		return nil, err
	}
	return &PageResult{Headers: HeadersPartialCompletion(), Rows: rows, Total: total, Page: f.Page, PageSize: f.PageSize}, nil
}

func (s *Service) Export(ctx context.Context, reportType, format string, fetch func(context.Context) (*PageResult, error)) ([]byte, string, string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "csv"
	}
	// Export ignores pagination — fetch all with large page size.
	result, err := fetch(ctx)
	if err != nil {
		return nil, "", "", err
	}
	table := export.Table{
		SheetName: reportType,
		Headers:   result.Headers,
		Rows:      result.Rows,
	}
	var data []byte
	switch format {
	case "xlsx", "excel":
		data, err = export.ToExcel(table)
		format = "xlsx"
	default:
		data, err = export.ToCSV(table)
		format = "csv"
	}
	if err != nil {
		return nil, "", "", err
	}
	return data, export.Filename(reportType, format), export.ContentType(format), nil
}

func (s *Service) ExportStockOnHand(ctx context.Context, f StockOnHandFilters, format string) ([]byte, string, string, error) {
	f.Page = 1
	f.PageSize = 10000
	return s.Export(ctx, "stock-on-hand", format, func(ctx context.Context) (*PageResult, error) {
		return s.StockOnHand(ctx, f)
	})
}

func (s *Service) ExportMovementLedger(ctx context.Context, f MovementLedgerFilters, format string) ([]byte, string, string, error) {
	f.Page = 1
	f.PageSize = 10000
	return s.Export(ctx, "movement-ledger", format, func(ctx context.Context) (*PageResult, error) {
		return s.MovementLedger(ctx, f)
	})
}

func (s *Service) ExportWOSummary(ctx context.Context, f WOSummaryFilters, format string) ([]byte, string, string, error) {
	f.Page = 1
	f.PageSize = 10000
	return s.Export(ctx, "wo-summary", format, func(ctx context.Context) (*PageResult, error) {
		return s.WOSummary(ctx, f)
	})
}

func (s *Service) ExportShortageDamage(ctx context.Context, f ShortageDamageFilters, format string) ([]byte, string, string, error) {
	f.Page = 1
	f.PageSize = 10000
	return s.Export(ctx, "shortage-damage", format, func(ctx context.Context) (*PageResult, error) {
		return s.ShortageDamage(ctx, f)
	})
}

func (s *Service) ExportAuditTrail(ctx context.Context, f AuditTrailFilters, format string) ([]byte, string, string, error) {
	f.Page = 1
	f.PageSize = 10000
	return s.Export(ctx, "audit-trail", format, func(ctx context.Context) (*PageResult, error) {
		return s.AuditTrail(ctx, f)
	})
}

func (s *Service) ExportPartialCompletion(ctx context.Context, f PartialCompletionFilters, format string) ([]byte, string, string, error) {
	f.Page = 1
	f.PageSize = 10000
	return s.Export(ctx, "partial-completion", format, func(ctx context.Context) (*PageResult, error) {
		return s.PartialCompletion(ctx, f)
	})
}
