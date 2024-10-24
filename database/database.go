package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Global veritabanı bağlantı nesneleri
var (
	DB  *gorm.DB
	RedisClient *redis.Client
	Ctx = context.Background() 
)


func ConnectDB() {
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	port := os.Getenv("POSTGRES_PORT")
	
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	} else {
		fmt.Println("Connected to PostgreSQL!")
	}
}

// Redis bağlantısını başlatma
func ConnectRedis() {
	// Çevresel değişkenlerden Redis bilgilerini al
	addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	password := os.Getenv("REDIS_PASSWORD")
	db := os.Getenv("REDIS_DB")
	intDB, err := strconv.Atoi(db)
	if err != nil {
		log.Fatalf("Redis DB değeri integer'a çevrilemedi: %v", err)
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, 
		DB:       intDB,    
	})

	// Redis bağlantısını test et
	_, pingErr := RedisClient.Ping(Ctx).Result()
	if pingErr != nil {
		log.Fatalf("Redis bağlantısı kurulamadı: %v", pingErr)
	}

	log.Println("Redis bağlantısı başarılı")
}
