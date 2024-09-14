package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

func (h *WishlistHandler) RemoveFromWishlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to remove item from wishlist", nil, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to remove item from wishlist", nil, "Invalid product ID")
		return
	}

	isEmpty, err := h.wishlistUseCase.RemoveFromWishlist(r.Context(), userID, productID)
	if err != nil {
		switch err {
		case utils.ErrProductNotInWishlist:
			api.SendResponse(w, http.StatusNotFound, "Failed to remove item from wishlist", nil, "Product not found in wishlist")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to remove item from wishlist", nil, "You don't have permission to remove this item")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to remove item from wishlist", nil, "An unexpected error occurred")
		}
		return
	}

	response := map[string]interface{}{
		"message": "Product successfully removed from wishlist",
	}
	if isEmpty {
		response["wishlist_status"] = "empty"
	}

	api.SendResponse(w, http.StatusOK, "Item removed from wishlist successfully", response, "")
}

func (h *WishlistHandler) GetUserWishlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to retrieve wishlist", nil, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	wishlistItems, totalCount, err := h.wishlistUseCase.GetUserWishlist(r.Context(), userID, page, limit, sortBy, order)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to retrieve wishlist", nil, "User not found")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve wishlist", nil, "An unexpected error occurred")
		}
		return
	}

	response := map[string]interface{}{
		"items":       wishlistItems,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": (totalCount + int64(limit) - 1) / int64(limit),
	}

	api.SendResponse(w, http.StatusOK, "Wishlist retrieved successfully", response, "")
}
