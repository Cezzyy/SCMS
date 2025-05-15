package handlers

import (
	"net/http"
	"strconv"

	"github.com/Cezzyy/SCMS/backend/internal/models"
	"github.com/Cezzyy/SCMS/backend/internal/repository"
	"github.com/labstack/echo/v4"
)

// ContactHandler handles HTTP requests for contacts
type ContactHandler struct {
	contactRepo  *repository.ContactRepository
	customerRepo *repository.CustomerRepository
}

// NewContactHandler creates a new contact handler with the provided repositories
func NewContactHandler(contactRepo *repository.ContactRepository, customerRepo *repository.CustomerRepository) *ContactHandler {
	return &ContactHandler{
		contactRepo:  contactRepo,
		customerRepo: customerRepo,
	}
}

// GetAllContacts returns all contacts
func (h *ContactHandler) GetAllContacts(c echo.Context) error {
	ctx := c.Request().Context()

	// Check for search parameter
	searchTerm := c.QueryParam("search")
	var contacts []models.Contact
	var err error

	if searchTerm != "" {
		contacts, err = h.contactRepo.SearchContacts(ctx, searchTerm)
	} else {
		contacts, err = h.contactRepo.GetAll(ctx)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve contacts",
		})
	}

	return c.JSON(http.StatusOK, contacts)
}

// GetContactsByCustomer returns all contacts for a specific customer
func (h *ContactHandler) GetContactsByCustomer(c echo.Context) error {
	ctx := c.Request().Context()

	customerID, err := strconv.Atoi(c.Param("customer_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

	// Verify customer exists
	_, err = h.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Customer not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify customer",
		})
	}

	contacts, err := h.contactRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve contacts",
		})
	}

	return c.JSON(http.StatusOK, contacts)
}

// GetContactByID returns a contact by ID
func (h *ContactHandler) GetContactByID(c echo.Context) error {
	ctx := c.Request().Context()

	// Check if this is a scoped or global request
	customerIDParam := c.Param("customer_id")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid contact ID",
		})
	}

	contact, err := h.contactRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "contact not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Contact not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve contact",
		})
	}

	// If request is scoped to a customer, verify contact belongs to that customer
	if customerIDParam != "" {
		customerID, err := strconv.Atoi(customerIDParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid customer ID",
			})
		}

		if contact.CustomerID != customerID {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Contact not found for this customer",
			})
		}
	}

	return c.JSON(http.StatusOK, contact)
}

// CreateContact creates a new contact
func (h *ContactHandler) CreateContact(c echo.Context) error {
	ctx := c.Request().Context()

	customerID, err := strconv.Atoi(c.Param("customer_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

	// Verify customer exists
	_, err = h.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Customer not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify customer",
		})
	}

	var contact models.Contact
	if err := c.Bind(&contact); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Override customerID with the one from the path parameter
	contact.CustomerID = customerID

	// Validate required fields
	if contact.FirstName == "" || contact.LastName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "First name and last name are required",
		})
	}

	err = h.contactRepo.Create(ctx, &contact)
	if err != nil {
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "A contact with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create contact",
		})
	}

	return c.JSON(http.StatusCreated, contact)
}

// UpdateContact updates an existing contact
func (h *ContactHandler) UpdateContact(c echo.Context) error {
	ctx := c.Request().Context()

	customerID, err := strconv.Atoi(c.Param("customer_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid contact ID",
		})
	}

	// Verify customer exists
	_, err = h.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Customer not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify customer",
		})
	}

	// Verify contact exists and belongs to the customer
	existingContact, err := h.contactRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "contact not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Contact not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve contact",
		})
	}

	if existingContact.CustomerID != customerID {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Contact not found for this customer",
		})
	}

	var contact models.Contact
	if err := c.Bind(&contact); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Ensure ID and CustomerID in path match values in payload
	contact.ContactID = id
	contact.CustomerID = customerID

	// Validate required fields
	if contact.FirstName == "" || contact.LastName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "First name and last name are required",
		})
	}

	err = h.contactRepo.Update(ctx, &contact)
	if err != nil {
		if err.Error() == "contact not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Contact not found",
			})
		}
		if err == repository.ErrDuplicateKey {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "A contact with this information already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update contact",
		})
	}

	return c.JSON(http.StatusOK, contact)
}

// DeleteContact deletes a contact
func (h *ContactHandler) DeleteContact(c echo.Context) error {
	ctx := c.Request().Context()

	customerID, err := strconv.Atoi(c.Param("customer_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid customer ID",
		})
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid contact ID",
		})
	}

	// Verify contact belongs to customer
	contact, err := h.contactRepo.GetByID(ctx, id)
	if err != nil {
		if err.Error() == "contact not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Contact not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify contact",
		})
	}

	if contact.CustomerID != customerID {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Contact not found for this customer",
		})
	}

	err = h.contactRepo.Delete(ctx, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete contact",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// CheckEmailExists checks if an email already exists
func (h *ContactHandler) CheckEmailExists(c echo.Context) error {
	ctx := c.Request().Context()

	email := c.QueryParam("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email is required",
		})
	}

	exists, err := h.contactRepo.CheckEmailExists(ctx, email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to check email existence",
		})
	}

	return c.JSON(http.StatusOK, map[string]bool{
		"exists": exists,
	})
}
