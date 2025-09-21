package routes

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Ajay-308/shortenUrl/database"
	"github.com/Ajay-308/shortenUrl/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

type request struct{
	URL          string  	 	 `json:"url"`
	CustomShort  string   	  	 `json:"customShort"`
	Expiry       time.Duration	 `json:"expiry"`
}

type response struct {
	URL             string
	ShortenedURL 	string    		`json:"shortenedUrl"`
	CustomShort     string          `json:"customShort"`
	Expiry       	time.Duration 	`json:"expiry"`
	XRateRemaining  int             `json:"rateLimitRemaining"`
	XRateLimitRest  int             `json:"rateLimitReset"`
}

func ShortenedURL(c *fiber.Ctx) error{
	body := new(request)
	if err := c.BodyParser(body);err != nil{
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"error":"cannot parse JSON",
		})
	}
	// implement rate limiting here
	// har time jab koi request aayegi to uski ip nikal ke usko check karo ki ye ip
	// kitni baar request kar chuki hai agar wo limit se zayda hai to usko 429 bejh do
	
	r2 := database.CreateClient(2);
	defer r2.Close();
	// increment the count for the ip
	val ,err := r2.Incr(database.Ctx, c.IP()).Result();
	if err == redis.Nil{
		// first time request aa rahi ha is ip se to isko set karde
		_ =r2.Set(database.Ctx,c.IP(),os.Getenv("API_QUOTA"), time.Hour *1).Err();
	}else{
		// agar limit cross kar chuki hai to usko 429 bejh do
		if val <= 0 {
			ttl , _ := r2.TTL(database.Ctx,c.IP()).Result();
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":"rate limit exceeded",
				"rateLimitReset": int(ttl.Seconds()),
			})
		}
	}
	
	// check if the input if an actual url or not
	if !govalidator.IsURL(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"invalid URL",
		})
	}
    // check about domain error 
	if !helpers.RemoveDomainError(body.URL){
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"invalid domain",
		})
	}
	// check if the custom short is already taken or not
	// only check if user provided a custom short
	if body.CustomShort != "" {
		if helpers.CheckCustomShort(body.CustomShort) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Custom short is already taken",
			})
		}
	}

	// enforce http request if not present
	body.URL = helpers.EnforceHttp(body.URL);
	// store the url in the database
	r := database.CreateClient(0);
	defer r.Close();
	// generate a custom short if the user has not provided one

	if body.CustomShort == ""{
		body.CustomShort = helpers.GenerateRandomString(6);
	}else{
		body.CustomShort = strings.TrimSpace(body.CustomShort)
	}
	// set default expriy time of 24 hourse if not provided by user
	if body.Expiry == 0{
		body.Expiry = 24;
	}
	err = r.Set(database.Ctx, body.CustomShort , body.URL , body.Expiry * time.Hour).Err();
	if err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error ":"not able to connect to the database",
		})
	}
	// everything went fine so return the response
	res := response{
		URL: body.URL,
		ShortenedURL: os.Getenv("DOMAIN") + "/" + body.CustomShort,
		Expiry: body.Expiry,
		XRateRemaining: int(val),
		XRateLimitRest: int(time.Now().Add(time.Hour).Unix()),
	}
	r2.Decr(database.Ctx , c.IP());
	valStr, _ := r2.Get(database.Ctx, c.IP()).Result()
	intVal, _ := strconv.ParseInt(valStr, 10, 64)
	res.XRateRemaining = int(intVal)
	
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	res.XRateLimitRest = int(ttl.Seconds())
	res.CustomShort = os.Getenv("DOMAIN") + "/" + body.CustomShort

	return c.Status(fiber.StatusOK).JSON(res)
}


