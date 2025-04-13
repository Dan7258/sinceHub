package models

import "gorm.io/gorm"

type Subscribs struct {
	ProfilesId    uint64 `json:"profiles_id" gorm:"primaryKey;not null"`
	SubscribersID uint64 `json:"subscribers_id" gorm:"primaryKey;not null"`
}

func CheckMySubscribesForProfile(ID uint64, profileID uint64) bool {
	sub := new(Subscribs)
	result := DB.Where("profiles_id = ? AND subscribers_id = ?", profileID, ID).First(sub)
	if result.RowsAffected == 0 || result.Error != nil {
		return false
	}
	return true
}

func AddSubscriberToProfile(subID uint64, profileID uint64) error {
	sub := new(Subscribs)
	sub.ProfilesId = profileID
	sub.SubscribersID = subID
	result := DB.Model(sub).Create(sub)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeleteSubscriberFromProfile(subID uint64, profileID uint64) error {
	sub := new(Subscribs)
	sub.ProfilesId = profileID
	sub.SubscribersID = subID
	result := DB.Model(sub).Delete(sub)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetMySubscribersList(profileID uint64) ([]Profiles, error) {
	profile := new(Profiles)
	result := DB.Preload("SubscribersList", func(db *gorm.DB) *gorm.DB {
		return DB.Select("id, first_name, last_name, middle_name, country, vac, appointment")
	}).Preload("Publications").First(profile, profileID)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile.SubscribersList, nil
}

func GetMySubscribesList(profileID uint64) ([]Profiles, error) {
	profile := new(Profiles)
	result := DB.Preload("MySubscribesList", func(db *gorm.DB) *gorm.DB {
		return DB.Select("id, first_name, last_name, middle_name, country, vac, appointment")
	}).Preload("Publications").First(profile, profileID)
	if result.Error != nil {
		return nil, result.Error
	}
	return profile.MySubscribesList, nil
}
