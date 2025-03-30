package reservations

import (
	"database/sql"
	"fmt"
	"gothstack/plugins/auth"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate" // Using validation library
	"github.com/go-chi/chi/v5"
)

// ServicePageData holds data for the service list page
type ServicePageData struct {
	Services []Service
}

// ServiceFormData holds data for the service create/edit form
type ServiceFormData struct {
	id uint64
	// Service field removed - FormValues will be pre-populated on edit view
	FormValues     ServiceFormValues // Submitted values OR pre-populated values on edit
	FormErrors     v.Errors
	SuccessMessage string
	ErrorMessage   string
}

// ServiceFormValues holds the raw values from the form submission
type ServiceFormValues struct {
	Name         string `form:"name"`
	Description  string `form:"description"`
	Duration     string `form:"duration"` // Parse to int
	Price        string `form:"price"`    // Parse to float64
	Color        string `form:"color"`
	IsActive     string `form:"is_active"` // Checkbox value 'on' or ''
	BufferBefore string `form:"buffer_before"`
	BufferAfter  string `form:"buffer_after"`
	MaxAttendees string `form:"max_attendees"`
	Location     string `form:"location"`
}

// Validation schema for service creation/update
var serviceSchema = v.Schema{
	"name":     v.Rules(v.Min(1), v.Max(100)),
	"duration": v.Rules(v.Min(1)), // Must be > 0 minutes
	// Add more rules as needed (e.g., price format, buffer numbers)
}

// HandleServiceList displays the list of services for the user.
func HandleServiceList(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID

	services, err := ListServices(userID, false) // Get all services (active and inactive)
	if err != nil {
		slog.Error("Failed to list services", "userID", userID, "err", err)
		// Handle error appropriately - maybe render page with an error message
		return fmt.Errorf("could not load services: %w", err)
	}

	data := ServicePageData{
		Services: services,
	}
	return kit.Render(ServiceList(data))
}

// HandleServiceCreateView displays the form for creating a new service.
func HandleServiceCreateView(kit *kit.Kit) error {
	// Pass empty ServiceFormData, defaults are handled in the template/parsing
	data := ServiceFormData{
		FormValues: ServiceFormValues{ // Set default checkbox state if desired
			IsActive:     "on", // Pre-check the active box
			MaxAttendees: "1",  // Default attendee count
		},
	}
	return kit.Render(ServiceCreateForm(data)) // Call the create form
}

// HandleServiceCreatePost processes the submission for creating a new service.
func HandleServiceCreatePost(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	var values ServiceFormValues
	errors, ok := v.Request(kit.Request, &values, serviceSchema)

	// Prepare data for re-rendering form in case of error/success
	formData := ServiceFormData{
		FormValues: values, // Keep submitted values on error
		FormErrors: errors,
	}

	if !ok {
		formData.ErrorMessage = "Please correct the errors below."
		return kit.Render(ServiceCreateForm(formData)) // Render CREATE form on error
	}

	// Parse and convert form values
	service, parseErr := parseServiceFormValues(values, userID, 0) // 0 ID for create
	if parseErr != nil {
		formData.ErrorMessage = fmt.Sprintf("Error processing form data: %v", parseErr)
		return kit.Render(ServiceCreateForm(formData)) // Render CREATE form on error
	}

	// Create the service
	createdService, err := CreateService(service)
	if err != nil {
		slog.Error("Failed to create service", "userID", userID, "err", err)
		formData.ErrorMessage = "Failed to save the service. Please try again."
		return kit.Render(ServiceCreateForm(formData)) // Render CREATE form on error
	}

	// Success - render the CREATE form again with a success message and clear values
	successData := ServiceFormData{
		FormValues: ServiceFormValues{ // Reset form with defaults
			IsActive:     "on",
			MaxAttendees: "1",
		},
		SuccessMessage: fmt.Sprintf("Service '%s' created successfully!", createdService.Name),
	}
	return kit.Render(ServiceCreateForm(successData)) // Render CREATE form on success
}

