package config

import (
	"fmt"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DBHost               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBPort               string
	JWTSecret            string
	AccessTokenExpDays   int
	RefreshTokenExpDays  int
	CloudinaryCloudName  string
	CloudinaryAPIKey     string
	CloudinaryAPISecret  string
	CloudinaryBaseFolder string
}

func New() *Config {
	accessTokenExp, _ := strconv.Atoi(getEnv("ACCESS_TOKEN_EXP_DAYS", "1"))     
	refreshTokenExp, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_EXP_DAYS", "7")) 
	return &Config{
		DBHost:               getEnv("DB_HOST", "db"),
		DBUser:               getEnv("DB_USER", "user"),
		DBPassword:           getEnv("DB_PASSWORD", "satudua"),
		DBName:               getEnv("DB_NAME", "chat_app_db"),
		DBPort:               getEnv("DB_PORT", "5432"),
		JWTSecret:            getEnv("JWT_SECRET", "kirarabernstein"),
		AccessTokenExpDays:   accessTokenExp,
		RefreshTokenExpDays:  refreshTokenExp,
		CloudinaryCloudName:  getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:     getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret:  getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryBaseFolder: getEnv("CLOUDINARY_BASE_FOLDER", "chat-app"),
	}
}

func (c *Config) ConnectDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
