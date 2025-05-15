package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	orderRepo *repository.OrderRepository
}

// NewOrderHandler creates a new order handler with the provided repository
func NewOrderHandler(orderRepo *repository.OrderRepository) *OrderHandler {
	return &OrderHandler{
		orderRepo: orderRepo,
	}
}

// GetAllOrders returns all orders
func (h *OrderHandler) GetAllOrders(c echo.Context) error {
	ctx := c.Request().Context()

	orders, err := h.orderRepo.GetAll(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve orders",
		})
	}

	return c.JSON(http.StatusOK, orders)
}

// GetOrderByID returns an order by ID
func (h *OrderHandler) GetOrderByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid order ID",
		})
	}

	order, err := h.orderRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "order not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Order not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve order",
		})
	}

	// Get order items
	items, err := h.orderRepo.GetOrderItems(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve order items",
		})
	}

	// Return order with items
	return c.JSON(http.StatusOK, map[string]interface{}{
		"order": order,
		"items": items,
	})
}

// CreateOrderRequest represents the structure of the JSON payload for creating orders
type CreateOrderRequest struct {
	Order     models.Order       `json:"order"`
	Items     []models.OrderItem `json:"items"`
	Quotation *struct {
		QuotationID int `json:"quotation_id"`
	} `json:"quotation,omitempty"`
}

// CreateOrder creates a new order with items
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	ctx := c.Request().Context()

	// Define a struct to receive the order data with items
	var orderData CreateOrderRequest

	if err := c.Bind(&orderData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload: " + err.Error(),
		})
	}

	// Validate required fields
	if orderData.Order.CustomerID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Customer ID is required",
		})
	}

	if len(orderData.Items) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Order must have at least one item",
		})
	}

	// If the request includes a quotation reference, set the quotation ID in the order
	if orderData.Quotation != nil && orderData.Quotation.QuotationID > 0 {
		quotationID := orderData.Quotation.QuotationID
		orderData.Order.QuotationID = &quotationID
	}

	// Create the order with items in a single transaction
	err := h.orderRepo.CreateOrderWithItems(ctx, &orderData.Order, orderData.Items)
	if err != nil {
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "An order with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create order: " + err.Error(),
		})
	}

	// Return the created order with items
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"order": orderData.Order,
		"items": orderData.Items,
	})
}

// UpdateOrder updates an existing order
func (h *OrderHandler) UpdateOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid order ID",
		})
	}

	var order models.Order
	if err := c.Bind(&order); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Ensure ID in path matches ID in payload
	order.OrderID = id

	// Validate required fields
	if order.CustomerID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Customer ID is required",
		})
	}

	err = h.orderRepo.Update(ctx, &order)
	if err != nil {
		if err.Error() == "order not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Order not found",
			})
		}
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "An order with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update order",
		})
	}

	return c.JSON(http.StatusOK, order)
}

// DeleteOrder deletes an order
func (h *OrderHandler) DeleteOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid order ID",
		})
	}

	err = h.orderRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "order not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Order not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete order",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// StatusUpdate represents the status update request
type StatusUpdate struct {
	Status string `json:"status"`
}

// UpdateOrderStatus updates just the status of an order
func (h *OrderHandler) UpdateOrderStatus(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid order ID",
		})
	}

	var statusUpdate StatusUpdate
	if err := c.Bind(&statusUpdate); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate required field
	if statusUpdate.Status == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Status is required",
		})
	}

	// Validate status value
	validStatuses := map[string]bool{
		"Pending":   true,
		"Shipped":   true,
		"Delivered": true,
		"Cancelled": true,
	}
	if !validStatuses[statusUpdate.Status] {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid status value. Must be one of: Pending, Shipped, Delivered, Cancelled",
		})
	}

	// Update the status
	err = h.orderRepo.UpdateStatus(ctx, id, statusUpdate.Status)
	if err != nil {
		if err.Error() == "order not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Order not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update order status: " + err.Error(),
		})
	}

	// Return updated order
	order, err := h.orderRepo.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Order status updated but failed to retrieve updated order",
		})
	}

	return c.JSON(http.StatusOK, order)
}
