package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/poke-factory/cheri-berry/internal/models"
	"github.com/poke-factory/cheri-berry/internal/requests"
	"github.com/poke-factory/cheri-berry/internal/services"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

func Login(c *gin.Context) {
	var request requests.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Type == "user" {
		var okMessage string
		user, err := services.FindUser(request.Name)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		password := []byte(request.Password)

		if user == nil {
			hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
			err = services.CreateUser(models.User{
				Username:    request.Name,
				Password:    string(hashedPassword),
				Permissions: "npm",
			})

			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			okMessage = fmt.Sprintf("%s '%s' created", request.Type, request.Name)
		} else {
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), password)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
				return
			}
			if !strings.Contains(user.Permissions, "npm") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not allowed to login"})
				return
			}
			okMessage = fmt.Sprintf("you are authenticated as %s", request.Name)
		}

		token := GenToken(request.Name)

		if len(token) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": okMessage, "token": token})
	}
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": "you are logged out"})
}
