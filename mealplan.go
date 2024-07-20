package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/lawn-chair/mealplan/models"
)

func main() {
	db, err := sqlx.Connect("postgres", "user=admin password=admin dbname=mealplan port=32772 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	app.Use(cors.New())
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

	log.Fatal(app.Listen(":8080"))
}
