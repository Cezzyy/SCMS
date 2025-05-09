package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

// ReportHandler handles HTTP requests for dashboard reports
type ReportHandler struct {
	reportRepo *repository.ReportRepository
}

// NewReportHandler creates a new report handler with the provided repository
func NewReportHandler(reportRepo *repository.ReportRepository) *ReportHandler {
	return &ReportHandler{
		reportRepo: reportRepo,
	}
}

// GetDashboardSummary returns all dashboard data in a single request
func (h *ReportHandler) GetDashboardSummary(c echo.Context) error {
	ctx := c.Request().Context()

	// Get days parameter, default to 7 if not provided
	daysStr := c.QueryParam("days")
	days := 7
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid days parameter. Must be a positive integer.",
			})
		}
	}

	// Get dashboard summary
	summary, err := h.reportRepo.GetDashboardSummary(ctx, days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve dashboard data: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, summary)
}

// GetSalesTrends returns sales trend data for the specified period
func (h *ReportHandler) GetSalesTrends(c echo.Context) error {
	ctx := c.Request().Context()

	// Get days parameter, default to 7 if not provided
	daysStr := c.QueryParam("days")
	days := 7
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid days parameter. Must be a positive integer.",
			})
		}
	}

	// Get sales trends
	trends, err := h.reportRepo.GetSalesTrends(ctx, days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve sales trends: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, trends)
}

// GetLowStockItems returns inventory items that are below their reorder level
func (h *ReportHandler) GetLowStockItems(c echo.Context) error {
	ctx := c.Request().Context()

	// Get low stock items
	items, err := h.reportRepo.GetLowStockItems(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve low stock items: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, items)
}

// GetTopCustomers returns top customers by sales amount
func (h *ReportHandler) GetTopCustomers(c echo.Context) error {
	ctx := c.Request().Context()

	// Get limit parameter, default to 5 if not provided
	limitStr := c.QueryParam("limit")
	limit := 5
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid limit parameter. Must be a positive integer.",
			})
		}
	}

	// Get days parameter, default to 365 if not provided (1 year)
	daysStr := c.QueryParam("days")
	days := 365
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid days parameter. Must be a positive integer.",
			})
		}
	}

	// Get top customers
	customers, err := h.reportRepo.GetTopCustomers(ctx, limit, days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve top customers: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, customers)
}

// ExportSalesTrendsCSV exports sales trend data as CSV
func (h *ReportHandler) ExportSalesTrendsCSV(c echo.Context) error {
	ctx := c.Request().Context()

	// Get days parameter, default to 7 if not provided
	daysStr := c.QueryParam("days")
	days := 7
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid days parameter. Must be a positive integer.",
			})
		}
	}

	// Get sales trends
	trends, err := h.reportRepo.GetSalesTrends(ctx, days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve sales trends: " + err.Error(),
		})
	}

	// Set headers for CSV download
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=sales_trends_%d_days.csv", days))

	// Write CSV headers
	csvWriter := csv.NewWriter(c.Response().Writer)
	csvWriter.Write([]string{"Date", "Total Sales"})

	// Write CSV data
	for _, trend := range trends {
		csvWriter.Write([]string{
			trend.Day,
			fmt.Sprintf("%.2f", trend.TotalAmount),
		})
	}

	csvWriter.Flush()
	return nil
}

// ExportLowStockItemsCSV exports low stock items data as CSV
func (h *ReportHandler) ExportLowStockItemsCSV(c echo.Context) error {
	ctx := c.Request().Context()

	// Get low stock items
	items, err := h.reportRepo.GetLowStockItems(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve low stock items: " + err.Error(),
		})
	}

	// Set headers for CSV download
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=low_stock_items.csv")

	// Write CSV headers
	csvWriter := csv.NewWriter(c.Response().Writer)
	csvWriter.Write([]string{"ID", "Product ID", "Product Name", "Current Stock", "Reorder Level", "Unit Price"})

	// Write CSV data
	for _, item := range items {
		csvWriter.Write([]string{
			fmt.Sprintf("%d", item.ID),
			fmt.Sprintf("%d", item.ProductID),
			item.ProductName,
			fmt.Sprintf("%d", item.CurrentStock),
			fmt.Sprintf("%d", item.ReorderLevel),
			fmt.Sprintf("%.2f", item.UnitPrice),
		})
	}

	csvWriter.Flush()
	return nil
}

// ExportTopCustomersCSV exports top customers data as CSV
func (h *ReportHandler) ExportTopCustomersCSV(c echo.Context) error {
	ctx := c.Request().Context()

	// Get limit parameter, default to 20 if not provided (export more than displayed)
	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid limit parameter. Must be a positive integer.",
			})
		}
	}

	// Get days parameter, default to 365 if not provided (1 year)
	daysStr := c.QueryParam("days")
	days := 365
	if daysStr != "" {
		var err error
		days, err = strconv.Atoi(daysStr)
		if err != nil || days <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid days parameter. Must be a positive integer.",
			})
		}
	}

	// Get top customers
	customers, err := h.reportRepo.GetTopCustomers(ctx, limit, days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve top customers: " + err.Error(),
		})
	}

	// Set headers for CSV download
	c.Response().Header().Set(echo.HeaderContentType, "text/csv")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=top_customers_%d_days.csv", days))

	// Write CSV headers
	csvWriter := csv.NewWriter(c.Response().Writer)
	csvWriter.Write([]string{"Customer ID", "Company Name", "Contact Name", "Total Spent", "Order Count"})

	// Write CSV data
	for _, customer := range customers {
		csvWriter.Write([]string{
			fmt.Sprintf("%d", customer.ID),
			customer.Name,
			customer.ContactName,
			fmt.Sprintf("%.2f", customer.TotalSpent),
			fmt.Sprintf("%d", customer.OrderCount),
		})
	}

	csvWriter.Flush()
	return nil
}
