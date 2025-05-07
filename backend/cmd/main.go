package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Cezzyy/SCMS/backend/internal/database"
	"github.com/Cezzyy/SCMS/backend/internal/handlers"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
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

	// Enhanced CORS configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
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

	// JWT secret for user authentication
	// jwtSecret := "your-secret-key-here"

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(db)
	contactRepo := repository.NewContactRepository(db)
	// userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	quotationRepo := repository.NewQuotationRepository(db)

	// Initialize handlers
	customerHandler := handlers.NewCustomerHandler(customerRepo)
	contactHandler := handlers.NewContactHandler(contactRepo, customerRepo)
	// userHandler := handlers.NewUserHandler(userRepo, jwtSecret)
	productHandler := handlers.NewProductHandler(productRepo)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo, productRepo)
	quotationHandler := handlers.NewQuotationHandler(quotationRepo, customerRepo, productRepo)

	// // JWT middleware for protected routes
	// jwtMiddleware := middleware.JWTWithConfig(middleware.JWTConfig{
	// 	SigningKey: []byte(jwtSecret),
	// 	Skipper: func(c echo.Context) bool {
	// 		// Skip authentication for login and health check routes
	// 		if c.Path() == "/api/auth/login" || c.Path() == "/api/health" {
	// 			return true
	// 		}
	// 		return false
	// 	},
	// })

	// API Routes
	// Health check
	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Auth routes - public
	// e.POST("/api/auth/login", userHandler.Login)

	// User routes - protected
	// usersGroup := e.Group("/api/users", jwtMiddleware)
	// usersGroup.GET("", userHandler.GetAllUsers)
	// usersGroup.GET("/:id", userHandler.GetUserByID)
	// usersGroup.POST("", userHandler.CreateUser)
	// usersGroup.PUT("/:id", userHandler.UpdateUser)
	// usersGroup.DELETE("/:id", userHandler.DeleteUser)
	// usersGroup.PUT("/:id/password", userHandler.ChangePassword)

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
