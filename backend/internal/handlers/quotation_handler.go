package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

	// Read the raw request body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body: " + err.Error(),
		})
	}

	// Log the raw body for debugging
	fmt.Println("Raw request body:", string(bodyBytes))

	// Restore the body for binding
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Define a struct to hold the request body
	type QuotationRequest struct {
		Quotation models.Quotation       `json:"quotation"`
		Items     []models.QuotationItem `json:"items"`
	}

	var req QuotationRequest
	if err := c.Bind(&req); err != nil {
		fmt.Println("Binding error:", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload: " + err.Error(),
		})
	}

	// Log the bound request for debugging
	fmt.Printf("Bound request: %+v\n", req)
	fmt.Printf("Quotation CustomerID: %d\n", req.Quotation.CustomerID)

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
			// Calculate line total based on the exact same formula as the database
			lineTotal := (float64(item.Quantity) * item.UnitPrice) - item.Discount
			total += lineTotal
		}
		req.Quotation.TotalAmount = total
	}

	// Create the quotation with its items
	err = h.quotationRepo.CreateQuotationWithItems(ctx, &req.Quotation, req.Items)
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
		// CSS will be injected by the PDF generator
	}

	log.Printf("Prepared template data with %d items", len(itemsWithProducts))

	// Generate the PDF using our PDF service
	log.Printf("Generating PDF for quotation ID: %d", id)

	// Use relative paths as expected by the PDF generator
	templateName := "quotation/template.html"
	cssName := "quotation.css"

	log.Printf("Using template: %s", templateName)
	log.Printf("Using CSS: %s", cssName)

	pdfContent, err := h.pdfGenerator.GenerateFromTemplate(
		templateName, // Template path relative to template directory
		cssName,      // CSS file name
		templateData, // Template data
	)

	if err != nil {
		log.Printf("Failed to generate PDF: %v", err)

		// FALLBACK: Return a simple PDF response with basic information
		log.Printf("Attempting fallback PDF generation")

		// Try to create a very basic PDF as a fallback
		fallbackHTML := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Quotation %d</title>
    <style>
        body { 
            font-family: 'Segoe UI', Arial, sans-serif; 
            margin: 30px; 
            line-height: 1.6; 
            color: #333; 
            font-size: 12px;
            background-color: #fff;
        }
        .header { 
            display: flex;
            justify-content: space-between;
            border-bottom: 2px solid #2c5282; 
            padding-bottom: 20px; 
            margin-bottom: 30px; 
        }
        .document-title {
            color: #2c5282;
            font-size: 28px;
            font-weight: bold;
            margin-bottom: 8px;
            letter-spacing: 1px;
        }
        .generation-date {
            color: #666;
            font-size: 12px;
        }
        .company-header {
            text-align: right;
        }
        .company-name { 
            font-size: 18px; 
            font-weight: bold; 
            color: #2c5282;
            letter-spacing: 0.5px;
        }
        .company-info {
            font-size: 12px;
            color: #555;
            line-height: 1.5;
        }
        .quotation-details {
            display: flex;
            margin-bottom: 40px;
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 6px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .quotation-info {
            flex: 1;
        }
        .info-row {
            display: flex;
            margin-bottom: 10px;
        }
        .info-label { 
            font-weight: 600; 
            width: 120px; 
            color: #4a5568;
        }
        .info-value {
            flex: 1;
            color: #2d3748;
        }
        table { 
            width: 100%%; 
            border-collapse: collapse; 
            margin: 30px 0;
            box-shadow: 0 2px 5px rgba(0,0,0,0.05);
        }
        th { 
            background-color: #2c5282; 
            color: white; 
            padding: 12px 15px; 
            text-align: left; 
            font-size: 13px;
            font-weight: 600;
            letter-spacing: 0.5px;
        }
        td { 
            padding: 12px 15px; 
            border-bottom: 1px solid #e2e8f0; 
        }
        tr:nth-child(even) {
            background-color: #f8fafc;
        }
        tr:hover {
            background-color: #f0f4f8;
        }
        .amount-cell { 
            text-align: right; 
            font-family: 'Consolas', 'Courier New', monospace; 
        }
        .total-section {
            display: flex;
            justify-content: flex-end;
            margin: 25px 0;
            padding: 15px;
            background-color: #f1f5f9;
            border-radius: 6px;
        }
        .total-label {
            font-weight: bold;
            padding-right: 30px;
            font-size: 14px;
            color: #2d3748;
        }
        .total-amount {
            font-weight: bold;
            font-family: 'Consolas', 'Courier New', monospace;
            min-width: 150px;
            text-align: right;
            font-size: 16px;
            color: #2c5282;
        }
        .terms-section { 
            margin-top: 40px;
            border: 1px solid #e2e8f0;
            padding: 20px;
            border-radius: 6px;
            background-color: #f8fafc;
        }
        .terms-heading {
            color: #2c5282;
            font-size: 15px;
            font-weight: bold;
            border-bottom: 1px solid #e2e8f0;
            padding-bottom: 10px;
            margin-bottom: 15px;
        }
        .terms-list {
            padding-left: 20px;
        }
        .terms-list li {
            margin-bottom: 8px;
            color: #4a5568;
        }
        .footer { 
            margin-top: 50px; 
            text-align: center; 
            font-size: 11px; 
            color: #666; 
            border-top: 1px solid #e2e8f0; 
            padding-top: 20px; 
        }
        .logo {
            max-width: 150px;
            margin-bottom: 10px;
        }
        .watermark {
            position: fixed;
            top: 50%%;
            left: 50%%;
            transform: translate(-50%%, -50%%) rotate(-45deg);
            font-size: 80px;
            font-weight: bold;
            color: rgba(220, 230, 240, 0.15);
            z-index: -1;
            user-select: none;
        }
        @media print {
            body {
                margin: 0;
                padding: 20px;
            }
            .header, .footer {
                page-break-inside: avoid;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <div>
            <div class="document-title">QUOTATION</div>
            <div class="generation-date">Reference: CISC-Q-%d | Generated on %s</div>
        </div>
        <div class="company-header">
            <div class="company-name">CENTER INDUSTRIAL SUPPLY CORPORATION</div>
            <div class="company-info">
                10 South AA Street, Quezon City<br>
                Metro Manila, Philippines, 1103<br>
                Tel: (02) 8373-9651<br>
                Email: info@centerindustrial.com
            </div>
        </div>
    </div>

    <div class="quotation-details">
        <div class="quotation-info">
            <div class="info-row">
                <div class="info-label">Customer:</div>
                <div class="info-value">%s</div>
            </div>
            <div class="info-row">
                <div class="info-label">Date:</div>
                <div class="info-value">%s</div>
            </div>
            <div class="info-row">
                <div class="info-label">Valid Until:</div>
                <div class="info-value">%s</div>
            </div>
            <div class="info-row">
                <div class="info-label">Status:</div>
                <div class="info-value">%s</div>
            </div>
        </div>
    </div>
    
    <table>
        <thead>
            <tr>
                <th style="width: 40%%;">Product</th>
                <th style="width: 10%%;">Quantity</th>
                <th style="width: 20%%;">Unit Price</th>
                <th style="width: 10%%;">Discount</th>
                <th style="width: 20%%;">Line Total</th>
            </tr>
        </thead>
        <tbody>`,
			quotation.QuotationID,
			quotation.QuotationID,
			time.Now().Format("January 2, 2006"),
			customer.CompanyName,
			quotation.QuoteDate.Format("January 2, 2006"),
			quotation.ValidityDate.Format("January 2, 2006"),
			quotation.Status)

		// Format money values with thousand separators
		formatMoney := func(amount float64) string {
			// Format with two decimal places
			formattedAmount := fmt.Sprintf("%.2f", amount)

			// Split into integer and decimal parts
			parts := strings.Split(formattedAmount, ".")
			integerPart := parts[0]
			decimalPart := parts[1]

			// Add thousand separators to integer part
			for i := len(integerPart) - 3; i > 0; i -= 3 {
				integerPart = integerPart[:i] + "," + integerPart[i:]
			}

			return "₱" + integerPart + "." + decimalPart
		}

		// Add item rows
		for _, item := range itemsWithProducts {
			// Calculate discount percentage if applicable
			discountText := "-"

			// Get discount from the database item record directly
			if item.QuotationItem.Discount > 0 {
				discountPercent := 0.0
				// Calculate discount percentage based on line total before discount
				beforeDiscountTotal := float64(item.QuotationItem.Quantity) * item.QuotationItem.UnitPrice
				if beforeDiscountTotal > 0 {
					discountPercent = (item.QuotationItem.Discount / beforeDiscountTotal) * 100
				}
				discountText = fmt.Sprintf("%.1f%%", discountPercent)
			}

			fallbackHTML += fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td class="amount-cell">%d</td>
            <td class="amount-cell">%s</td>
            <td class="amount-cell">%s</td>
            <td class="amount-cell">%s</td>
        </tr>`,
				item.ProductName,
				item.QuotationItem.Quantity,
				formatMoney(item.QuotationItem.UnitPrice),
				discountText,
				formatMoney(item.QuotationItem.LineTotal))
		}

		// Total amount section
		fallbackHTML += fmt.Sprintf(`
        </tbody>
    </table>
    
    <div class="total-section">
        <div class="total-label">Total Amount:</div>
        <div class="total-amount">%s</div>
    </div>

    <div class="terms-section">
        <div class="terms-heading">Terms and Conditions</div>
        <ol class="terms-list">
            <li>This quotation is valid until the date specified above.</li>
            <li>Prices are in Philippine Peso (₱) and subject to change without notice after the validity period.</li>
            <li>Payment terms: 50%% advance payment upon order confirmation, 50%% prior to delivery.</li>
            <li>Delivery timeframes are subject to stock availability.</li>
            <li>All prices are exclusive of applicable taxes unless otherwise stated.</li>
        </ol>
    </div>

    <div class="footer">
        <p>Thank you for your business!</p>
        <p>Center Industrial Supply Corporation | Your Welding and Cutting Solutions Provider</p>
    </div>
</body>
</html>`, formatMoney(quotation.TotalAmount))

		// Create a temporary file for the fallback HTML
		tempFile, err := os.CreateTemp("", "fallback-*.html")
		if err != nil {
			log.Printf("Failed to create temp file for fallback: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to generate PDF: %v", err),
			})
		}
		tempPath := tempFile.Name()
		defer os.Remove(tempPath) // Clean up

		// Write the fallback HTML
		tempFile.WriteString(fallbackHTML)
		tempFile.Close()

		// Output path for the PDF
		pdfPath := tempPath + ".pdf"
		defer os.Remove(pdfPath) // Clean up

		// Call wkhtmltopdf directly with minimal options
		cmd := exec.Command(
			"C:\\Program Files\\wkhtmltopdf\\bin\\wkhtmltopdf.exe",
			"--quiet",
			tempPath,
			pdfPath,
		)

		cmdOutput, cmdErr := cmd.CombinedOutput()
		if cmdErr != nil {
			log.Printf("Fallback PDF generation failed: %v\nOutput: %s", cmdErr, string(cmdOutput))
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to generate PDF: %v", err),
			})
		}

		// Read the fallback PDF
		pdfContent, err = os.ReadFile(pdfPath)
		if err != nil {
			log.Printf("Failed to read fallback PDF: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to generate PDF: %v", err),
			})
		}

		log.Printf("Fallback PDF generation successful, size: %d bytes", len(pdfContent))
	}
	log.Printf("PDF generation successful, content length: %d bytes", len(pdfContent))

	// Set headers
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=quotation_%d.pdf", quotation.QuotationID))

	// Write the PDF to the response
	return c.Blob(http.StatusOK, "application/pdf", pdfContent)
}

// UpdateQuotationStatus updates the status of an existing quotation
func (h *QuotationHandler) UpdateQuotationStatus(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse the quotation ID from the URL
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid quotation ID",
		})
	}

	// Define a struct to hold the status data
	type StatusUpdate struct {
		Status string `json:"status"`
	}

	// Bind the request body to the struct
	var statusUpdate StatusUpdate
	if err := c.Bind(&statusUpdate); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	// Validate the status
	validStatuses := map[string]bool{
		"Pending":  true,
		"Approved": true,
		"Rejected": true,
		"Expired":  true,
	}

	if !validStatuses[statusUpdate.Status] {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid status. Must be one of: Pending, Approved, Rejected, Expired",
		})
	}

	// Get the quotation to check if it exists
	_, err = h.quotationRepo.GetByID(ctx, id)
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

	// Update the status
	err = h.quotationRepo.UpdateStatus(ctx, id, statusUpdate.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update quotation status: " + err.Error(),
		})
	}

	// Get the updated quotation
	updatedQuotation, err := h.quotationRepo.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Status updated successfully, but failed to retrieve updated quotation",
		})
	}

	return c.JSON(http.StatusOK, updatedQuotation)
}
