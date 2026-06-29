package workorder

import (
	"context"

	"github.com/jackc/pgx/v5"
	"optistock/pkg/apperror"
)

func proposeAllocationForRequirement(ctx context.Context, tx pgx.Tx, repo *Repository, req Requirement) (AllocationProposal, error) {
	proposal := AllocationProposal{
		ReqID:            req.ReqID,
		RequiredItemCode: req.RequiredItemCode,
		TotalNeededQty:   req.TotalNeededQty,
		Proposed:         []ReserveAllocationInput{},
	}

	lots, err := repo.ListEligibleLotsForUpdate(ctx, tx, req.RequiredItemCode)
	if err != nil {
		return proposal, err
	}

	remaining := req.TotalNeededQty
	for _, lot := range lots {
		if cmp, ok := cmpRat(remaining, "0"); ok && cmp <= 0 {
			break
		}
		take := lot.AvailableQty
		if cmp, ok := cmpRat(take, remaining); ok && cmp > 0 {
			take = remaining
		}
		if !isPositive(take) {
			continue
		}
		proposal.Proposed = append(proposal.Proposed, ReserveAllocationInput{
			ReqID:         req.ReqID,
			LOTInternalID: lot.LOTInternalID,
			ReservedQty:   take,
		})
		var ok bool
		remaining, ok = subRat(remaining, take)
		if !ok {
			return proposal, apperror.ErrInternal
		}
	}

	if cmp, ok := cmpRat(remaining, "0"); ok && cmp > 0 {
		proposal.Shortage = true
		proposal.ShortageQty = remaining
	}
	return proposal, nil
}

func proposeAllAllocations(ctx context.Context, tx pgx.Tx, repo *Repository, requirements []Requirement) ([]AllocationProposal, error) {
	out := make([]AllocationProposal, 0, len(requirements))
	for _, req := range requirements {
		p, err := proposeAllocationForRequirement(ctx, tx, repo, req)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func applyAllocations(ctx context.Context, tx pgx.Tx, repo *Repository, allocations []ReserveAllocationInput) error {
	for _, a := range allocations {
		if a.ReqID == "" || a.LOTInternalID == "" || !isPositive(a.ReservedQty) {
			continue
		}
		itemCode, err := repo.GetLotItemCode(ctx, tx, a.LOTInternalID)
		if err != nil {
			return err
		}
		req, err := repo.GetRequirementByID(ctx, a.ReqID)
		if err != nil || req == nil {
			return err
		}
		if itemCode != req.RequiredItemCode {
			return apperror.WithMessage(apperror.ErrValidation, "lot does not match required item")
		}
		if err := repo.UpsertAllocation(ctx, tx, a.ReqID, a.LOTInternalID, a.ReservedQty); err != nil {
			return err
		}
	}
	return nil
}

func validateRequirementCoverage(ctx context.Context, tx pgx.Tx, repo *Repository, requirements []Requirement) ([]Requirement, error) {
	flagged := make([]Requirement, 0, len(requirements))
	for _, req := range requirements {
		reserved, err := repo.SumReservedForReq(ctx, tx, req.ReqID)
		if err != nil {
			return nil, err
		}
		if cmp, ok := cmpRat(reserved, req.TotalNeededQty); ok && cmp < 0 {
			short := req
			short.Shortage = true
			shortageQty, _ := subRat(req.TotalNeededQty, reserved)
			short.AvailableQty = shortageQty
			flagged = append(flagged, short)
		}
	}
	return flagged, nil
}
