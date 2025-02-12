package models

import "fmt"

type Profiles struct {
	ID                  uint64 `json:"id" gorm:"primaryKey"`
	Login               string `json:"login" gorm:"size:1000;not null;unique" validate:"required,min=4,max=1000"`
	Password            string `json:"password" gorm:"size:1000;not null" validate:"required,min=6,max=1000"`
	FirstName 			string `json:"first_name" gorm:"size:1000;not null" validate:"required,min=2,max=1000"`
	LastName 			string `json:"last_name" gorm:"size:1000;not null" validate:"required,min=2,max=1000"`
	MiddleName          string `json:"middle_name" gorm:"size:1000;" validate:"max=1000"`
	Country             string `json:"country" gorm:"size:100;" validate:"max=100"`
	AcademicDegree      string `json:"academin_degree" gorm:"size:1000;" validate:"max=1000"`
	VAC                 string `json:"vac" gorm:"size:1000;" validate:"max=1000"`
	Appointment         string `json:"appointment" gorm:"size:1000;" validate:"max=1000"`
	Subscribers         uint64 `json:"subscribers" gorm:"default:0"`
	MySubscribes        uint64 `json:"my_subscribes" gorm:"default:0"`
}


func CreateProfile(profile *Profiles) error {
	result := DB.Create(profile)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetProfileByID(ID int) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.First(profile, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func GetProfileByLogin(login string) (*Profiles, error) {
	profile := new(Profiles)
	result := DB.Where("login = ?", login).First(profile)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile, nil
}

func DeleteProfileByID(ID int) error {
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

func UpdateProfileByID(ID int, updProfile *Profiles) error {
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
	if result.RowsAffected == 0 {
		return fmt.Errorf("Профиль с login: %s не найден", login)
	}
	return nil
}

func GetAllProfiles() ([]Profiles, error) {
	var profiles []Profiles
	result := DB.Find(&profiles)
	if result.Error != nil {
		return nil, result.Error
	}
	return profiles, nil
}