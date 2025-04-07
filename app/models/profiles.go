package models

import (
	"fmt"
	"gorm.io/gorm"
)

type Profiles struct {
	ID               uint64         `json:"id" gorm:"primaryKey"`
	Login            string         `json:"login" gorm:"size:1000;not null;unique" validate:"omitempty,email,min=4,max=1000"`
	Password         string         `json:"password" gorm:"size:1000;not null" validate:"omitempty,min=8,max=1000"`
	FirstName        string         `json:"first_name" gorm:"size:1000;not null" validate:"omitempty,min=3,max=1000"`
	LastName         string         `json:"last_name" gorm:"size:1000;not null" validate:"omitempty,min=3,max=1000"`
	MiddleName       string         `json:"middle_name" gorm:"size:1000;" validate:"max=1000"`
	Country          string         `json:"country" gorm:"size:100;" validate:"max=100"`
	AcademicDegree   string         `json:"academic_degree" gorm:"size:1000;" validate:"max=1000"`
	VAC              string         `json:"vac" gorm:"size:1000;" validate:"max=1000"`
	Appointment      string         `json:"appointment" gorm:"size:1000;" validate:"max=1000"`
	Publications     []Publications `gorm:"many2many:profile_publications;"`
	SubscribersList  []Profiles     `gorm:"many2many:subscribs;joinForeignKey:profiles_id;joinReferences:subscribers_id"`
	MySubscribesList []Profiles     `gorm:"many2many:subscribs;joinForeignKey:subscribers_id;joinReferences:profiles_id"`
}

func CreateProfile(profile *Profiles) error {
	result := DB.Create(profile)
	if result.Error != nil {
		return result.Error
	}
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetProfileByID(ID uint64) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.Select("id, first_name, last_name, middle_name, country, vac, appointment").Preload("Publications").Preload("SubscribersList").Preload("MySubscribesList").First(profile, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func ThsProfilesIsExist(login string) bool {
	profile := new(Profiles)
	result := DB.Select("login").Where("login = ?", login).First(profile)
	if result.Error != nil {
		return false
	}
	return true
}

func GetUserProfile(ID uint64) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.
		Preload("Publications.Profiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		Preload("Publications.Tags").
		Preload("SubscribersList").
		Preload("MySubscribesList").
		First(profile, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func GetProfileLoginData(login string) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.Select("id, login, password").Where("login = ?", login).First(profile)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func GetAllProfiles() ([]Profiles, error) {
	var profiles []Profiles
	result := DB.Preload("Publications").Preload("SubscribersList").Preload("MySubscribesList").Find(&profiles)
	if result.Error != nil {
		return nil, result.Error
	}
	return profiles, nil
}

func GetAllProfileIDAndNames() ([]Profiles, error) {
	var profiles []Profiles
	result := DB.Select("id, first_name, last_name, middle_name").Find(&profiles)
	if result.Error != nil {
		return nil, result.Error
	}
	return profiles, nil
}

func DeleteProfileByID(ID uint64) error {
	profile := new(Profiles)
	result := DB.Delete(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Профиль с ID %d не найден", ID)
	}
	return nil
}

func DeleteProfileByLogin(login string) error {
	profile := new(Profiles)
	result := DB.Delete(profile, login)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Профиль с login: %s не найден", login)
	}
	return nil
}

func UpdateProfileByID(ID uint64, updProfile *Profiles) error {
	result := DB.Model(new(Profiles)).Where("id = ?", ID).Updates(updProfile)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Профиль с ID %d не найден", ID)
	}
	return nil
}

func UpdateProfileByLogin(login string, updProfile *Profiles) error {
	result := DB.Model(new(Profiles)).Where("login = ?", login).Updates(updProfile)
	if result.Error != nil {
		return result.Error
	}
	//if result.RowsAffected == 0 {
	//	return fmt.Errorf("Профиль с login: %s не найден", login)
	//}
	return nil
}

func AddPublicationsToProfile(ID uint64, pubIDs []uint64) error {
	profile := new(Profiles)
	var pubs []Publications
	result := DB.First(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(profile).Association("Publications").Append(pubs)
	if err != nil {
		return err
	}
	return nil
}

func DeletePublicationsFromProfile(ID uint64, pubIDs []uint64) error {
	profile := new(Profiles)
	var pubs []Publications
	result := DB.First(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&pubs, pubIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(profile).Association("Publications").Delete(pubs)
	if err != nil {
		return err
	}
	return nil
}

func AddSubscribersToProfile(ID uint64, subIDs []uint64) error {
	profile := new(Profiles)
	var subs []Profiles
	result := DB.First(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&subs, subIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(profile).Association("SubscribersList").Append(subs)
	if err != nil {
		return err
	}
	return nil
}

func DeleteSubscribersFromProfile(ID uint64, subIDs []uint64) error {
	profile := new(Profiles)
	var subs []Profiles
	result := DB.First(profile, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&subs, subIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(profile).Association("SubscribersList").Delete(subs)
	if err != nil {
		return err
	}
	return nil
}
