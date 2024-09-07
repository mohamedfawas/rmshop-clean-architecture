package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CartHandler struct {
	cartUseCase usecase.CartUseCase
}

func NewCartHandler(cartUseCase usecase.CartUseCase) *CartHandler {
	return &CartHandler{cartUseCase: cartUseCase}
}

func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	// extract the user id from the context key
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to add item to cart", nil, "User not authenticated")
		return
	}

	var input domain.AddToCartInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to add item to cart", nil, "Invalid request body")
		return
	}

	cartItem, err := h.cartUseCase.AddToCart(r.Context(), userID, input.ProductID, input.Quantity)
	if err != nil {
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to add item to cart", nil, "Product not found")
		case utils.ErrInvalidQuantity:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add item to cart", nil, "Invalid quantity")
		case utils.ErrExceedsMaxQuantity:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add item to cart", nil, "Maximum quantity limit is 10")
		case utils.ErrInsufficientStock:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add item to cart", nil, "Insufficient stock")
		case utils.ErrCartFull:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add item to cart", nil, "Cart is full")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to add item to cart", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Item added to cart successfully", cartItem, "")
}

func (h *CartHandler) GetUserCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to retrieve cart", nil, "User not authenticated")
		return
	}

	cart, err := h.cartUseCase.GetUserCart(r.Context(), userID)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to retrieve cart", nil, "User not found")
		case utils.ErrUserBlocked:
			api.SendResponse(w, http.StatusForbidden, "Failed to retrieve cart", nil, "User account is blocked")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve cart", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Cart retrieved successfully", cart, "")
}

func (h *CartHandler) UpdateCartItemQuantity(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to update cart item", nil, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	itemID, err := strconv.ParseInt(vars["itemId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update cart item", nil, "Invalid item ID")
		return
	}

	var input struct {
		Quantity int `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update cart item", nil, "Invalid request body")
		return
	}

	err = h.cartUseCase.UpdateCartItemQuantity(r.Context(), userID, itemID, input.Quantity)
	if err != nil {
		switch err {
		case utils.ErrCartItemNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update cart item", nil, "Cart item not found")
		case utils.ErrInvalidQuantity:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update cart item", nil, "Invalid quantity")
		case utils.ErrExceedsMaxQuantity:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update cart item", nil, "Maximum quantity is limited to 10")
		case utils.ErrInsufficientStock:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update cart item", nil, "Insufficient stock")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to update cart item", nil, "Unauthorized access to this cart item")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update cart item", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Cart item updated successfully", nil, "")
}

func (h *CartHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	// extract the user id from the context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to delete item from cart", nil, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	itemID, err := strconv.ParseInt(vars["itemId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to delete item from cart", nil, "Invalid item ID")
		return
	}

	err = h.cartUseCase.DeleteCartItem(r.Context(), userID, itemID)
	if err != nil {
		switch err {
		case utils.ErrCartItemNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete item from cart", nil, "Item not found in cart")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to delete item from cart", nil, "You don't have permission to remove this item")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to delete item from cart", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Item removed from cart successfully", nil, "")
}
