package handlers

import (
	"fmt"
	"gothstack/app/translations"
	"gothstack/app/views/landing"

	"github.com/anthdm/superkit/kit"
)

// HandleLandingIndex handles the landing page.
func HandleLandingIndex(k *kit.Kit) error {
	return k.Render(landing.Index())
}

// HandleTestIndex handles rendering the test component.
func HandleTestIndex(k *kit.Kit) error {
	// Get translations using the helper T and context
	title := translations.T(k.Request.Context(), "landing_title")
	description := translations.T(k.Request.Context(), "landing_description")

	// Get current language code from context (set by middleware)
	currentLangCode := translations.M.DefaultLanguage() // Default if not found
	if langVal := k.Request.Context().Value(translations.ContextKey{}); langVal != nil {
		if code, ok := langVal.(string); ok {
			currentLangCode = code
		}
	}

	// Get available languages from the manager
	availableLangs := translations.M.GetLanguages()

	// Get the current request path
	currentPath := k.Request.URL.Path

	// Render the Test component, passing context, translations, languages, current language, and path
	return k.Render(landing.Test(k.Request.Context(), title, description, availableLangs, currentLangCode, currentPath))
}

// HandleTestAction handles the POST request from the test button.
func HandleTestAction(k *kit.Kit) error {
	fmt.Println("test")
	// Renders the action result component
	return k.Render(landing.TestActionResult(k.Request.Context()))
}
