package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/Cezzyy/SCMS/backend/internal/services"
	"github.com/labstack/echo/v4"
)

// QuotationHandler handles HTTP requests for quotations
type QuotationHandler struct {
	quotationRepo *repository.QuotationRepository
	customerRepo  *repository.CustomerRepository
	productRepo   *repository.ProductRepository
	pdfGenerator  *services.PDFGenerator
}

// NewQuotationHandler creates a new quotation handler with the provided repositories
func NewQuotationHandler(
	quotationRepo *repository.QuotationRepository,
	customerRepo *repository.CustomerRepository,
	productRepo *repository.ProductRepository,
	pdfGenerator *services.PDFGenerator,
) *QuotationHandler {
	return &QuotationHandler{
		quotationRepo: quotationRepo,
		customerRepo:  customerRepo,
		productRepo:   productRepo,
		pdfGenerator:  pdfGenerator,
	}
}

// GetAllQuotations returns all quotations
func (h *QuotationHandler) GetAllQuotations(c echo.Context) error {
	ctx := c.Request().Context()

	// Check for customer filter
	customerIDStr := c.QueryParam("customer_id")
	var quotations []models.Quotation
	var err error

	if customerIDStr != "" {
		customerID, parseErr := strconv.Atoi(customerIDStr)
		if parseErr != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid customer ID",
			})
		}
		quotations, err = h.quotationRepo.GetByCustomerID(ctx, customerID)
	} else {
		quotations, err = h.quotationRepo.GetAll(ctx)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve quotations",
		})
	}

	return c.JSON(http.StatusOK, quotations)
}

// GetQuotationByID returns a quotation by ID
func (h *QuotationHandler) GetQuotationByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid quotation ID",
		})
	}

	// Get the quotation with its items
	quotation, items, err := h.quotationRepo.GetFullQuotation(ctx, id)
	if err != nil {
		if err.Error() == "quotation not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Quotation not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve quotation",
		})
	}

	// Return both the quotation and its items
	return c.JSON(http.StatusOK, map[string]interface{}{
		"quotation": quotation,
		"items":     items,
	})
}

// CreateQuotation creates a new quotation with items
func (h *QuotationHandler) CreateQuotation(c echo.Context) error {
	ctx := c.Request().Context()

	// Define a struct to hold the request body
	type QuotationRequest struct {
		Quotation models.Quotation       `json:"quotation"`
		Items     []models.QuotationItem `json:"items"`
	}

	var req QuotationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate required fields
	if req.Quotation.CustomerID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Customer ID is required",
		})
	}

	if req.Quotation.QuoteDate.IsZero() {
		req.Quotation.QuoteDate = time.Now()
	}

	if req.Quotation.ValidityDate.IsZero() {
		// Default validity: 30 days from quote date
		req.Quotation.ValidityDate = req.Quotation.QuoteDate.AddDate(0, 0, 30)
	}

	if req.Quotation.Status == "" {
		req.Quotation.Status = "PENDING"
	}

	// Calculate total if not provided
	if req.Quotation.TotalAmount == 0 && len(req.Items) > 0 {
		var total float64
		for _, item := range req.Items {
			total += item.LineTotal
		}
		req.Quotation.TotalAmount = total
	}

	// Create the quotation with its items
	err := h.quotationRepo.CreateQuotationWithItems(ctx, &req.Quotation, req.Items)
	if err != nil {
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "A quotation with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create quotation: " + err.Error(),
		})
	}

	// Get the newly created quotation with its items
	quotation, items, err := h.quotationRepo.GetFullQuotation(ctx, req.Quotation.QuotationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Quotation created but failed to retrieve it",
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"quotation": quotation,
		"items":     items,
	})
}

// GenerateQuotationPDF generates a PDF for a quotation using wkhtmltopdf
func (h *QuotationHandler) GenerateQuotationPDF(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid quotation ID",
		})
	}

	// Get the quotation with its items
	quotation, items, err := h.quotationRepo.GetFullQuotation(ctx, id)
	if err != nil {
		if err.Error() == "quotation not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Quotation not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve quotation",
		})
	}

	// Get customer information
	customer, err := h.customerRepo.GetByID(ctx, quotation.CustomerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve customer information",
		})
	}

	// Get product details for each item
	type ItemWithProduct struct {
		models.QuotationItem
		ProductName string `json:"product_name"`
	}

	itemsWithProducts := make([]ItemWithProduct, len(items))
	for i, item := range items {
		product, err := h.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to retrieve product information",
			})
		}

		itemsWithProducts[i] = ItemWithProduct{
			QuotationItem: item,
			ProductName:   product.ProductName,
		}
	}

	// Create a data structure for the template
	templateData := map[string]interface{}{
		"Quotation":        quotation,
		"Customer":         customer,
		"ItemsWithProduct": itemsWithProducts,
		"GenerationDate":   time.Now().Format("January 2, 2006"),
	}

	// Generate the PDF using our PDF service
	pdfContent, err := h.pdfGenerator.GenerateFromTemplate(
		"quotation/template.html", // Template path
		"quotation.css",           // CSS file name
		templateData,              // Template data
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to generate PDF: %v", err),
		})
	}

	// Set headers
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=quotation_%d.pdf", quotation.QuotationID))

	// Write the PDF to the response
	return c.Blob(http.StatusOK, "application/pdf", pdfContent)
}
