package models

import (
	"fmt"
	"time"
)

type Publications struct {
	ID        uint64 `gorm:"primaryKey"`
	Title     string `gorm:"size:1000;not null" validate:"required,min=2,max=1000"`
	Abstract  string `gorm:"size:1000;"`
	Content   string `gorm:"type:text;not null" validate:"required,min=2"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func CreatePublication(pub *Publications) error {
	result := DB.Create(pub)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeletePublicationByID(ID int) error {
	result := DB.Delete(new(Publications), ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Публикация с ID %d не найден", ID)
	}
	return nil
}

func UpdatePublicationByID(ID int, updPub *Publications) error {
	result := DB.Model(new(Publications)).Where("id = ?", ID).Updates(updPub)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Публикация с ID %d не найден", ID)
	}
	return nil
}

func GetPublicationByID(ID int) (*Publications, error) {
	pub := new(Publications)
	result := DB.First(pub, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return pub, nil
}

func GetAllPublications() ([]Publications, error) {
	var pub []Publications
	result := DB.Find(&pub)
	if result.Error != nil {
		return nil, result.Error
	}
	return pub, nil
}
