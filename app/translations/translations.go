package translations

import (
	"context"
	"net/http"
	"strings"
)

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

// NewManager creates a new language manager
func NewManager() *Manager {
	return &Manager{
		defaultLang: "en",
		languages: map[string]Language{
			"en": {
				Code: "en",
				Name: "English",
				Texts: map[string]string{
					"welcome":             "Welcome",
					"greeting":            "Hello, world!",
					"about":               "About Us",
					"landing_title":       "Welcome to GothStack",
					"landing_description": "A powerful Go web application framework",
				},
			},
			"es": {
				Code: "es",
				Name: "Español",
				Texts: map[string]string{
					"welcome":             "Bienvenido",
					"greeting":            "¡Hola, mundo!",
					"about":               "Sobre Nosotros",
					"landing_title":       "Bienvenido a GothStack",
					"landing_description": "Un potente framework de aplicaciones web Go",
				},
			},
			"fr": {
				Code: "fr",
				Name: "Français",
				Texts: map[string]string{
					"welcome":             "Bienvenue",
					"greeting":            "Bonjour, monde!",
					"about":               "À Propos",
					"landing_title":       "Bienvenue sur GothStack",
					"landing_description": "Un puissant framework d'applications web Go",
				},
			},
		},
	}
}

// GetText retrieves text for a given key and language
func (m *Manager) GetText(langCode, key string) string {
	if lang, exists := m.languages[langCode]; exists {
		if text, textExists := lang.Texts[key]; textExists {
			return text
		}
	}
	// Fallback to default language
	return m.languages[m.defaultLang].Texts[key]
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

		// If no cookie, check Accept-Language header
		if langCode == "" {
			acceptLang := r.Header.Get("Accept-Language")
			if acceptLang != "" {
				// Take the first language code
				langCodes := strings.Split(acceptLang, ",")
				if len(langCodes) > 0 {
					// Extract language code (e.g., "en" from "en-US")
					parts := strings.Split(langCodes[0], "-")
					langCode = strings.ToLower(parts[0])
				}
			}
		}

		// Validate language code
		if _, exists := m.languages[langCode]; !exists {
			langCode = m.defaultLang
		}

		// Create context with language
		ctx := context.WithValue(r.Context(), "language", langCode)

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