// HandleServiceEditView displays the form for editing an existing service.
func HandleServiceEditView(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	serviceIDStr := chi.URLParam(kit.Request, "serviceID")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid service ID requested for edit", "id", serviceIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, "/reservations/services")
	}

	service, err := GetService(uint(serviceID), userID)
	if err != nil {
		slog.Error("Failed to get service for edit", "userID", userID, "serviceID", serviceID, "err", err)
		return kit.Redirect(http.StatusSeeOther, "/reservations/services")
	}

	// --- Populate FormValues from the fetched Service ---
	formVals := ServiceFormValues{
		Name:         service.Name,
		Description:  service.Description.String, // Defaults "" if !Valid
		Duration:     strconv.Itoa(service.Duration),
		Color:        service.Color.String,
		BufferBefore: strconv.Itoa(service.BufferBefore),
		BufferAfter:  strconv.Itoa(service.BufferAfter),
		MaxAttendees: strconv.Itoa(service.MaxAttendees),
		Location:     service.Location.String,
	}
	if service.Price.Valid {
		formVals.Price = fmt.Sprintf("%.2f", service.Price.Float64)
	}
	if service.IsActive {
		formVals.IsActive = "on" // Value needed for checkbox checked state
	}
	// --- End population ---

	// Pass the *populated* FormValues to the edit form
	data := ServiceFormData{
		FormValues: formVals,
		id:         serviceID,
		// Service field is removed
		// FormErrors will be empty initially
	}
	return kit.Render(ServiceEditForm(data)) // Call the EDIT form
}

// HandleServiceEditPost processes the submission for updating an existing service.
func HandleServiceEditPost(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	serviceIDStr := chi.URLParam(kit.Request, "serviceID")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid service ID in POST request for edit", "id", serviceIDStr, "err", err)
		return kit.Redirect(http.StatusSeeOther, "/reservations/services")
	}

	// Fetch existing service details only needed for CreatedAt preservation
	// and potentially displaying an error if the record disappeared between GET and POST
	existingService, err := GetService(uint(serviceID), userID)
	if err != nil {
		// Handle case where service was deleted between page load and form submission
		slog.Error("Failed to get service for edit POST consistency check", "userID", userID, "serviceID", serviceID, "err", err)
		// Maybe render a generic error or redirect with flash message
		return kit.Redirect(http.StatusSeeOther, "/reservations/services")
	}

	var values ServiceFormValues
	errors, ok := v.Request(kit.Request, &values, serviceSchema)

	// Prepare data for re-rendering form
	formData := ServiceFormData{
		// No Service field needed here anymore
		FormValues: values, // Show submitted values if error
		FormErrors: errors,
	}

	if !ok {
		formData.ErrorMessage = "Please correct the errors below."
		return kit.Render(ServiceEditForm(formData)) // Render EDIT form on validation error
	}

	// Parse and convert form values, passing the correct ID
	updatedServiceData, parseErr := parseServiceFormValues(values, userID, uint(serviceID))
	if parseErr != nil {
		formData.ErrorMessage = fmt.Sprintf("Error processing form data: %v", parseErr)
		return kit.Render(ServiceEditForm(formData)) // Render EDIT form on parse error
	}

	// Preserve fields not directly editable in this form (like CreatedAt)
	updatedServiceData.CreatedAt = existingService.CreatedAt // Get CreatedAt from the check above

	// Update the service
	_, err = UpdateService(updatedServiceData) // Don't need the returned service here anymore
	if err != nil {
		slog.Error("Failed to update service", "userID", userID, "serviceID", serviceID, "err", err)
		formData.ErrorMessage = "Failed to update the service. Please try again."
		return kit.Render(ServiceEditForm(formData)) // Render EDIT form on save error
	}

	// Success - Fetch the *latest* data to display
	updatedService, err := GetService(uint(serviceID), userID)
	if err != nil {
		// Should be rare, but handle if fetch after update fails
		slog.Error("Failed to get service after successful update", "userID", userID, "serviceID", serviceID, "err", err)
		// Redirect to list with generic success message maybe?
		kit.Redirect(http.StatusSeeOther, "/reservations/services")
	}

	// Repopulate form values from the *actual* updated record
	formVals := ServiceFormValues{
		Name:         updatedService.Name,
		Description:  updatedService.Description.String,
		Duration:     strconv.Itoa(updatedService.Duration),
		Color:        updatedService.Color.String,
		BufferBefore: strconv.Itoa(updatedService.BufferBefore),
		BufferAfter:  strconv.Itoa(updatedService.BufferAfter),
		MaxAttendees: strconv.Itoa(updatedService.MaxAttendees),
		Location:     updatedService.Location.String,
	}
	if updatedService.Price.Valid {
		formVals.Price = fmt.Sprintf("%.2f", updatedService.Price.Float64)
	}
	if updatedService.IsActive {
		formVals.IsActive = "on"
	}

	// Render EDIT form again with success message and updated data
	successFormData := ServiceFormData{
		FormValues:     formVals, // Show the newly saved data
		SuccessMessage: fmt.Sprintf("Service '%s' updated successfully!", updatedService.Name),
		FormErrors:     v.Errors{},
	}
	return kit.Render(ServiceEditForm(successFormData)) // Render EDIT form on success
}

