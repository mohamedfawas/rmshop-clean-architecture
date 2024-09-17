package usecase

import (
	"context"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
)

type AnalyticsUseCase interface {
	GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int, sortBy string) ([]domain.TopProduct, error)
}

type analyticsUseCase struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsUseCase(analyticsRepo repository.AnalyticsRepository) AnalyticsUseCase {
	return &analyticsUseCase{analyticsRepo: analyticsRepo}
}

func (u *analyticsUseCase) GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int, sortBy string) ([]domain.TopProduct, error) {
	// If no date range is specified, use the last 30 days
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	return u.analyticsRepo.GetTopProducts(ctx, startDate, endDate, limit, sortBy)
}
