package calendar

import (
	"fmt"
	"gothstack/plugins/auth"
	"net/http"
	"strconv"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
)

// Validation schema for work resource creation and update
var workResourceSchema = v.Schema{
	"name": v.Rules(v.Min(1)), // ensure a non-empty name
	// No validation for resources_percentage in schema - we'll handle it manually
}

// Custom validation for resources percentage
func validateResourcesPercentage(values WorkResourceFormValues, errors v.Errors) bool {
	if values.ResourcesPercentage < 0 || values.ResourcesPercentage > 100 {
		errors.Add("resources_percentage", "Resources percentage must be between 0 and 100")
		return false
	}
	return true
}

// WorkResourcePageData holds data for the work resource pages
type WorkResourcePageData struct {
	WorkResources []WorkResource
	Calendar      Calendar
	FormValues    WorkResourceFormValues
	FormErrors    v.Errors
}

// WorkResourceFormValues holds form data for creating/updating a work resource
type WorkResourceFormValues struct {
	Name                string `form:"name"`
	ResourcesPercentage int    `form:"resources_percentage"`
	SuccessMessage      string
}

// HandleWorkResourceList renders the work resources list page
func HandleWorkResourceList(kit *kit.Kit) error {
	// Get the calendar ID from the URL parameter
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}

	// Retrieve the calendar details
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	calendar, err := GetCalendar(uint(calendarID), userID)
	if err != nil {
		return err
	}

	// Retrieve the work resources for this calendar
	resources, err := ListWorkResourcesByCalendar(uint(calendarID))
	if err != nil {
		return err
	}
	total := 0
	for _, resource := range resources {
		total += resource.ResourcesPercentage
	}
	// Render the work resources list page
	data := WorkResourcePageData{
		WorkResources: resources,
		Calendar:      calendar,
	}
	return kit.Render(WorkResourceList(data, total))
}

// HandleWorkResourceCreate renders the work resource creation form (GET request)
func HandleWorkResourceCreate(kit *kit.Kit) error {
	// Get the calendar ID from the URL parameter
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}

	// Retrieve the calendar details
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	calendar, err := GetCalendar(uint(calendarID), userID)
	if err != nil {
		return err
	}

	// Render the work resource creation form
	data := WorkResourcePageData{
		Calendar:   calendar,
		FormValues: WorkResourceFormValues{},
	}
	return kit.Render(WorkResourceCreate(data))
}

// HandleWorkResourceCreatePost processes the form submission (POST request) for creating a work resource
func HandleWorkResourceCreatePost(kit *kit.Kit) error {
	// Get the calendar ID from the URL parameter
	calendarIDStr := chi.URLParam(kit.Request, "id")
	calendarID, err := strconv.ParseUint(calendarIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid calendar ID: %w", err)
	}

	// Parse and validate the form values
	var values WorkResourceFormValues
	errors, ok := v.Request(kit.Request, &values, workResourceSchema)
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	// Retrieve the calendar details for re-rendering the form if needed
	calendar, err := GetCalendar(uint(calendarID), userID)
	if err != nil {
		return err
	}

	// Perform additional validation for resources percentage
	if !validateResourcesPercentage(values, errors) {
		ok = false
	}

	if !ok {
		return kit.Render(WorkResourceForm(values, errors, calendar))
	}
	// Create the new work resource
	resource, err := CreateWorkResource(values.Name, userID, uint(calendarID), values.ResourcesPercentage)
	if err != nil {
		errors.Add("general", "Failed to create work resource")
		return kit.Render(WorkResourceForm(values, errors, calendar))
	}

	// Set a success message and re-render the form
	success := fmt.Sprintf("New work resource created: %s with ID %d", values.Name, resource.ID)
	return kit.Render(WorkResourceForm(WorkResourceFormValues{SuccessMessage: success}, errors, calendar))
}

// HandleWorkResourceEdit renders the work resource edit form (GET request)
func HandleWorkResourceEdit(kit *kit.Kit) error {
	// Get the work resource ID from the URL parameter
	resourceIDStr := chi.URLParam(kit.Request, "resource_id")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid resource ID: %w", err)
	}

	// Retrieve the work resource details
	resource, err := GetWorkResource(uint(resourceID))
	if err != nil {
		return err
	}
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	// Retrieve the calendar details
	calendar, err := GetCalendar(resource.CalendarID, userID)
	if err != nil {
		return err
	}

	// Populate form values from the existing resource
	values := WorkResourceFormValues{
		Name:                resource.Name,
		ResourcesPercentage: resource.ResourcesPercentage,
	}

	// Render the work resource edit form
	data := WorkResourcePageData{
		Calendar:   calendar,
		FormValues: values,
	}
	return kit.Render(WorkResourceEdit(data, uint(resourceID)))
}

// HandleWorkResourceEditPost processes the form submission (POST request) for updating a work resource
func HandleWorkResourceEditPost(kit *kit.Kit) error {
	// Get the work resource ID from the URL parameter
	resourceIDStr := chi.URLParam(kit.Request, "resource_id")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid resource ID: %w", err)
	}

	// Retrieve the work resource details
	resource, err := GetWorkResource(uint(resourceID))
	if err != nil {
		return err
	}
	auth := kit.Auth().(auth.Auth)
	userID := auth.UserID
	// Retrieve the calendar details
	calendar, err := GetCalendar(resource.CalendarID, userID)
	if err != nil {
		return err
	}

	// Parse and validate the form values
	var values WorkResourceFormValues
	errors, ok := v.Request(kit.Request, &values, workResourceSchema)

	// Perform additional validation for resources percentage
	if !validateResourcesPercentage(values, errors) {
		ok = false
	}

	if !ok {
		return kit.Render(WorkResourceEditForm(values, errors, calendar, uint(resourceID)))
	}

	// Update the work resource
	updatedResource, err := UpdateWorkResource(uint(resourceID), values.Name, values.ResourcesPercentage)
	if err != nil {
		errors.Add("general", "Failed to update work resource")
		return kit.Render(WorkResourceEditForm(values, errors, calendar, uint(resourceID)))
	}

	// Set a success message
	values.SuccessMessage = fmt.Sprintf("Work resource updated: %s", updatedResource.Name)
	return kit.Render(WorkResourceEditForm(values, errors, calendar, uint(resourceID)))
}

// HandleWorkResourceDelete processes the request to delete a work resource
func HandleWorkResourceDelete(kit *kit.Kit) error {
	// Get the work resource ID from the URL parameter
	resourceIDStr := chi.URLParam(kit.Request, "resource_id")
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid resource ID: %w", err)
	}

	// Retrieve the work resource details to get the calendar ID
	resource, err := GetWorkResource(uint(resourceID))
	if err != nil {
		return err
	}

	calendarID := resource.CalendarID

	// Delete the work resource
	err = DeleteWorkResource(uint(resourceID))
	if err != nil {
		return err
	}

	// Redirect to the work resources list page

	return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/calendars/%d/resources", calendarID))
}

// return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/calendars/%d/resources", calendarID))
/* auth := kit.Auth().(auth.Auth)
userID := auth.UserID
if userID == 0 {
	errors.Add("general", "User not authenticated")
	return kit.Render(WorkResourceForm(values, errors, calendar))
} */
