package db

import "gorm.io/gorm"

type ForwardTargetsRepo struct {
	db *gorm.DB
}

func NewForwardTargetsRepo(db *gorm.DB) *ForwardTargetsRepo {
	return &ForwardTargetsRepo{db: db}
}

func (r *ForwardTargetsRepo) FindEmailByName(targetName string) (string, error) {
	panic("not implemented")
}
