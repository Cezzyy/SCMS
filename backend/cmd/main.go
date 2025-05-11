package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Cezzyy/SCMS/backend/internal/database"
	"github.com/Cezzyy/SCMS/backend/internal/handlers"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/Cezzyy/SCMS/backend/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

func main() {
	e := echo.New()
	// Initialize database connection
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS configuration - Must specify exact origins when using credentials
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))

	// Security middleware
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Initialize PDF generator service
	// Use absolute paths to avoid inconsistent path resolution
	templatesDir := "C:\\Users\\karl\\Dropbox\\PC\\Desktop\\SCMS\\backend\\cmd\\templates"
	cssDir := "C:\\Users\\karl\\Dropbox\\PC\\Desktop\\SCMS\\backend\\cmd\\templates\\css"

	// Log the actual paths for debugging
	log.Printf("Templates directory (fixed): %s", templatesDir)
	log.Printf("CSS directory (fixed): %s", cssDir)

	// Ensure all template directories exist
	err = services.EnsureTemplateDirectories(templatesDir, "css", "quotation")
	if err != nil {
		log.Printf("Warning: Failed to create template directories: %v", err)
	}

	// Detect wkhtmltopdf location
	wkhtmltopdfPath := "C:\\Program Files\\wkhtmltopdf\\bin\\wkhtmltopdf.exe"
	log.Printf("Using wkhtmltopdf from: %s", wkhtmltopdfPath)

	// Create PDF generator service
	pdfGenerator := services.NewPDFGenerator(templatesDir, cssDir, wkhtmltopdfPath)

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(db)
	contactRepo := repository.NewContactRepository(db)
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	quotationRepo := repository.NewQuotationRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	reportRepo := repository.NewReportRepository(db)

	// Initialize auth service
	authService := services.NewAuthService(userRepo)

	// Initialize handlers
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	contactHandler := handlers.NewContactHandler(contactRepo, customerRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo, productRepo)
	quotationHandler := handlers.NewQuotationHandler(quotationRepo, customerRepo, productRepo, pdfGenerator)
	orderHandler := handlers.NewOrderHandler(orderRepo)
	reportHandler := handlers.NewReportHandler(reportRepo)

	// API Routes
	// Health check
	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Auth routes - Direct Echo handler instead of wrapper
	e.POST("/api/auth/login", func(c echo.Context) error {
		// Parse request body
		var loginReq services.LoginRequest
		if err := c.Bind(&loginReq); err != nil {
			return c.JSON(http.StatusBadRequest, "Invalid request")
		}

		// Validate input
		if loginReq.Email == "" || loginReq.Password == "" {
			return c.JSON(http.StatusBadRequest, "Email and password are required")
		}

		// Attempt to login
		authResponse, err := authService.Login(c.Request().Context(), loginReq)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, err.Error())
		}

		// Set session cookie
		cookie := new(http.Cookie)
		cookie.Name = "session_id"
		cookie.Value = authResponse.SessionID
		cookie.Path = "/"
		cookie.HttpOnly = true
		cookie.Secure = c.Request().TLS != nil
		cookie.SameSite = http.SameSiteLaxMode
		cookie.MaxAge = 86400 // 24 hours in seconds
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, authResponse)
	})

	e.POST("/api/auth/logout", func(c echo.Context) error {
		// Clear the session cookie
		cookie := new(http.Cookie)
		cookie.Name = "session_id"
		cookie.Value = ""
		cookie.Path = "/"
		cookie.HttpOnly = true
		cookie.MaxAge = -1 // Delete the cookie
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
	})

	// Customer routes
	e.GET("/api/customers", customerHandler.GetAllCustomers)
	e.GET("/api/customers/:id", customerHandler.GetCustomerByID)
	e.POST("/api/customers", customerHandler.CreateCustomer)
	e.PUT("/api/customers/:id", customerHandler.UpdateCustomer)
	e.DELETE("/api/customers/:id", customerHandler.DeleteCustomer)
	e.GET("/api/customers/check", customerHandler.CheckCompanyExists)

	// Contact routes - scoped under customer
	e.GET("/api/customers/:customer_id/contacts", contactHandler.GetContactsByCustomer)
	e.GET("/api/customers/:customer_id/contacts/:id", contactHandler.GetContactByID)
	e.POST("/api/customers/:customer_id/contacts", contactHandler.CreateContact)
	e.PUT("/api/customers/:customer_id/contacts/:id", contactHandler.UpdateContact)
	e.DELETE("/api/customers/:customer_id/contacts/:id", contactHandler.DeleteContact)

	// Global contact routes
	e.GET("/api/contacts", contactHandler.GetAllContacts)
	e.GET("/api/contacts/:id", contactHandler.GetContactByID)
	e.GET("/api/contacts/check", contactHandler.CheckEmailExists)

	// Product routes
	e.GET("/api/products", productHandler.GetAllProducts)
	e.GET("/api/products/:id", productHandler.GetProductByID)
	e.POST("/api/products", productHandler.CreateProduct)
	e.PUT("/api/products/:id", productHandler.UpdateProduct)
	e.DELETE("/api/products/:id", productHandler.DeleteProduct)

	// Inventory routes
	e.GET("/api/inventory", inventoryHandler.GetAllInventory)
	e.GET("/api/inventory/:id", inventoryHandler.GetInventoryByID)
	e.GET("/api/inventory/product/:product_id", inventoryHandler.GetInventoryByProductID)
	e.POST("/api/inventory", inventoryHandler.CreateInventory)
	e.PUT("/api/inventory/:id", inventoryHandler.UpdateInventory)
	e.PUT("/api/inventory/:id/stock", inventoryHandler.UpdateStock)
	e.DELETE("/api/inventory/:id", inventoryHandler.DeleteInventory)

	// Low stock routes
	e.GET("/api/inventory/low-stock", inventoryHandler.GetLowStockItems)
	e.GET("/api/inventory/low-stock/details", inventoryHandler.GetLowStockWithProductInfo)

	// Quotation routes
	e.GET("/api/quotations", quotationHandler.GetAllQuotations)
	e.GET("/api/quotations/:id", quotationHandler.GetQuotationByID)
	e.POST("/api/quotations", quotationHandler.CreateQuotation)
	e.GET("/api/quotations/:id/pdf", quotationHandler.GenerateQuotationPDF)
	e.POST("/api/quotations/:id/status", quotationHandler.UpdateQuotationStatus)

	// Order routes
	e.GET("/api/orders", orderHandler.GetAllOrders)
	e.GET("/api/orders/:id", orderHandler.GetOrderByID)
	e.POST("/api/orders", orderHandler.CreateOrder)
	e.PUT("/api/orders/:id", orderHandler.UpdateOrder)
	e.DELETE("/api/orders/:id", orderHandler.DeleteOrder)
	e.POST("/api/orders/:id/status", orderHandler.UpdateOrderStatus)

	// Dashboard & Report routes
	e.GET("/api/dashboard", reportHandler.GetDashboardSummary)
	e.GET("/api/reports/sales-trends", reportHandler.GetSalesTrends)
	e.GET("/api/reports/low-stock", reportHandler.GetLowStockItems)
	e.GET("/api/reports/top-customers", reportHandler.GetTopCustomers)

	// Export CSV routes
	e.GET("/api/reports/sales-trends/export", reportHandler.ExportSalesTrendsCSV)
	e.GET("/api/reports/low-stock/export", reportHandler.ExportLowStockItemsCSV)
	e.GET("/api/reports/top-customers/export", reportHandler.ExportTopCustomersCSV)

	// Start server
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info().
				Str("URI", v.URI).
				Str("method", v.Method).
				Int("status", v.Status).
				Msg("request")
			return nil
		},
	}))
	fmt.Println("Registered routes:")
	for _, route := range e.Routes() {
		fmt.Printf("%-6s %s\n", route.Method, route.Path)
	}
	e.Logger.Fatal(e.Start(":8081"))
}
