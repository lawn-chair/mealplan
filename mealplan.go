package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/clerk/clerk-sdk-go/v2/user"

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
		fmt.Println(c.GetReqHeaders())
		claims, ok := clerk.SessionClaimsFromContext(c.Context())
		if !ok {
			fmt.Println("Unauthorized")
			c.SendStatus(403)
			return c.JSON([]byte(`{"access": "unauthorized"}`))
		}

		usr, err := user.Get(c.Context(), claims.Subject)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf(`{"user_id": "%s", "user_banned": "%t"}\n`, usr.ID, usr.Banned)

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

		err = models.DeleteRecipe(db, id)
		if err != nil {
			c.SendStatus(500)
			return c.JSON(err)
		}

		return c.SendStatus(204)
	})

	fmt.Println("Server starting on port 8080")
	log.Fatal(app.Listen(":" + getEnv("PORT", "8080")))
}
