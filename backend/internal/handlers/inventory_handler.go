package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

// InventoryHandler handles HTTP requests for inventory
type InventoryHandler struct {
	inventoryRepo *repository.InventoryRepository
	productRepo   *repository.ProductRepository
}

// NewInventoryHandler creates a new inventory handler with the provided repositories
func NewInventoryHandler(inventoryRepo *repository.InventoryRepository, productRepo *repository.ProductRepository) *InventoryHandler {
	return &InventoryHandler{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo,
	}
}

// GetAllInventory returns all inventory items
func (h *InventoryHandler) GetAllInventory(c echo.Context) error {
	ctx := c.Request().Context()

	inventory, err := h.inventoryRepo.GetAll(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve inventory items",
		})
	}

	return c.JSON(http.StatusOK, inventory)
}

// GetInventoryByID returns an inventory item by ID
func (h *InventoryHandler) GetInventoryByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid inventory ID",
		})
	}

	inventory, err := h.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "inventory item not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Inventory item not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve inventory item",
		})
	}

	return c.JSON(http.StatusOK, inventory)
}

// GetInventoryByProductID returns inventory for a specific product
func (h *InventoryHandler) GetInventoryByProductID(c echo.Context) error {
	ctx := c.Request().Context()

	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	// First check if product exists
	_, err = h.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify product",
		})
	}

	inventory, err := h.inventoryRepo.GetByProductID(ctx, productID)
	if err != nil {
		if err.Error() == "inventory for product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Inventory for product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve inventory",
		})
	}

	return c.JSON(http.StatusOK, inventory)
}

// CreateInventory creates a new inventory item
func (h *InventoryHandler) CreateInventory(c echo.Context) error {
	ctx := c.Request().Context()

	var inventory models.Inventory
	if err := c.Bind(&inventory); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate required fields and values
	if inventory.ProductID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Valid product ID is required",
		})
	}

	if inventory.CurrentStock < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Current stock cannot be negative",
		})
	}

	if inventory.ReorderLevel < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Reorder level cannot be negative",
		})
	}

	// Verify product exists
	_, err := h.productRepo.GetByID(ctx, inventory.ProductID)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify product",
		})
	}

	err = h.inventoryRepo.Create(ctx, &inventory)
	if err != nil {
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Inventory for this product already exists",
			})
		}
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create inventory item",
		})
	}

	return c.JSON(http.StatusCreated, inventory)
}

// UpdateInventory updates an existing inventory item
func (h *InventoryHandler) UpdateInventory(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid inventory ID",
		})
	}

	var inventory models.Inventory
	if err := c.Bind(&inventory); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Ensure ID in path matches ID in payload
	inventory.InventoryID = id

	// Validate required fields and values
	if inventory.ProductID <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Valid product ID is required",
		})
	}

	if inventory.CurrentStock < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Current stock cannot be negative",
		})
	}

	if inventory.ReorderLevel < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Reorder level cannot be negative",
		})
	}

	err = h.inventoryRepo.Update(ctx, &inventory)
	if err != nil {
		if err.Error() == "inventory item not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Inventory item not found",
			})
		}
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Inventory with this information already exists",
			})
		}
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update inventory item",
		})
	}

	return c.JSON(http.StatusOK, inventory)
}

// UpdateStock updates just the stock level of an inventory item
func (h *InventoryHandler) UpdateStock(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid inventory ID",
		})
	}

	// Simple payload with just the new stock level
	var stockUpdate struct {
		CurrentStock int `json:"current_stock"`
	}

	if err := c.Bind(&stockUpdate); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	if stockUpdate.CurrentStock < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Current stock cannot be negative",
		})
	}

	err = h.inventoryRepo.UpdateStock(ctx, id, stockUpdate.CurrentStock)
	if err != nil {
		if err.Error() == "inventory item not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Inventory item not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update stock level",
		})
	}

	// Get the updated inventory item to return
	inventory, err := h.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Stock updated but failed to retrieve updated inventory",
		})
	}

	return c.JSON(http.StatusOK, inventory)
}

// DeleteInventory deletes an inventory item
func (h *InventoryHandler) DeleteInventory(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid inventory ID",
		})
	}

	err = h.inventoryRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "inventory item not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Inventory item not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete inventory item",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetLowStockItems returns inventory items that are low on stock
func (h *InventoryHandler) GetLowStockItems(c echo.Context) error {
	ctx := c.Request().Context()

	inventory, err := h.inventoryRepo.GetLowStockItems(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve low stock items",
		})
	}

	return c.JSON(http.StatusOK, inventory)
}

// GetLowStockWithProductInfo returns low stock items with product details
func (h *InventoryHandler) GetLowStockWithProductInfo(c echo.Context) error {
	ctx := c.Request().Context()

	items, err := h.inventoryRepo.GetLowStockWithProductInfo(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve low stock items with product info",
		})
	}

	return c.JSON(http.StatusOK, items)
} 