// HandleServiceDelete handles the deletion of a service.
func HandleServiceDelete(kit *kit.Kit) error {
	userID := kit.Auth().(auth.Auth).UserID
	serviceIDStr := chi.URLParam(kit.Request, "serviceID")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid service ID requested for delete", "id", serviceIDStr, "err", err)
		// Respond appropriately for HTMX if used (e.g., no content, error message)
		return kit.Redirect(http.StatusSeeOther, "/reservations/services") // Fallback redirect
	}

	// Optional: Check if service can be deleted (e.g., no future bookings)

	err = DeleteService(uint(serviceID), userID)
	if err != nil {
		slog.Error("Failed to delete service", "userID", userID, "serviceID", serviceID, "err", err)
		// Respond with error for HTMX or set flash message and redirect
		// For HTMX, could return a toast message component
		//kit.Session.Put(kit.Request.Context(), "flash_error", "Failed to delete service. It might be in use.")
		return kit.Redirect(http.StatusSeeOther, "/reservations/services")
	}

	// Success
	// For HTMX: Return HTTP 200 OK (often with no content, as target element is removed)
	// Or set flash message and redirect
	//kit.Session.Put(kit.Request.Context(), "flash_success", "Service deleted successfully.")
	// If using HTMX to remove the row, we might not redirect, just return OK.
	// For now, redirecting to the list:
	return kit.Redirect(http.StatusSeeOther, "/reservations/services")
}

// parseServiceFormValues converts form values to a Service struct.
// Handles nullable fields and basic type conversions.
func parseServiceFormValues(values ServiceFormValues, userID uint, serviceID uint) (Service, error) {
	service := Service{
		ID:     serviceID,
		UserID: userID,
		Name:   values.Name,
	}

	// Nullable fields
	service.Description = sql.NullString{String: values.Description, Valid: values.Description != ""}
	service.Color = sql.NullString{String: values.Color, Valid: values.Color != ""}
	service.Location = sql.NullString{String: values.Location, Valid: values.Location != ""}

	// Integers
	var err error
	service.Duration, err = strconv.Atoi(values.Duration)
	if err != nil {
		return service, fmt.Errorf("invalid duration: %w", err)
	}
	service.BufferBefore, _ = strconv.Atoi(values.BufferBefore) // Ignore error, defaults to 0
	service.BufferAfter, _ = strconv.Atoi(values.BufferAfter)   // Ignore error, defaults to 0
	service.MaxAttendees, err = strconv.Atoi(values.MaxAttendees)
	if err != nil || service.MaxAttendees < 1 { // Ensure at least 1 attendee
		service.MaxAttendees = 1 // Default to 1 if error or invalid
	}

	// Float (Price)
	if values.Price != "" {
		price, err := strconv.ParseFloat(values.Price, 64)
		if err != nil {
			return service, fmt.Errorf("invalid price format: %w", err)
		}
		service.Price = sql.NullFloat64{Float64: price, Valid: true}
	} else {
		service.Price = sql.NullFloat64{Valid: false}
	}

	// Boolean (Checkbox)
	service.IsActive = values.IsActive == "on"

	return service, nil
}
