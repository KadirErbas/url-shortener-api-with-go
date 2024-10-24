package middleware

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gofiber/fiber/v2"
    "url-shortener/database"
	"github.com/go-redis/redis/v8"
    
)

const (
    MaxLimit    = 15                // İzin verilen maksimum istek sayısı
    LimitPeriod = 1 * time.Hour    // Sınır periyodu (örneğin: 1 saat)
)

func RateLimitByIP(c *fiber.Ctx) error {
    ip := c.IP() // Kullanıcının IP adresini al
    redisKey := fmt.Sprintf("ip:%s:shorten_count", ip)

    // Redis'ten IP sayacını al
    count, err := database.RedisClient.Get(database.Ctx, redisKey).Int()
    if err != nil && err != redis.Nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch rate limit data",
        })
    }

    if count >= MaxLimit {
        return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
            "error": "Rate limit exceeded. Please try again later.",
        })
    }

    // Sayaç yoksa başlat, varsa artır
    if count == 0 {
        err = database.RedisClient.Set(database.Ctx, redisKey, 1, LimitPeriod).Err()
    } else {
        err = database.RedisClient.Incr(database.Ctx, redisKey).Err()
    }

    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to update rate limit data",
        })
    }

    return c.Next()
}