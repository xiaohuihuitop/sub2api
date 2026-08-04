package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *userSubscriptionRepository) ListActiveByUserIDAndGroupID(
	ctx context.Context,
	userID, groupID int64,
) ([]service.UserSubscription, error) {
	return r.listActiveByUserID(ctx, userID, usersubscription.GroupIDEQ(groupID))
}

func (r *userSubscriptionRepository) ListActiveByUserIDAndGroupIDs(
	ctx context.Context,
	userID int64,
	groupIDs []int64,
) ([]service.UserSubscription, error) {
	if len(groupIDs) == 0 {
		return []service.UserSubscription{}, nil
	}
	return r.listActiveByUserID(ctx, userID, usersubscription.GroupIDIn(groupIDs...))
}

func (r *userSubscriptionRepository) listActiveByUserID(
	ctx context.Context,
	userID int64,
	groupPredicate predicate.UserSubscription,
) ([]service.UserSubscription, error) {
	client := clientFromContext(ctx, r.client)
	entities, err := client.UserSubscription.Query().
		Where(
			usersubscription.UserIDEQ(userID),
			usersubscription.StatusEQ(service.SubscriptionStatusActive),
			usersubscription.ExpiresAtGT(time.Now()),
			groupPredicate,
		).
		WithGroup().
		Order(
			dbent.Asc(usersubscription.FieldExpiresAt),
			dbent.Asc(usersubscription.FieldCreatedAt),
			dbent.Asc(usersubscription.FieldID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return userSubscriptionEntitiesToService(entities), nil
}
