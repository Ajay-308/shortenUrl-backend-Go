package helpers

import (
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Ajay-308/shortenUrl/database"
	"github.com/go-redis/redis/v8"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))



func EnforceHttp(url string) string {
	if url[:4] != "http" {
		url = "http://" + url
	}
	return url
}

// createClient creates and returns a Redis client.
// Replace the options below with your actual Redis configuration.



func RemoveDomainError(url string) bool{
	// ye function sare prefixs ko hata dega 
	// just like http://, https:// , www. etx
	// and then check karega ki jo bacha hai wo hamare domain ke barabar hai ki nhi
	//agar hai to false return karega
	if url == os.Getenv("DOMAIN"){
		return false
	}
	newUrl := strings.Replace(url, "http://", "", 1)
	newUrl = strings.Replace(newUrl, "https://", "", 1)
	newUrl = strings.Replace(newUrl, "www.", "", 1)
	return newUrl != os.Getenv("DOMAIN")
}

func CheckCustomShort(customShort string) bool {
	if customShort == "" {
		return false
	}

	db := database.CreateClient(0)
	defer db.Close()

	val, err := db.Get(database.Ctx, customShort).Result()
	if err == redis.Nil {
		return false // key does not exist
	}
	if err != nil {
		// Log error for debugging
		println("Redis error in CheckCustomShort:", err.Error())
		return false
	}
	return val != ""
}

func GenerateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	randStr := make([]rune, n)
	for i := range randStr{
		randStr[i] = letters[seededRand.Intn(len(letters))]
	}
	return string(randStr)
}