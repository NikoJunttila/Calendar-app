package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gothstack/app"
	"gothstack/plugins/reservations"
	"gothstack/public"

	"github.com/anthdm/superkit/kit"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	reservations.SetupMailer()
	kit.Setup()
	router := chi.NewMux()
	app.InitializeMiddleware(router)

	isDev := kit.IsDevelopment() || os.Getenv("SUPERKIT_ENV") == "dev"
	isProd := kit.IsProduction() || os.Getenv("SUPERKIT_ENV") == "prod"

	if isDev {
		router.Handle("/public/*", disableCache(http.StripPrefix("/public/", http.FileServer(http.Dir("public")))))
	} else if isProd {
		router.Handle("/public/*", http.StripPrefix("/public/", http.FileServerFS(public.AssetsFS)))
	}

	app.InitializeRoutes(router)
	app.RegisterEvents()

	kit.UseErrorHandler(app.ErrorHandler)
	router.NotFound(kit.Handler(app.NotFoundHandler))

	listenAddr := os.Getenv("HTTP_LISTEN_ADDR")
	// In development link the full Templ proxy url.
	url := "http://localhost:7331"
	if kit.IsProduction() {
		url = fmt.Sprintf("http://localhost%s", listenAddr)
	}

	fmt.Printf("application running in %s at %s\n", kit.Env(), url)

	http.ListenAndServe(listenAddr, router)
}

func disableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
	}
	log.Println("SUPERKIT_ENV:", os.Getenv("SUPERKIT_ENV"))
	log.Println("DB_NAME:", os.Getenv("DB_NAME"))
	fmt.Println("loaded env")
}
