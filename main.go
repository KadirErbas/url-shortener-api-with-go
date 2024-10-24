package main

import (
	"log"
	"url-shortener/models"
	"url-shortener/database"
	"url-shortener/handler"
	"url-shortener/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"fmt"
)

func main() {

	
	// .env dosyasını yükle
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file:: %v", err)
	}


	database.ConnectDB()
	database.ConnectRedis()

	database.DB.AutoMigrate(&models.URL{})

	app := fiber.New()
	
	go listenForKeyExpiration()

	app.Post("/shorten",middleware.RateLimitByIP, handler.ShortenURL)

	app.Get("/", handler.ListURLs)
	app.Post("/shorten", handler.ShortenURL)
	app.Get("/:shortURL", handler.RedirectToOriginalURL)
	app.Get("/:shortURL/stats", handler.GetURLStats)
	app.Delete("/:shortURL", handler.DeleteShortURL)

	log.Fatal(app.Listen(":3000"))
}


// Redis keyspace notification'ları dinleyen fonksiyon
func listenForKeyExpiration() {
    pubsub := database.RedisClient.PSubscribe(database.Ctx, "__keyevent@0__:expired")

    // Süresi dolan anahtarları dinleme döngüsü
    for {
        msg, err := pubsub.ReceiveMessage(database.Ctx)
        if err != nil {
            log.Println("Error receiving Redis message:", err)
            continue
        }

        // Anahtarın süresi dolmuş. PostgreSQL'den silme işlemi yapıyoruz.
        fmt.Println("Expired key:", msg.Payload)
        deleteURLFromPostgres(msg.Payload)
    }
}

// PostgreSQL'den URL'yi silen fonksiyon
func deleteURLFromPostgres(shortURL string) {
    // 1. PostgreSQL'den kısa URL ile orijinal URL'yi bul
    var urlRecord models.URL
    result := database.DB.Where("short_url = ?", shortURL).First(&urlRecord)
    if result.Error != nil {
		log.Printf("URL not found: %v", result.Error)
    }else{
		now := time.Now()
		urlRecord.DeletedAt = &now
		if err := database.DB.Save(&urlRecord).Error;
		err != nil {
			log.Printf("Failed to update DeleteAt field in the database: %v", err)        
		}else {
			log.Printf("updated DeleteAt from database")
		}
	}
}
