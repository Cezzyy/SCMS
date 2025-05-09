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

// CreateOrder creates a new order with items
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	ctx := c.Request().Context()

	// Define a struct to receive the order data with items
	var orderData struct {
		Order models.Order       `json:"order"`
		Items []models.OrderItem `json:"items"`
	}

	if err := c.Bind(&orderData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
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

	// Create the order with items in a single transaction
	err := h.orderRepo.CreateOrderWithItems(ctx, &orderData.Order, orderData.Items)
	if err != nil {
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "An order with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create order",
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
