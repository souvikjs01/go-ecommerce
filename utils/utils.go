package utils

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt/v5"
	"github.com/souvikjs01/go-ecommerce/config"
	"golang.org/x/crypto/bcrypt"
)

// hashpassword
func HashPassword(password string) (string, error) {
	hashedPass_bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(hashedPass_bytes), err
}

// verify Password
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetRedis() *redis.Client {

	// Get the Redis URI
	cfg, _ := config.SetConfig()
	REDIS_URI := cfg.UPSTASH_URI
	fmt.Println("REDIS_URI", REDIS_URI)

	opt, _ := redis.ParseURL(REDIS_URI)
	client := redis.NewClient(opt)
	fmt.Println("client", client)

	return client
}

func CreateJWTToken(userId string, username string, isAdmin bool) (string, error) {
	config, _ := config.SetConfig()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":       userId,
			"username": username,
			"isAdmin":  isAdmin,
			"exp":      time.Now().Add(time.Hour * 72).Unix(),
		},
	)

	tokenString, err := token.SignedString([]byte(config.JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
