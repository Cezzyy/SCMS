package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	productRepo *repository.ProductRepository
}

// NewProductHandler creates a new product handler with the provided repository
func NewProductHandler(productRepo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
	}
}

// GetAllProducts returns all products
func (h *ProductHandler) GetAllProducts(c echo.Context) error {
	ctx := c.Request().Context()

	// Check for search parameter
	searchTerm := c.QueryParam("search")
	var products []models.Product
	var err error

	if searchTerm != "" {
		products, err = h.productRepo.SearchProducts(ctx, searchTerm)
	} else {
		products, err = h.productRepo.GetAll(ctx)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve products",
		})
	}

	return c.JSON(http.StatusOK, products)
}

// GetProductByID returns a product by ID
func (h *ProductHandler) GetProductByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	product, err := h.productRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve product",
		})
	}

	return c.JSON(http.StatusOK, product)
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	var product models.Product
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate required fields
	if product.ProductName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Product name is required",
		})
	}

	err := h.productRepo.Create(ctx, &product)
	if err != nil {
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "A product with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create product",
		})
	}

	return c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates an existing product
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	var product models.Product
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Ensure ID in path matches ID in payload
	product.ProductID = id

	// Validate required fields
	if product.ProductName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Product name is required",
		})
	}

	err = h.productRepo.Update(ctx, &product)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "A product with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update product",
		})
	}

	return c.JSON(http.StatusOK, product)
}

// DeleteProduct deletes a product
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	err = h.productRepo.Delete(ctx, id)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete product",
		})
	}

	return c.NoContent(http.StatusNoContent)
} 