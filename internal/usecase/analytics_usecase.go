package usecase

import (
	"context"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type AnalyticsUseCase interface {
	GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int, sortBy string) ([]domain.TopProduct, error)
	GetTopCategories(ctx context.Context, startDate, endDate time.Time, limit int, sortOrder string) ([]domain.TopCategory, error)
	GetTopSubcategories(ctx context.Context, params domain.SubcategoryAnalyticsParams) ([]domain.TopSubcategory, error)
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

func (u *analyticsUseCase) GetTopCategories(ctx context.Context, startDate, endDate time.Time, limit int, sortOrder string) ([]domain.TopCategory, error) {
	if limit <= 0 || limit > 10 {
		limit = 10
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	params := domain.TopCategoriesParams{
		StartDate: startDate,
		EndDate:   endDate,
		Limit:     limit,
		SortOrder: sortOrder,
	}

	return u.analyticsRepo.GetTopCategories(ctx, params)
}

func (u *analyticsUseCase) GetTopSubcategories(ctx context.Context, params domain.SubcategoryAnalyticsParams) ([]domain.TopSubcategory, error) {
	// Validate and set default values
	if params.StartDate.IsZero() {
		params.StartDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}
	if params.EndDate.IsZero() {
		params.EndDate = time.Now()
	}
	if params.EndDate.After(time.Now()) {
		params.EndDate = time.Now()
	}
	if params.StartDate.After(params.EndDate) {
		return nil, utils.ErrInvalidDateRange
	}
	if params.Limit == 0 {
		params.Limit = 10
	}
	if params.Page == 0 {
		params.Page = 1
	}
	if params.SortBy != "revenue" && params.SortBy != "quantity" {
		params.SortBy = "quantity"
	}

	return u.analyticsRepo.GetTopSubcategories(ctx, params)
}
