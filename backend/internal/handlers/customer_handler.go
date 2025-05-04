package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

// CustomerHandler handles HTTP requests for customers
type CustomerHandler struct {
	customerRepo *repository.CustomerRepository
}

// NewCustomerHandler creates a new customer handler with the provided repository
func NewCustomerHandler(customerRepo *repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{
		customerRepo: customerRepo,
	}
}

// Register registers the routes for the customer handler
func (h *CustomerHandler) Register(e *echo.Echo) {
	e.GET("/api/customers", h.GetAllCustomers)
	e.GET("/api/customers/:id", h.GetCustomerByID)
	e.POST("/api/customers", h.CreateCustomer)
	e.PUT("/api/customers/:id", h.UpdateCustomer)
	e.DELETE("/api/customers/:id", h.DeleteCustomer)
}

// GetAllCustomers returns all customers
func (h *CustomerHandler) GetAllCustomers(c echo.Context) error {
	ctx := c.Request().Context()

	// Check for search parameter
	searchTerm := c.QueryParam("search")
	var customers []models.Customer
	var err error

	if searchTerm != "" {
		customers, err = h.customerRepo.SearchCustomers(ctx, searchTerm)
	} else {
		customers, err = h.customerRepo.GetAll(ctx)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve customers",
		})
	}

	return c.JSON(http.StatusOK, customers)
}

// GetCustomerByID returns a customer by ID
func (h *CustomerHandler) GetCustomerByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

	customer, err := h.customerRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Customer not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve customer",
		})
	}

	return c.JSON(http.StatusOK, customer)
}

// CreateCustomer creates a new customer
func (h *CustomerHandler) CreateCustomer(c echo.Context) error {
	ctx := c.Request().Context()

	var customer models.Customer
	if err := c.Bind(&customer); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate required fields
	if customer.CompanyName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Company name is required",
		})
	}

	err := h.customerRepo.Create(ctx, &customer)
	if err != nil {
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "A customer with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create customer",
		})
	}

	return c.JSON(http.StatusCreated, customer)
}

// UpdateCustomer updates an existing customer
func (h *CustomerHandler) UpdateCustomer(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

	var customer models.Customer
	if err := c.Bind(&customer); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Ensure ID in path matches ID in payload
	customer.CustomerID = id

	// Validate required fields
	if customer.CompanyName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Company name is required",
		})
	}

	err = h.customerRepo.Update(ctx, &customer)
	if err != nil {
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Customer not found",
			})
		}
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "A customer with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update customer",
		})
	}

	return c.JSON(http.StatusOK, customer)
}

// DeleteCustomer deletes a customer
func (h *CustomerHandler) DeleteCustomer(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

	err = h.customerRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Customer not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete customer",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
