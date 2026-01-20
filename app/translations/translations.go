package translations

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed locales/*.json
var localesFS embed.FS

// Define a custom type for the context key to avoid collisions
type ContextKey struct{} // Make it exported by capitalizing

// Language represents a supported language
type Language struct {
	Code  string
	Name  string
	Texts map[string]string
}

// Manager handles language-related operations
type Manager struct {
	defaultLang string
	languages   map[string]Language
}

// M holds the singleton instance of the Manager
var M *Manager

// Init initializes the singleton language Manager.
func Init() {
	// Only initialize if it hasn't been already
	if M == nil {
		M = NewManager()
	}
}

// NewManager creates a new language manager by loading translations from embedded JSON files
func NewManager() *Manager {
	m := &Manager{
		defaultLang: "fi",
		languages:   make(map[string]Language),
	}

	// Read from embedded FS
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		log.Printf("Error reading embedded locales directory: %v. No translations loaded.", err)
		return m
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue // Skip directories and non-JSON files
		}

		langCode := strings.TrimSuffix(entry.Name(), ".json")
		filePath := filepath.Join("locales", entry.Name())

		// Read file from embedded FS
		fileBytes, err := localesFS.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading translation file '%s': %v", filePath, err)
			continue // Skip this language
		}

		var texts map[string]string
		err = json.Unmarshal(fileBytes, &texts)
		if err != nil {
			log.Printf("Error parsing JSON from file '%s': %v", filePath, err)
			continue // Skip this language
		}

		// Extract language name from the JSON data, default to code if not found
		langName, nameExists := texts["_language_name"]
		if !nameExists {
			log.Printf("Warning: '_language_name' key not found in '%s', using code '%s' as name.", filePath, langCode)
			langName = langCode // Use the language code as the name as a fallback
		}
		delete(texts, "_language_name") // Remove the meta key from the actual translations

		m.languages[langCode] = Language{
			Code:  langCode,
			Name:  langName,
			Texts: texts,
		}
		log.Printf("Loaded language: %s (%s)", langName, langCode)
	}

	// Check if the default language was loaded successfully
	if _, exists := m.languages[m.defaultLang]; !exists {
		log.Printf("Warning: Default language '%s' not found.", m.defaultLang)
		// Add a minimal default language entry if it's missing, to prevent panics in GetText
		if len(m.languages) == 0 { // Only if NO languages were loaded at all
			log.Println("Warning: No languages loaded. Adding minimal English fallback.")
			m.languages[m.defaultLang] = Language{
				Code:  m.defaultLang,
				Name:  "English (Default)",
				Texts: map[string]string{"welcome": "Welcome"}, // Add at least one key
			}
		}
	}

	return m
}

// GetText retrieves text for a given key and language
func (m *Manager) GetText(langCode, key string) string {
	if lang, exists := m.languages[langCode]; exists {
		if text, textExists := lang.Texts[key]; textExists {
			return text
		}
		// Optional: Log missing key for the specific language
		// log.Printf("Warning: Translation key '%s' not found for language '%s'", key, langCode)
	}
	// Fallback to default language
	if defaultLang, defaultExists := m.languages[m.defaultLang]; defaultExists {
		if text, textExists := defaultLang.Texts[key]; textExists {
			return text
		}
	}
	// Ultimate fallback: return the key itself if not found anywhere
	log.Printf("Warning: Translation key '%s' not found for language '%s' or default '%s'", key, langCode, m.defaultLang)
	return key
}

// Middleware adds language to the context
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prioritize language selection:
		// 1. URL parameter
		// 2. Cookie
		// 3. Accept-Language header
		// 4. Default language

		// Check URL parameter
		langCode := r.URL.Query().Get("lang")

		// If no URL parameter, check cookie
		if langCode == "" {
			cookie, err := r.Cookie("lang")
			if err == nil {
				langCode = cookie.Value
			}
		}
		// Validate language code
		if _, exists := m.languages[langCode]; !exists {
			langCode = m.defaultLang
		}

		// Create context with language using the custom key type
		ctx := context.WithValue(r.Context(), ContextKey{}, langCode) // Use exported type

		// Set language cookie for persistence
		http.SetCookie(w, &http.Cookie{
			Name:   "lang",
			Value:  langCode,
			Path:   "/",
			MaxAge: 60 * 60 * 24 * 365, // 1 year
		})

		// Call next handler with modified context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// T is a helper function to get translated text based on the context language.
func T(ctx context.Context, key string) string {
	// Default to default language if context/manager is missing
	langCode := M.defaultLang
	if ctx != nil {
		// Retrieve the value using the custom key type
		if langVal := ctx.Value(ContextKey{}); langVal != nil { // Use exported type
			if code, ok := langVal.(string); ok {
				langCode = code
			}
		}
	}

	if M == nil {
		log.Println("Error: Translation Manager (M) is not initialized.")
		return key // Fallback if manager isn't initialized
	}

	return M.GetText(langCode, key)
}

// // LanguageSelector creates a language selection dropdown
// templ LanguageSelector(currentLang string, languages map[string]Language) {
// 	<div class="language-selector">
// 		<select
// 			hx-get="/change-language"
// 			hx-target="body"
// 			hx-swap="unset"
// 			name="lang"
// 		>
// 			for _, lang := range languages {
// 				<option
// 					value={ lang.Code }
// 					selected?={ lang.Code == currentLang }
// 				>
// 					{ lang.Name }
// 				</option>
// 			}
// 		</select>
// 	</div>
// }

// GetLanguages returns the map of available languages.
func (m *Manager) GetLanguages() map[string]Language {
	// Return a copy to prevent external modification? For now, direct return is fine.
	return m.languages
}

// DefaultLanguage returns the configured default language code.
func (m *Manager) DefaultLanguage() string {
	return m.defaultLang
}
