package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Todo struct {
	ID         uint       `json:"id,omitempty" gorm:"primaryKey"`
	Completed  bool       `json:"completed"`
	Body       string     `json:"body"`
	gorm.Model `json:"-"` // Inherit timestamps and other GORM features
}

var db *gorm.DB

func main() {
	fmt.Println("hello world")

	if os.Getenv("ENV") != "production" {
		// Load the .env file if not in production
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	var err error
	POSTGRES_URI := os.Getenv("POSTGRES_URI")
	db, err = gorm.Open(postgres.Open(POSTGRES_URI), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatal(err)
		}
		sqlDB.Close()
	}()

	err = db.AutoMigrate(&Todo{}) // Migrate or create the todos table
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to Postgres")

	app := fiber.New()

	//CORS Middleware
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "http://localhost:5173",
	// 	AllowHeaders: "Origin,Content-Type,Accept",
	// }))

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	// app.Get("/api/todos/:id", getTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))

}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	result := db.Find(&todos)
	if result.Error != nil {
		return result.Error
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	// {id:0,completed:false,body:""}

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	result := db.Create(&todo)
	if result.Error != nil {
		return result.Error
	}

	return c.Status(201).JSON(todo)
}

// func getTodo(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var todo Todo

// 	result := db.First(&todo, id)
// 	if result.Error != nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Invalid todo ID"})
// 	}

// 	return c.JSON(todo)
// }

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	var todo Todo

	result := db.First(&todo, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	todo.Completed = true
	// todo.Completed = !todo.Completed

	result = db.Save(&todo)
	if result.Error != nil {
		return result.Error
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	var todo Todo

	result := db.First(&todo, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	result = db.Delete(&todo)
	if result.Error != nil {
		return result.Error
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}
