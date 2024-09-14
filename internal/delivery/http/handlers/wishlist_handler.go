package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type WishlistHandler struct {
	wishlistUseCase usecase.WishlistUseCase
}

func NewWishlistHandler(wishlistUseCase usecase.WishlistUseCase) *WishlistHandler {
	return &WishlistHandler{wishlistUseCase: wishlistUseCase}
}

func (h *WishlistHandler) AddToWishlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to add item to wishlist", nil, "User not authenticated")
		return
	}

	var input struct {
		ProductID int64 `json:"product_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to add item to wishlist", nil, "Invalid request body")
		return
	}

	if input.ProductID <= 0 {
		api.SendResponse(w, http.StatusBadRequest, "Failed to add item to wishlist", nil, "Invalid product ID")
		return
	}

	wishlistItem, err := h.wishlistUseCase.AddToWishlist(r.Context(), userID, input.ProductID)
	if err != nil {
		log.Printf("error : %v", err)
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to add item to wishlist", nil, "Product not found")
		case utils.ErrDuplicateWishlistItem:
			api.SendResponse(w, http.StatusConflict, "Failed to add item to wishlist", nil, "Product already in wishlist")
		case utils.ErrWishlistFull:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add item to wishlist", nil, "Wishlist is full")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to add item to wishlist", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Item added to wishlist successfully", wishlistItem, "")
}
