package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"log"
	"url-shortener/database"
	"url-shortener/models"

	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v8"
)
/*
Kısaltılmış URL'ler belirlenen bir zamanda geçerli olmalıdır. Bu süre sonunda kısaltılmış URL'lerin
kullanılabilirliği sona ermelidir. (redis kullanılabilir, ttl kavramı incelenmeli) Bu değer environment içerisinden değiştirilebilmelidir.
2. Süre bittikten sonra URL kullanılamasa bile istatistiklere ulaşabilmek için saklanmalıdır.
*/

// 1- delete yaptığında deleted_at kısmını doldur
// 2-  ekleme işleminde postgrede kayı  zaten varsa güncelleme yapmamız lazım
// ttl süresi dolduğunda redisten silinen verinin ilgili postgre şeysi tetikleniyor mu kontrol et
// *time.time ye bak belki de bunu yapmalısın

// önceden üretilmiş short url'leri tutmak için
var generatedURLs = make(map[string]bool)

// Rastgele alfanumerik bir kısaltılmış URL oluşturur.
func generateShortURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// Yeni bir rastgele sayı üreteci oluştur
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		shortURL := make([]byte, 6)
		for i := range shortURL {
			shortURL[i] = charset[r.Intn(len(charset))]
		}

		// oluşturulan url önceden üretilmiş mi diye kontrol et
		url := string(shortURL)
		if _, exists := generatedURLs[url]; !exists{
			generatedURLs[url] = true
			return url
		} 
	}
}
const (
	ActiveTime = 5 * time.Minute // 5 yazan kısım .env den çekliecek
)
func ShortenURL(c *fiber.Ctx) error {
    // İstek gövdesini ayrıştır
    type Request struct {
        OriginalURL string `json:"original_url" binding:"required"`
    }
    var body Request
    if err := c.BodyParser(&body); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Cannot parse JSON",
        })
    }

    // Benzersiz kısaltılmış URL üret
    shortURL := generateShortURL()

    // Veritabanında aynı original_url var mı kontrol et
    var existingURL models.URL
    result := database.DB.Where("original_url = ?", body.OriginalURL).First(&existingURL)

    if result.Error == nil {
        // Redis'te shortURL kayıtlı mı kontrol et
        redisValue, err := database.RedisClient.Get(database.Ctx, existingURL.ShortURL).Result()

        if err == nil && redisValue == body.OriginalURL {
            // Hem veritabanında hem de Redis'te mevcut
            return c.Status(fiber.StatusConflict).JSON(fiber.Map{
                "message":     "This URL has already been shortened",
                "short_url":   existingURL.ShortURL,
                "created_at":  existingURL.CreatedAt,
                "expires_at":  existingURL.ExpiresAt,
                "usage_count": existingURL.UsageCount,
            })
        }
    }

    // Yeni URL oluşturma veya güncelleme işlemi
    url := models.URL{
        OriginalURL: body.OriginalURL,
        ShortURL:    shortURL,
        CreatedAt:   time.Now(),
        ExpiresAt:   time.Now().Add(ActiveTime),
        UsageCount:  0,
    }

    // Veritabanına kaydet
    if err := database.DB.Save(&url).Error; err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to save URL",
        })
    }

    // Redis'e TTL ile kaydet
    err := database.RedisClient.Set(database.Ctx, shortURL, body.OriginalURL, ActiveTime).Err()
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to save to Redis",
        })
    }

    // JSON yanıtı döndür
    response := fiber.Map{
        "created_at":   url.CreatedAt,
        "original_url": url.OriginalURL,
        "short_url":    "http://localhost:3000/" + url.ShortURL,
        "expires_at":   url.ExpiresAt,
        "usage_count":  url.UsageCount,
    }

    return c.Status(fiber.StatusOK).JSON(response)
}

func RedirectToOriginalURL(c *fiber.Ctx) error {
    shortURL := c.Params("shortURL")
	// Redis'ten shortURL anahtarıyla orijinal URL'yi bul
	originalURL, err := database.RedisClient.Get(database.Ctx, shortURL).Result()
	if err == redis.Nil {
		// Anahtar Redis'te bulunmazsa hata döndür
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "URL not found",
		})
	} else if err != nil {
		// Redis'ten okuma hatası
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch URL from Redis",
		})
	}
	

	// Kullanıcıyı orijinal URL'ye yönlendir
	return c.Redirect(originalURL, fiber.StatusFound)
}


func ListURLs(c *fiber.Ctx) error {
	var urls []models.URL

	// Veritabanından tüm URL kayıtlarını çek
	if err := database.DB.Find(&urls).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch URLs",
		})
	}

	// Yanıt olarak JSON formatında tüm URL'leri döndür
	return c.Status(fiber.StatusOK).JSON(urls)
}


// GetURLStats kısaltılmış URL'nin istatistiklerini döner
func GetURLStats(c *fiber.Ctx) error {
	shortURL := c.Params("shortURL")

	// Veritabanından kısaltılmış URL'yi bul
	var urlRecord models.URL
	result := database.DB.Where("short_url = ?", shortURL).First(&urlRecord)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "URL not found",
		})
	}

	// İstatistikleri döndür
	return c.JSON(fiber.Map{
		"original_url": urlRecord.OriginalURL,
		"short_url":    fmt.Sprintf("http://localhost:3000/%s", urlRecord.ShortURL),
		"created_at":   urlRecord.CreatedAt,
		"expires_at":   urlRecord.ExpiresAt,
		"usage_count":  urlRecord.UsageCount,
	})
}

// DeleteShortURL kısaltılmış URL'yi redisten sil ve veri tabanında DeleteAt kısmını güncelle
func DeleteShortURL(c *fiber.Ctx) error {
    shortURL := c.Params("shortURL")

    // Redis'ten shortURL anahtarıyla orijinal URL'yi bul
    originalURL, err := database.RedisClient.Get(database.Ctx, shortURL).Result()
    if err == redis.Nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Redis: Key not found for shortURL",
        })
    } else if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to fetch URL from Redis",
        })
    }

    // PostgreSQL'den kısa URL kaydını bul
    var urlRecord models.URL
    result := database.DB.Where("short_url = ?", shortURL).First(&urlRecord)
    if result.Error != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "URL not found in database",
        })
    }

    // DeletedAt alanını güncelle
    now := time.Now()
    urlRecord.DeletedAt = &now
    if err := database.DB.Save(&urlRecord).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to update DeleteAt field in the database",
        })
    }

    // Redis'ten kısa URL anahtarını sil
    if err := deleteFromRedis(shortURL); err != nil {
        log.Printf("Redis deletion error: %v", err)
    }

    // Başarılı yanıt
    return c.JSON(fiber.Map{
        "message":      "URL successfully deleted",
        "short_url":    shortURL,
        "original_url": originalURL, // Yanıta ekledik
    })
}

// Redis'ten kısa URL anahtarını silme işlemi
func deleteFromRedis(shortURL string) error {
    ctx := database.Ctx
    err := database.RedisClient.Del(ctx, shortURL).Err()
    if err != nil {
        log.Printf("Error deleting key from Redis: %v", err)
        return err
    }

    log.Printf("Successfully deleted key %s from Redis", shortURL)
    return nil
}