package item

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

func (r *Repository) CreateItem(ctx context.Context, input CreateItemInput) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO a1_items (
			item_code, description, unit, item_type, min_stock_level,
			expiry_threshold_days, shelf_life_days
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING item_code, description, unit, item_type, min_stock_level::text,
		          expiry_threshold_days, shelf_life_days, created_at, updated_at
	`,
		input.ItemCode, input.Description, input.Unit, input.ItemType, input.MinStockLevel,
		input.ExpiryThresholdDays, input.ShelfLifeDays,
	)

	var it Item
	if err := row.Scan(
		&it.ItemCode, &it.Description, &it.Unit, &it.ItemType, &it.MinStockLevel,
		&it.ExpiryThresholdDays, &it.ShelfLifeDays, &it.CreatedAt, &it.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *Repository) UpdateItem(ctx context.Context, itemCode string, input UpdateItemInput) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE a1_items
		SET
			description = COALESCE($2, description),
			unit = COALESCE($3, unit),
			item_type = COALESCE($4, item_type),
			min_stock_level = COALESCE($5::numeric, min_stock_level),
			expiry_threshold_days = CASE WHEN $6::boolean THEN $7::int ELSE expiry_threshold_days END,
			shelf_life_days = CASE WHEN $8::boolean THEN $9::int ELSE shelf_life_days END,
			updated_at = NOW()
		WHERE item_code = $1
		RETURNING item_code, description, unit, item_type, min_stock_level::text,
		          expiry_threshold_days, shelf_life_days, created_at, updated_at
	`,
		itemCode,
		input.Description,
		input.Unit,
		input.ItemType,
		input.MinStockLevel,
		input.ExpiryThresholdDays != nil, derefIntPtrPtr(input.ExpiryThresholdDays),
		input.ShelfLifeDays != nil, derefIntPtrPtr(input.ShelfLifeDays),
	)

	var it Item
	if err := row.Scan(
		&it.ItemCode, &it.Description, &it.Unit, &it.ItemType, &it.MinStockLevel,
		&it.ExpiryThresholdDays, &it.ShelfLifeDays, &it.CreatedAt, &it.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &it, nil
}

func (r *Repository) GetItem(ctx context.Context, itemCode string) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT item_code, description, unit, item_type, min_stock_level::text,
		       expiry_threshold_days, shelf_life_days, created_at, updated_at
		FROM a1_items
		WHERE item_code = $1
	`, itemCode)

	var it Item
	if err := row.Scan(
		&it.ItemCode, &it.Description, &it.Unit, &it.ItemType, &it.MinStockLevel,
		&it.ExpiryThresholdDays, &it.ShelfLifeDays, &it.CreatedAt, &it.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &it, nil
}

func (r *Repository) ListItems(ctx context.Context, itemType string) ([]Item, error) {
	var rows pgx.Rows
	var err error

	if itemType != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT item_code, description, unit, item_type, min_stock_level::text,
			       expiry_threshold_days, shelf_life_days, created_at, updated_at
			FROM a1_items
			WHERE item_type = $1
			ORDER BY item_code ASC
		`, itemType)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT item_code, description, unit, item_type, min_stock_level::text,
			       expiry_threshold_days, shelf_life_days, created_at, updated_at
			FROM a1_items
			ORDER BY item_code ASC
		`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Item, 0, 64)
	for rows.Next() {
		var it Item
		if err := rows.Scan(
			&it.ItemCode, &it.Description, &it.Unit, &it.ItemType, &it.MinStockLevel,
			&it.ExpiryThresholdDays, &it.ShelfLifeDays, &it.CreatedAt, &it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Repository) CreateBOMRow(ctx context.Context, input CreateBOMRowInput) (*BOMRow, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO a2_bom (
			parent_item_code, component_item_code, qty_per_set, bom_version, is_active
		)
		VALUES ($1,$2,$3::numeric,$4,$5)
		RETURNING bom_id, parent_item_code, component_item_code, qty_per_set::text, bom_version, is_active, created_at
	`,
		input.ParentItemCode, input.ComponentItemCode, input.QtyPerSet, input.BomVersion, input.IsActive,
	)

	var b BOMRow
	if err := row.Scan(&b.BomID, &b.ParentItemCode, &b.ComponentItemCode, &b.QtyPerSet, &b.BomVersion, &b.IsActive, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) ListBOMByParent(ctx context.Context, parentItemCode string) ([]BOMRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT bom_id, parent_item_code, component_item_code, qty_per_set::text, bom_version, is_active, created_at
		FROM a2_bom
		WHERE parent_item_code = $1
		ORDER BY bom_version DESC, created_at ASC
	`, parentItemCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]BOMRow, 0, 64)
	for rows.Next() {
		var b BOMRow
		if err := rows.Scan(&b.BomID, &b.ParentItemCode, &b.ComponentItemCode, &b.QtyPerSet, &b.BomVersion, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Repository) ActivateBOMVersion(ctx context.Context, parentItemCode, bomVersion string) (activatedCount int64, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if _, err = tx.Exec(ctx, `
		UPDATE a2_bom
		SET is_active = FALSE
		WHERE parent_item_code = $1 AND is_active = TRUE
	`, parentItemCode); err != nil {
		return 0, err
	}

	tag, err := tx.Exec(ctx, `
		UPDATE a2_bom
		SET is_active = TRUE
		WHERE parent_item_code = $1 AND bom_version = $2
	`, parentItemCode, bomVersion)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func derefIntPtrPtr(v **int) any {
	if v == nil {
		return nil
	}
	return *v
}

