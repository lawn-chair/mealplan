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
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Cache-Control", "Pragma", "Expires"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(clerkhttp.WithHeaderAuthorization())

	r.Route("/api", func(apir chi.Router) {
		apir.Use(DbCtx(db))

		apir.Route("/pantry", func(pantry chi.Router) {
			pantry.Use(AuthCtx)
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
			plans.Use(AuthCtx)
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

		apir.Route("/shopping-list", func(shoppingList chi.Router) {
			shoppingList.Use(AuthCtx)
			shoppingList.Get("/", api.GetShoppingList)
			shoppingList.Put("/", api.UpdateShoppingList)
		})

		apir.Get("/tags", api.ListTagsHandler)

		apir.Post("/images", api.PostImageHandler)

		apir.Route("/household", func(household chi.Router) {
			household.Use(AuthCtx)
			household.Get("/", api.GetUserHouseholdHandler)
			household.Post("/join-code", api.GenerateJoinCodeHandler)
			household.Post("/join", api.JoinHouseholdHandler)
			household.Post("/leave", api.LeaveHouseholdHandler)
			household.Post("/remove-member", api.RemoveHouseholdMemberHandler)
		})
	})

	// Serve OpenAPI spec as static file
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		http.ServeFile(w, r, "./openapi.yaml")
	})

	// Serve Redoc documentation viewer
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html>
<html>
  <head>
    <title>Mealplan API Docs</title>
    <meta charset=\"utf-8\" />
    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />
    <style>body { margin: 0; padding: 0; }</style>
  </head>
  <body>
    <redoc spec-url='/openapi.yaml'></redoc>
	<script src='https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js'></script>
  </body>
</html>`))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusFound)
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

func AuthCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := api.RequiresAuthentication(r)
		if err != nil {
			api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
			return
		}

		db := r.Context().Value("db").(*sqlx.DB)
		householdID, err := api.GetHouseholdIDForUser(db, user.ID)
		if err != nil {
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		ctx = context.WithValue(ctx, "household", householdID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
