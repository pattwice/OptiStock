package approval

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) InsertPending(ctx context.Context, tx pgx.Tx, woNumber, userID, completionPct string) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO d2_wo_approvals (
			wo_number, approval_type, requested_by, completion_pct_at_request, approval_status
		) VALUES ($1, 'PARTIAL_COMPLETION', $2::uuid, $3::numeric, 'PENDING')
		RETURNING approval_id::text
	`, woNumber, userID, completionPct).Scan(&id)
	return id, err
}

func (r *Repository) GetByID(ctx context.Context, approvalID string) (*Approval, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			a.approval_id::text, a.wo_number, a.approval_type, a.requested_by::text, u.name,
			a.requested_at, a.completion_pct_at_request::text, a.approval_status,
			a.resolved_by::text, ru.name, a.resolved_at, a.resolution_notes,
			w.target_fg_code, w.target_qty::text, w.actual_produced_qty::text
		FROM d2_wo_approvals a
		JOIN users u ON u.id = a.requested_by
		JOIN c1_wo_header w ON w.wo_number = a.wo_number
		LEFT JOIN users ru ON ru.id = a.resolved_by
		WHERE a.approval_id = $1::uuid
	`, approvalID)
	return scanApproval(row)
}

func (r *Repository) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, approvalID string) (*Approval, error) {
	row := tx.QueryRow(ctx, `
		SELECT
			a.approval_id::text, a.wo_number, a.approval_type, a.requested_by::text, u.name,
			a.requested_at, a.completion_pct_at_request::text, a.approval_status,
			a.resolved_by::text, ru.name, a.resolved_at, a.resolution_notes,
			w.target_fg_code, w.target_qty::text, w.actual_produced_qty::text
		FROM d2_wo_approvals a
		JOIN users u ON u.id = a.requested_by
		JOIN c1_wo_header w ON w.wo_number = a.wo_number
		LEFT JOIN users ru ON ru.id = a.resolved_by
		WHERE a.approval_id = $1::uuid
		FOR UPDATE OF a
	`, approvalID)
	return scanApproval(row)
}

func (r *Repository) Resolve(ctx context.Context, tx pgx.Tx, approvalID string, status ApprovalStatus, resolvedBy *string, notes *string) error {
	_, err := tx.Exec(ctx, `
		UPDATE d2_wo_approvals
		SET approval_status = $2,
		    resolved_by = $3::uuid,
		    resolved_at = NOW(),
		    resolution_notes = $4
		WHERE approval_id = $1::uuid
	`, approvalID, status, resolvedBy, notes)
	return err
}

func (r *Repository) WithdrawPendingForWO(ctx context.Context, tx pgx.Tx, woNumber string, resolvedBy *string, notes string) error {
	_, err := tx.Exec(ctx, `
		UPDATE d2_wo_approvals
		SET approval_status = 'WITHDRAWN',
		    resolved_by = $2::uuid,
		    resolved_at = NOW(),
		    resolution_notes = $3
		WHERE wo_number = $1 AND approval_status = 'PENDING'
	`, woNumber, resolvedBy, notes)
	return err
}

func (r *Repository) ListPending(ctx context.Context) ([]Approval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			a.approval_id::text, a.wo_number, a.approval_type, a.requested_by::text, u.name,
			a.requested_at, a.completion_pct_at_request::text, a.approval_status,
			a.resolved_by::text, ru.name, a.resolved_at, a.resolution_notes,
			w.target_fg_code, w.target_qty::text, w.actual_produced_qty::text
		FROM d2_wo_approvals a
		JOIN users u ON u.id = a.requested_by
		JOIN c1_wo_header w ON w.wo_number = a.wo_number
		LEFT JOIN users ru ON ru.id = a.resolved_by
		WHERE a.approval_status = 'PENDING'
		ORDER BY a.requested_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApprovalRows(rows)
}

func (r *Repository) ListByWO(ctx context.Context, woNumber string) ([]Approval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			a.approval_id::text, a.wo_number, a.approval_type, a.requested_by::text, u.name,
			a.requested_at, a.completion_pct_at_request::text, a.approval_status,
			a.resolved_by::text, ru.name, a.resolved_at, a.resolution_notes,
			w.target_fg_code, w.target_qty::text, w.actual_produced_qty::text
		FROM d2_wo_approvals a
		JOIN users u ON u.id = a.requested_by
		JOIN c1_wo_header w ON w.wo_number = a.wo_number
		LEFT JOIN users ru ON ru.id = a.resolved_by
		WHERE a.wo_number = $1
		ORDER BY a.requested_at ASC
	`, woNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanApprovalRows(rows)
}

func (r *Repository) HasPendingForWO(ctx context.Context, tx pgx.Tx, woNumber string) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM d2_wo_approvals
			WHERE wo_number = $1 AND approval_status = 'PENDING'
		)
	`, woNumber).Scan(&exists)
	return exists, err
}

func scanApproval(row pgx.Row) (*Approval, error) {
	var a Approval
	if err := row.Scan(
		&a.ApprovalID, &a.WONumber, &a.ApprovalType, &a.RequestedBy, &a.RequestedByName,
		&a.RequestedAt, &a.CompletionPctAtRequest, &a.ApprovalStatus,
		&a.ResolvedBy, &a.ResolvedByName, &a.ResolvedAt, &a.ResolutionNotes,
		&a.TargetFGCode, &a.TargetQty, &a.ActualProducedQty,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func scanApprovalRows(rows pgx.Rows) ([]Approval, error) {
	out := make([]Approval, 0, 16)
	for rows.Next() {
		var a Approval
		if err := rows.Scan(
			&a.ApprovalID, &a.WONumber, &a.ApprovalType, &a.RequestedBy, &a.RequestedByName,
			&a.RequestedAt, &a.CompletionPctAtRequest, &a.ApprovalStatus,
			&a.ResolvedBy, &a.ResolvedByName, &a.ResolvedAt, &a.ResolutionNotes,
			&a.TargetFGCode, &a.TargetQty, &a.ActualProducedQty,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
