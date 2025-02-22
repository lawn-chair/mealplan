package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/lawn-chair/mealplan/models"

	// Import the root CAs of the system - needed to allow Clerk to work in Docker
	_ "golang.org/x/crypto/x509roots/fallback" // CA bundle for FROM Scratch
)

func getEnv(key string, fallback string) string {
	if value, found := os.LookupEnv(key); found {
		return value
	}
	return fallback
}

func requiresAuthentication(r *http.Request) (*clerk.User, error) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		return nil, errors.New("unauthorized")
	}

	usr, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		return nil, errors.New("user not found")
	}
	fmt.Printf(`{"user_id": "%s", "user_banned": "%t"}\n`, usr.ID, usr.Banned)
	return usr, nil
}

func main() {
	godotenv.Load(".env.local")
	godotenv.Load()

	fmt.Println("Starting mealplan server...")
	db, err := sqlx.Connect("postgres", getEnv("DATABASE_URL", ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("Connected to database")

	clerk.SetKey(getEnv("CLERK_SECRET_KEY", "clerk_secret"))

	r := chi.NewRouter()
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

	r.Route("/api", func(api chi.Router) {
		api.Use(DbCtx(db))

		api.Route("/meals", func(meals chi.Router) {
			meals.Get("/", GetMeals)
			meals.Post("/", CreateMeal)
			meals.Route("/{id}", func(meal chi.Router) {
				meal.Get("/", GetMeal)
				meal.Put("/", UpdateMeal)
				meal.Delete("/", DeleteMeal)
			})
		})

		api.Route("/recipes", func(recipes chi.Router) {
			recipes.Get("/", GetRecipes)
			recipes.Post("/", CreateRecipe)
			recipes.Route("/{id}", func(recipe chi.Router) {
				recipe.Get("/", GetRecipe)
				recipe.Put("/", UpdateRecipe)
				recipe.Delete("/", DeleteRecipe)
			})
		})

		api.Post("/images", func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()

			_, err := requiresAuthentication(r)
			if err != nil {
				http.Error(w, "Unauthorized request", http.StatusUnauthorized)
				return
			}

			file, fileHeader, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()

			storageEndpoint := getEnv("AWS_ENDPOINT_URL_S3", "localhost:9000")
			storageBucket := getEnv("BUCKET_NAME", "mp-images")

			minioClient, err := minio.New(storageEndpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(getEnv("AWS_ACCESS_KEY_ID", ""), getEnv("AWS_SECRET_ACCESS_KEY", ""), ""),
				Secure: false,
				Region: "auto",
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			uuid := uuid.New()

			_, err = minioClient.PutObject(ctx,
				storageBucket,
				uuid.String()+fileHeader.Filename,
				file,
				fileHeader.Size,
				minio.PutObjectOptions{ContentType: "application/octet-stream"})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Println("Uploaded file: ", uuid.String()+fileHeader.Filename)
			json.NewEncoder(w).Encode(map[string]string{"url": "http://" + storageEndpoint + "/" + storageBucket + "/" + uuid.String() + fileHeader.Filename})
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 - Not Found", http.StatusNotFound)
	})

	fmt.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":"+getEnv("PORT", "8080"), r))
}

func DbCtx(db *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "db", db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetMeals(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)

	if slug := r.URL.Query().Get("slug"); slug != "" {
		id, err := models.GetMealIdFromSlug(db, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		meal, err := models.GetMeal(db, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(meal)
	} else {
		meals, err := models.GetMeals(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(meals)
	}
}

func GetMeal(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	meal, err := models.GetMeal(db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(meal)
}

func UpdateMeal(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = requiresAuthentication(r)
	if err != nil {
		http.Error(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Meal)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	meal, err := models.UpdateMeal(db, id, data)
	if err != nil {
		if err == models.ErrValidation {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(meal)
}

func CreateMeal(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	_, err := requiresAuthentication(r)
	if err != nil {
		http.Error(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Meal)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	meal, err := models.CreateMeal(db, data)
	if err != nil {
		if err == models.ErrValidation {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(meal)
}

func DeleteMeal(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = requiresAuthentication(r)
	if err != nil {
		http.Error(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	err = models.DeleteMeal(db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetRecipes(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)

	if slug := r.URL.Query().Get("slug"); slug != "" {
		id, err := models.GetRecipeIdFromSlug(db, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		recipe, err := models.GetRecipe(db, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recipe)
	} else {
		recipes, err := models.GetRecipes(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(recipes)
	}
}

func GetRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipe, err := models.GetRecipe(db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = requiresAuthentication(r)
	if err != nil {
		http.Error(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Recipe)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipe, err := models.UpdateRecipe(db, id, data)
	if err != nil {
		if err == models.ErrValidation {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func CreateRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	_, err := requiresAuthentication(r)
	if err != nil {
		http.Error(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	data := new(models.Recipe)
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipe, err := models.CreateRecipe(db, data)
	if err != nil {
		if err == models.ErrValidation {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(recipe)
}

func DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*sqlx.DB)
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = requiresAuthentication(r)
	if err != nil {
		http.Error(w, "Unauthorized request", http.StatusUnauthorized)
		return
	}

	err = models.DeleteRecipe(db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
