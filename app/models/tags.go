package models

import (
	"fmt"
)

type Tags struct {
	ID           uint64         `gorm:"primaryKey"`
	Name         string         `gorm:"size:1000;not null" validate:"required,min=2,max=1000"`
	Publications []Publications `gorm:"many2many:publication_tags;"`
}

func CreateTag(name string) error {
	tag := new(Tags)
	tag.Name = name
	result := DB.Create(tag)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetTagByID(ID int) (*Tags, error) {
	tag := new(Tags)
	result := DB.Preload("Publications").First(tag, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return tag, nil
}

func GetTagByName(name string) (*Tags, error) {
	tag := new(Tags)
	result := DB.Preload("Publications").Where("name = ?", name).First(tag)
	if result.Error != nil {
		return nil, result.Error
	}
	return tag, nil
}

func DeleteTagByID(ID int) error {
	tag := new(Tags)
	result := DB.Delete(tag, ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Тег с ID %d не найден", ID)
	}
	return nil
}

func UpdateTagByID(ID int, updTag *Tags) error {
	tag := new(Tags)
	result := DB.Preload("Publications").Model(tag).Where("id = ?", ID).Update("name", updTag.Name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Тег с ID %d не найден", ID)
	}
	return nil
}

func GetAllTags() ([]Tags, error) {
	var tags []Tags
	result := DB.Preload("Publications").Find(&tags)
	if result.Error != nil {
		return nil, result.Error
	}
	return tags, nil
}

func AddPublicationsToTag(ID uint64, pubIDs []uint64) error {
	tag := new(Tags)
	var pubs []Publications
	result := DB.First(tag, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(tag).Association("Publications").Append(pubs)
	if err != nil {
		return err
	}
	return nil
}

func DeletePublicationsFromTag(ID uint64, pubIDs []uint64) error {
	tag := new(Tags)
	var pubs []Publications
	result := DB.First(tag, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(tag).Association("Publications").Delete(pubs)
	if err != nil {
		return err
	}
	return nil
}
