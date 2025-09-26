package main

import (
	"log"
	"os"

	"github.com/Ajay-308/shortenUrl/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)



func setupRoutes(app *fiber.App){
	app.Get("/:url",routes.ResolveURL);
	app.Post("/api/v1/shorten", routes.ShortenedURL);
}
// testing n8n ( this is extra);

func main(){
	err := godotenv.Load()
	if err != nil{
		panic("error loading env file")
	}
	app := fiber.New();
	app.Use(logger.New());
	setupRoutes(app);
	log.Fatal(app.Listen(os.Getenv("APP_PORT")));

}