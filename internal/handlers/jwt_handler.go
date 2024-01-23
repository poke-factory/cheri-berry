package handlers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/poke-factory/cheri-berry/internal/config"
	"log"
	"time"
)

type CustomClaims struct {
	jwt.StandardClaims
	// 这里可以添加你自定义的声明字段
	Username string `json:"username"`
}

func GenToken(username string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 设置过期时间
			Issuer:    "snorlax",                             // 设置签发者
		},
		Username: username, // 自定义字段
	})

	tokenString, err := token.SignedString([]byte(config.Cfg.JwtSecret))

	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return ""
	}

	return tokenString
}
