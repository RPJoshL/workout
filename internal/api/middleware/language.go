package middleware

import (
	"context"
	"net/http"
	"strings"

	"git.rpjosh.de/RPJosh/workout/internal/models"
	"git.rpjosh.de/RPJosh/workout/internal/translator"
)

// LanguageMiddleware adds the prefered language of the user to the request as a context key
func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get language based on browser language
		lang := translator.English
		if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
			if strings.HasPrefix(acceptLang, "de") {
				lang = translator.German
			}
		}

		req := r.WithContext(context.WithValue(r.Context(), models.KeyLanguage, lang.String()))
		next.ServeHTTP(w, req)
	})
}
