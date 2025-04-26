package main

import (
	"context"
	"encoding/json"
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
	"github.com/lawn-chair/mealplan/models"
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
			recipes.Get("/", GetRecipes)
			recipes.Post("/", CreateRecipe)
			recipes.Route("/{id}", func(recipe chi.Router) {
				recipe.Use(IdCtx)
				recipe.Get("/", GetRecipe)
				recipe.Put("/", UpdateRecipe)
				recipe.Delete("/", DeleteRecipe)
			})
		})

		apir.Route("/plans", func(plans chi.Router) {
			plans.Get("/", GetPlans)
			plans.Post("/", CreatePlan)
			plans.Route("/{id}", func(plan chi.Router) {
				plan.Use(IdCtx)
				plan.Get("/", GetPlan)
				plan.Put("/", UpdatePlan)
				plan.Delete("/", DeletePlan)
				plan.Get("/ingredients", GetPlanIngredients)
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

func GetRecipes(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)

	if slug := r.URL.Query().Get("slug"); slug != "" {
		id, err := models.GetRecipeIdFromSlug(db, slug)
		if err != nil {
			api.ErrorResponse(w, err.Error(), http.StatusNotFound)
			return
		}
		recipe, err := models.GetRecipe(db, id)
		if err != nil {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recipe)
	} else {
		recipes, err := models.GetRecipes(db)
		if err != nil {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recipes)
	}
}

func GetRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	recipe, err := models.GetRecipe(db, id)
	if err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	_, err := api.RequiresAuthentication(r)
	if err != nil {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Recipe)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipe, err := models.UpdateRecipe(db, id, data)
	if err != nil {
		if err == models.ErrValidation {
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func CreateRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	_, err := api.RequiresAuthentication(r)
	if err != nil {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Recipe)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipe, err := models.CreateRecipe(db, data)
	if err != nil {
		if err == models.ErrValidation {
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = api.RequiresAuthentication(r)
	if err != nil {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	err = models.DeleteRecipe(db, id)
	if err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetPlans(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)

	if r.URL.Query().Get("last") != "" {
		user, err := api.RequiresAuthentication(r)
		if err != nil {
			api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
			return
		}

		plan, err := models.GetLastPlan(db, user.ID)
		if err != nil {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	} else if r.URL.Query().Get("next") != "" {
		user, err := api.RequiresAuthentication(r)
		if err != nil {
			api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
			return
		}

		plan, err := models.GetNextPlan(db, user.ID)
		if err != nil {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	}

	plans, err := models.GetPlans(db)
	if err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plans)
}

func GetPlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	plan, err := models.GetPlan(db, id)
	if err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plan)
}

func UpdatePlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	user, err := api.RequiresAuthentication(r)
	if err != nil {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Plan)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	plan, err := models.GetPlan(db, id)
	if err != nil || (plan.UserID != user.ID) || (data.UserID != "" && data.UserID != user.ID) {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
	}
	data.UserID = user.ID

	plan, err = models.UpdatePlan(db, id, data)
	if err != nil {
		if err == models.ErrValidation {
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(plan)
}

func CreatePlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	user, err := api.RequiresAuthentication(r)
	if err != nil {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Plan)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	data.UserID = user.ID

	plan, err := models.CreatePlan(db, data)
	if err != nil {
		if err == models.ErrValidation {
			api.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		} else {
			api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(plan)
}

func DeletePlan(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	user, err := api.RequiresAuthentication(r)
	if err != nil {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	plan, err := models.GetPlan(db, id)
	if err != nil || (plan.UserID != user.ID) {
		api.ErrorResponse(w, "Unauthorized request", http.StatusUnauthorized)
	}

	err = models.DeletePlan(db, id)
	if err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetPlanIngredients(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id := r.Context().Value("id").(int)

	ingredients, err := models.GetPlanIngredients(db, id)
	if err != nil {
		api.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ingredients)
}
