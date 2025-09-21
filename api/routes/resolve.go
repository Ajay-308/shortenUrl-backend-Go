package routes

import (
	"github.com/Ajay-308/shortenUrl/database"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)


func ResolveURL(c *fiber.Ctx) error{
	url := c.Params("url");
	r := database.CreateClient(0);
	// function return ho gya hai to connection ko close kardo
	defer r.Close();
	val, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":"shortened url not found in database",

		})
	}else if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":"cannot connect to database",
		})
	}
	// connect to redis db number 1 which used for storing stats
	r1 := database.CreateClient(1);
	defer r1.Close();
	// increment the visit count for the url
	_ = r1.Incr(database.Ctx, "counter").Err();
	// redirect karde mittar
	return c.Redirect(val,fiber.StatusMovedPermanently);
}