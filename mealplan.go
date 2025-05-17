package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/lawn-chair/mealplan/api"
	"github.com/lawn-chair/mealplan/utils"

	// Import the root CAs of the system - needed to allow Clerk to work in Docker
	_ "golang.org/x/crypto/x509roots/fallback" // CA bundle for FROM Scratch
)

func main() {
	godotenv.Load(".env.local")
	godotenv.Load()

	fmt.Println("Starting mealplan server...")
	db, err := sqlx.Connect("postgres", utils.GetEnv("DATABASE_URL", ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("Connected to database")

	clerk.SetKey(utils.GetEnv("CLERK_SECRET_KEY", "clerk_secret"))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(clerkhttp.WithHeaderAuthorization())

	r.Route("/api", func(apir chi.Router) {
		apir.Use(DbCtx(db))

		apir.Route("/pantry", func(pantry chi.Router) {
			pantry.Get("/", api.GetPantryHandler)
			pantry.Put("/", api.UpdatePantryHandler)
			pantry.Delete("/", api.DeletePantryHandler)
			pantry.Post("/", api.CreatePantryHandler)
		})

		apir.Route("/meals", func(meals chi.Router) {
			meals.Get("/", api.GetMealsHandler)
			meals.Post("/", api.CreateMealHandler)
			meals.Route("/{id}", func(meal chi.Router) {
				meal.Use(IdCtx)
				meal.Get("/", api.GetMealHandler)
				meal.Put("/", api.UpdateMealHandler)
				meal.Delete("/", api.DeleteMealHandler)
			})
		})

		apir.Route("/recipes", func(recipes chi.Router) {
			recipes.Get("/", api.GetRecipes)
			recipes.Post("/", api.CreateRecipe)
			recipes.Route("/{id}", func(recipe chi.Router) {
				recipe.Use(IdCtx)
				recipe.Get("/", api.GetRecipe)
				recipe.Put("/", api.UpdateRecipe)
				recipe.Delete("/", api.DeleteRecipe)
			})
		})

		apir.Route("/plans", func(plans chi.Router) {
			plans.Get("/", api.GetPlans)
			plans.Post("/", api.CreatePlan)
			plans.Route("/{id}", func(plan chi.Router) {
				plan.Use(IdCtx)
				plan.Get("/", api.GetPlan)
				plan.Put("/", api.UpdatePlan)
				plan.Delete("/", api.DeletePlan)
				plan.Get("/ingredients", api.GetPlanIngredients)
			})
		})

		apir.Get("/shopping-list", api.GetShoppingList)

		apir.Post("/images", api.PostImageHandler)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		api.ErrorResponse(w, "404 - Not Found", http.StatusNotFound)
	})

	fmt.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":"+utils.GetEnv("PORT", "8080"), r))
}

func DbCtx(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "db", db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func IdCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
