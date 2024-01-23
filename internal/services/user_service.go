package services

import (
	"errors"
	"github.com/poke-factory/cheri-berry/internal/database"
	"github.com/poke-factory/cheri-berry/internal/models"
	"gorm.io/gorm"
)

func FindUser(username string) (*models.User, error) {
	var user models.User
	result := database.DB.Where("username = ?", username).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func CreateUser(user models.User) error {
	result := database.DB.Create(&user)
	return result.Error
}
