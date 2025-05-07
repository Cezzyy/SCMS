package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

// QuotationHandler handles HTTP requests for quotations
type QuotationHandler struct {
	quotationRepo *repository.QuotationRepository
	customerRepo  *repository.CustomerRepository
	productRepo   *repository.ProductRepository
}

// NewQuotationHandler creates a new quotation handler with the provided repositories
func NewQuotationHandler(quotationRepo *repository.QuotationRepository, customerRepo *repository.CustomerRepository, productRepo *repository.ProductRepository) *QuotationHandler {
	return &QuotationHandler{
		quotationRepo: quotationRepo,
		customerRepo:  customerRepo,
		productRepo:   productRepo,
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
	templateData := struct {
		Quotation        models.Quotation
		Customer         models.Customer
		ItemsWithProduct []ItemWithProduct
		GenerationDate   string
	}{
		Quotation:        quotation,
		Customer:         customer,
		ItemsWithProduct: itemsWithProducts,
		GenerationDate:   time.Now().Format("January 2, 2006"),
	}

	// Create a temporary HTML file
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Quotation #{{.Quotation.QuotationID}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            color: #333;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .header h1 {
            color: #2c3e50;
            margin-bottom: 5px;
        }
        .section {
            margin-bottom: 20px;
        }
        .section h2 {
            color: #2c3e50;
            border-bottom: 1px solid #eee;
            padding-bottom: 5px;
        }
        .info-block {
            margin-bottom: 10px;
        }
        .info-label {
            font-weight: bold;
            display: inline-block;
            width: 150px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-top: 10px;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
        tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        .total-row {
            font-weight: bold;
        }
        .footer {
            margin-top: 30px;
            text-align: center;
            font-size: 0.8em;
            color: #777;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Quotation</h1>
        <p>Generated on {{.GenerationDate}}</p>
    </div>

    <div class="section">
        <h2>Quotation Details</h2>
        <div class="info-block">
            <span class="info-label">Quotation #:</span>
            <span>{{.Quotation.QuotationID}}</span>
        </div>
        <div class="info-block">
            <span class="info-label">Date:</span>
            <span>{{.Quotation.QuoteDate.Format "January 2, 2006"}}</span>
        </div>
        <div class="info-block">
            <span class="info-label">Valid until:</span>
            <span>{{.Quotation.ValidityDate.Format "January 2, 2006"}}</span>
        </div>
        <div class="info-block">
            <span class="info-label">Status:</span>
            <span>{{.Quotation.Status}}</span>
        </div>
    </div>

    <div class="section">
        <h2>Customer Information</h2>
        <div class="info-block">
            <span class="info-label">Company:</span>
            <span>{{.Customer.CompanyName}}</span>
        </div>
        {{if .Customer.Address}}
        <div class="info-block">
            <span class="info-label">Address:</span>
            <span>{{.Customer.Address}}</span>
        </div>
        {{end}}
        {{if .Customer.Phone}}
        <div class="info-block">
            <span class="info-label">Phone:</span>
            <span>{{.Customer.Phone}}</span>
        </div>
        {{end}}
        {{if .Customer.Email}}
        <div class="info-block">
            <span class="info-label">Email:</span>
            <span>{{.Customer.Email}}</span>
        </div>
        {{end}}
    </div>

    <div class="section">
        <h2>Items</h2>
        <table>
            <thead>
                <tr>
                    <th>Product</th>
                    <th>Quantity</th>
                    <th>Unit Price</th>
                    <th>Discount</th>
                    <th>Line Total</th>
                </tr>
            </thead>
            <tbody>
                {{range .ItemsWithProduct}}
                <tr>
                    <td>{{.ProductName}}</td>
                    <td>{{.Quantity}}</td>
                    <td>${{printf "%.2f" .UnitPrice}}</td>
                    <td>${{printf "%.2f" .Discount}}</td>
                    <td>${{printf "%.2f" .LineTotal}}</td>
                </tr>
                {{end}}
                <tr class="total-row">
                    <td colspan="4" style="text-align: right;">Total</td>
                    <td>${{printf "%.2f" .Quotation.TotalAmount}}</td>
                </tr>
            </tbody>
        </table>
    </div>

    <div class="footer">
        <p>This is a computer-generated document and does not require a signature.</p>
    </div>
</body>
</html>
`

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "quotation")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create temporary directory",
		})
	}
	defer os.RemoveAll(tempDir)

	// Create the HTML file path
	htmlFilePath := filepath.Join(tempDir, fmt.Sprintf("quotation_%d.html", quotation.QuotationID))

	// Parse and execute the template
	tmpl, err := template.New("quotation").Parse(htmlTemplate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to parse template",
		})
	}

	// Create the HTML file
	htmlFile, err := os.Create(htmlFilePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create HTML file",
		})
	}
	defer htmlFile.Close()

	// Execute the template with data
	err = tmpl.Execute(htmlFile, templateData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to execute template",
		})
	}

	// Create the PDF file path
	pdfFilePath := filepath.Join(tempDir, fmt.Sprintf("quotation_%d.pdf", quotation.QuotationID))

	// Execute wkhtmltopdf command
	cmd := exec.Command("wkhtmltopdf", htmlFilePath, pdfFilePath)
	err = cmd.Run()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate PDF: " + err.Error(),
		})
	}

	// Read the generated PDF
	pdfContent, err := os.ReadFile(pdfFilePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to read generated PDF",
		})
	}

	// Set headers
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=quotation_%d.pdf", quotation.QuotationID))

	// Write the PDF to the response
	return c.Blob(http.StatusOK, "application/pdf", pdfContent)
}
