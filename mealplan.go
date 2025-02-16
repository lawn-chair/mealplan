package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/lawn-chair/mealplan/models"
)

func getEnv(key string, fallback string) string {
	if value, found := os.LookupEnv(key); found {
		return value
	}
	return fallback
}

func requiresAuthentication(c *fiber.Ctx) (*clerk.User, error) {
	claims, ok := clerk.SessionClaimsFromContext(c.Context())
	if !ok {
		return nil, errors.New("unauthorized")
	}

	usr, err := user.Get(c.Context(), claims.Subject)
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
	db, err := sqlx.Connect("postgres", "user=admin password=admin dbname=mealplan port=32772 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("Connected to database")

	clerk.SetKey(getEnv("CLERK_SECRET_KEY", "clerk_secret"))

	app := fiber.New()
	app.Use(cors.New())
	app.Use(adaptor.HTTPMiddleware(clerkhttp.WithHeaderAuthorization()))

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	api := app.Group("/api")

	api.Get("/meals", func(c *fiber.Ctx) error {
		if c.Query("slug") != "" {
			id, err := models.GetMealIdFromSlug(db, c.Query("slug"))
			if err != nil {
				c.SendStatus(404)
				return c.JSON(err)
			}
			meal, err := models.GetMeal(db, id)
			if err != nil {
				c.SendStatus(500)
				return c.JSON(err)
			}
			return c.JSON(meal)
		} else {
			meals, err := models.GetMeals(db)
			if err != nil {
				c.SendStatus(500)
				return c.JSON(err)
			}
			return c.JSON(meals)
		}
	})

	api.Post("/meals", func(c *fiber.Ctx) error {
		_, err = requiresAuthentication(c)
		if err != nil {
			c.SendStatus(401)
			return c.JSON(err)
		}

		data := new(models.Meal)
		if err := c.BodyParser(data); err != nil {
			c.SendStatus(400)
			fmt.Println(err)
			return c.JSON(err)
		}

		recipe, err := models.CreateMeal(db, data)
		if err != nil {
			if err == models.ErrValidation {
				c.Status(400)
				return c.JSON(fiber.Map{"message": err.Error()})
			} else {
				c.Status(500)
				return c.JSON(err)
			}
		}
		return c.JSON(recipe)
	})

	api.Put("/meals/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			c.SendStatus(400)
			return c.JSON(err)
		}

		_, err = requiresAuthentication(c)
		if err != nil {
			c.SendStatus(401)
			return c.JSON(err)
		}

		data := new(models.Meal)
		if err := c.BodyParser(data); err != nil {
			c.SendStatus(400)
			fmt.Println(err)
			return c.JSON(err)
		}

		meal, err := models.UpdateMeal(db, id, data)
		if err != nil {
			if err == models.ErrValidation {
				c.Status(400)
				return c.JSON(fiber.Map{"message": err.Error()})
			} else {
				c.Status(500)
				return c.JSON(err)
			}
		}

		return c.JSON(meal)
	})

	api.Delete("/meals/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			c.SendStatus(400)
			return c.JSON(err)
		}

		_, err = requiresAuthentication(c)
		if err != nil {
			c.SendStatus(401)
			return c.JSON(err)
		}

		err = models.DeleteMeal(db, id)
		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}

		return c.SendStatus(204)
	})

	api.Get("/recipes", func(c *fiber.Ctx) error {
		recipes, err := models.GetRecipes(db)
		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}
		if c.Query("slug") != "" {
			fmt.Println(c.Query("slug"))
			id, err := models.GetRecipeIdFromSlug(db, c.Query("slug"))
			if err != nil {
				c.SendStatus(404)
				return c.JSON(err)
			}
			recipe, err := models.GetRecipe(db, id)
			if err != nil {
				c.SendStatus(500)
				return c.JSON(err)
			}
			fmt.Println(recipe)
			return c.JSON(recipe)
		}

		return c.JSON(recipes)
	})

	api.Get("/recipes/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			c.SendStatus(400)
			return c.JSON(err)
		}
		recipe, err := models.GetRecipe(db, id)
		fmt.Println(recipe)
		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}

		return c.JSON(recipe)
	})

	api.Post("/recipes", func(c *fiber.Ctx) error {
		_, err = requiresAuthentication(c)
		if err != nil {
			c.SendStatus(401)
			return c.JSON(err)
		}

		data := new(models.Recipe)
		if err := c.BodyParser(data); err != nil {
			c.SendStatus(400)
			fmt.Println(err)
			return c.JSON(err)
		}

		recipe, err := models.CreateRecipe(db, data)
		if err != nil {
			if err == models.ErrValidation {
				c.Status(400)
				return c.JSON(fiber.Map{"message": err.Error()})
			} else {
				c.Status(500)
				return c.JSON(err)
			}
		}
		return c.JSON(recipe)
	})

	api.Put("/recipes/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			c.SendStatus(400)
			return c.JSON(err)
		}

		_, err = requiresAuthentication(c)
		if err != nil {
			c.SendStatus(401)
			return c.JSON(err)
		}

		data := new(models.Recipe)
		if err := c.BodyParser(data); err != nil {
			c.SendStatus(400)
			fmt.Println(err)
			return c.JSON(err)
		}

		recipe, err := models.UpdateRecipe(db, id, data)
		if err != nil {
			if err == models.ErrValidation {
				c.Status(400)
				return c.JSON(fiber.Map{"message": err.Error()})
			} else {
				c.Status(500)
				return c.JSON(err)
			}
		}

		return c.JSON(recipe)
	})

	api.Delete("/recipes/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			c.SendStatus(400)
			return c.JSON(err)
		}

		_, err = requiresAuthentication(c)
		if err != nil {
			c.SendStatus(401)
			return c.JSON(err)
		}

		err = models.DeleteRecipe(db, id)
		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}

		return c.SendStatus(204)
	})

	api.Post("/images", func(c *fiber.Ctx) error {
		fmt.Println("Uploading image")
		ctx := context.Background()

		_, err = requiresAuthentication(c)
		if err != nil {
			c.SendStatus(401)
			return c.JSON(err)
		}

		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.SendStatus(400)
			return c.JSON(err)
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}
		defer file.Close()
		storageEndpoint := getEnv("S3_ENDPOINT", "localhost:9000")
		storageBucket := getEnv("S3_BUCKET", "mp-images")

		minioClient, err := minio.New(storageEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(getEnv("S3_ACCESS_KEY", ""), getEnv("S3_SECRET_KEY", ""), ""),
			Secure: false,
			Region: "auto",
		})
		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}
		uuid := uuid.New()

		_, err = minioClient.PutObject(ctx,
			storageBucket,
			uuid.String()+fileHeader.Filename,
			file,
			fileHeader.Size,
			minio.PutObjectOptions{ContentType: "application/octet-stream"})

		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}
		fmt.Println("Uploaded file: ", uuid.String()+fileHeader.Filename)
		return c.JSON(fiber.Map{"url": "http://" + storageEndpoint + "/" + storageBucket + "/" + uuid.String() + fileHeader.Filename})
	})

	app.Use(func(c *fiber.Ctx) error {
		fmt.Println("404 - ", c.Method(), c.OriginalURL())
		return c.SendStatus(404) // => 404 "Not Found"
	})

	fmt.Println("Server starting on port 8080")
	log.Fatal(app.Listen(":" + getEnv("PORT", "8080")))
}
