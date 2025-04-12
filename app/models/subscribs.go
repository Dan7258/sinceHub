package models

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
