package db

import (
	model "github.com/helloworldyuhaiyang/mail-handle/internal/models"
	"gorm.io/gorm"
)

type ForwardTargetsRepo struct {
	db *gorm.DB
}

func NewForwardTargetsRepo(db *gorm.DB) *ForwardTargetsRepo {
	return &ForwardTargetsRepo{db: db}
}

func (r *ForwardTargetsRepo) FindEmailByName(targetName string) (string, error) {
	// return "helloworldyang9@gmail.com", nil
	var forwardTarget model.ForwardTargets
	if err := r.db.Model(&model.ForwardTargets{}).Where("name = ?", targetName).First(&forwardTarget).Error; err != nil {
		return "", err
	}

	return forwardTarget.Email, nil
}
