package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	Client   *mongo.Client
	Database *mongo.Database
}

var dbase MongoConfig

const dbName = "go-hrms"
const MongoURI = "mongodb://127.0.0.1/" + dbName

func ConnectDb() error {
	clientOptions := options.Client().ApplyURI(MongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return fmt.Errorf("error connecting to mongodb")
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		return fmt.Errorf("error pinging mongodb")
	}

	dbase = MongoConfig{
		Client:   client,
		Database: client.Database(dbName),
	}
	fmt.Println("db connected")
	return nil
}

type Employee struct {
	Id     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name"`
	Age    float64 `json:"age"`
	Salary float64 `json:"salary"`
}

func main() {

	if err := ConnectDb(); err != nil {
		log.Fatal(err.Error())
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Hello")
	})

	app.Get("/employees", func(c *fiber.Ctx) error {
		var employee []Employee
		collection := dbase.Database.Collection("employee")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		data, err := collection.Find(ctx, bson.M{})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "error finding employee",
			})
		}

		err = data.All(ctx, &employee)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "error finding employee",
			})
		}

		return c.Status(fiber.StatusOK).JSON(employee)
	})

	app.Get("/employee/:id", func(c *fiber.Ctx) error {
		var employee Employee
		id := c.Params("id")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		collection := dbase.Database.Collection("employee")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid id",
			})
		}
		query := bson.D{{Key: "_id", Value: objID}}
		err = collection.FindOne(ctx, query).Decode(&employee)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "employee not found",
			})
		}
		return c.Status(fiber.StatusOK).JSON(employee)
	})

	app.Post("/employee", func(c *fiber.Ctx) error {
		var employee Employee
		collection := dbase.Database.Collection("employee")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.BodyParser(&employee); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error parsing employee",
			})
		}
		_, err := collection.InsertOne(ctx, employee)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "error creating employee",
			})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "employee created",
			"data":    employee,
		})
	})

	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var updateEmployee Employee
		collection := dbase.Database.Collection("employee")
		if err := c.BodyParser(&updateEmployee); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Error parsing employee",
			})
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		objId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid id",
			})
		}

		var currentEmployee Employee
		err = collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&currentEmployee)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "employee not found",
			})
		}
		if updateEmployee.Name != "" {
			updateEmployee.Name = currentEmployee.Name
		}
		if updateEmployee.Age != 0 {
			updateEmployee.Age = currentEmployee.Age
		}
		if updateEmployee.Salary != 0 {
			updateEmployee.Salary = currentEmployee.Salary
		}
		update := bson.M{
			"$set": bson.M{
				"name":   currentEmployee.Name,
				"age":    currentEmployee.Age,
				"salary": currentEmployee.Salary,
			},
		}

		query := bson.D{{Key: "_id", Value: objId}}
		_, err = collection.UpdateOne(ctx, query, update)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "error updating employee",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "employee updated",
		})
	})

	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		collection := dbase.Database.Collection("employee")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		objId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid id",
			})
		}
		query := bson.D{{Key: "_id", Value: objId}}
		_, err = collection.DeleteOne(ctx, query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "error deleting employee",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "employee deleted",
		})
	})

	app.Listen(":3000")
}
