package workorder

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"optistock/pkg/apperror"
)

func explodeBOM(ctx context.Context, repo *Repository, parentCode, multiplier string) (map[string]string, error) {
	hasBOM, err := repo.HasActiveBOM(ctx, parentCode)
	if err != nil {
		return nil, err
	}
	if !hasBOM {
		return nil, apperror.WithMessage(apperror.ErrValidation, fmt.Sprintf("no active BOM found for %s", parentCode))
	}

	rows, err := repo.ListActiveBOM(ctx, parentCode)
	if err != nil {
		return nil, err
	}

	needs := make(map[string]string)
	for _, row := range rows {
		componentQty, ok := mulRat(multiplier, row.QtyPerSet)
		if !ok {
			return nil, apperror.WithMessage(apperror.ErrValidation, "invalid BOM quantity")
		}

		switch row.ComponentType {
		case "RM":
			if existing, ok := needs[row.ComponentItemCode]; ok {
				sum, ok := addRat(existing, componentQty)
				if !ok {
					return nil, apperror.ErrInternal
				}
				needs[row.ComponentItemCode] = sum
			} else {
				needs[row.ComponentItemCode] = componentQty
			}
		case "SFG", "FG":
			nested, err := explodeBOM(ctx, repo, row.ComponentItemCode, componentQty)
			if err != nil {
				return nil, err
			}
			for code, qty := range nested {
				if existing, ok := needs[code]; ok {
					sum, ok := addRat(existing, qty)
					if !ok {
						return nil, apperror.ErrInternal
					}
					needs[code] = sum
				} else {
					needs[code] = qty
				}
			}
		default:
			return nil, apperror.WithMessage(apperror.ErrValidation, fmt.Sprintf("unsupported item type %s in BOM", row.ComponentType))
		}
	}
	return needs, nil
}

func runBOMExplosionTx(ctx context.Context, tx pgx.Tx, repo *Repository, woNumber, targetFGCode, targetQty string) error {
	needs, err := explodeBOM(ctx, repo, targetFGCode, targetQty)
	if err != nil {
		return err
	}
	if len(needs) == 0 {
		return apperror.WithMessage(apperror.ErrValidation, "BOM explosion produced no RM requirements")
	}
	for itemCode, qty := range needs {
		if _, err := repo.InsertRequirement(ctx, tx, woNumber, itemCode, qty); err != nil {
			return err
		}
	}
	return nil
}
