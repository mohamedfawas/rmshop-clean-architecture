package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CheckoutUseCase interface {
	CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
}

type checkoutUseCase struct {
	checkoutRepo repository.CheckoutRepository
	productRepo  repository.ProductRepository
}

func NewCheckoutUseCase(checkoutRepo repository.CheckoutRepository, productRepo repository.ProductRepository) CheckoutUseCase {
	return &checkoutUseCase{
		checkoutRepo: checkoutRepo,
		productRepo:  productRepo,
	}
}

func (u *checkoutUseCase) CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// Get cart items
	cartItems, err := u.checkoutRepo.GetCartItems(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Validate items and calculate total
	var totalAmount float64
	var checkoutItems []*domain.CheckoutItem

	for _, item := range cartItems {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}

		subtotal := float64(item.Quantity) * product.Price
		checkoutItems = append(checkoutItems, &domain.CheckoutItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Subtotal:  subtotal,
		})

		totalAmount += subtotal
	}

	// Create checkout session
	session, err := u.checkoutRepo.CreateCheckoutSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Add items to checkout session
	err = u.checkoutRepo.AddCheckoutItems(ctx, session.ID, checkoutItems)
	if err != nil {
		return nil, err
	}

	session.TotalAmount = totalAmount
	session.ItemCount = len(checkoutItems)

	return session, nil
}
