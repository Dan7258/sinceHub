package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Publications struct {
	ID        uint64     `json:"id" gorm:"primaryKey"`
	Title     string     `json:"title" gorm:"size:1000;not null" validate:"omitempty,min=2,max=1000"`
	Abstract  string     `json:"abstract" gorm:"size:1000;" validate:"omitempty,min=2"`
	FileLink  string     `json:"file_link" gorm:"not null"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	OwnerID   uint64     `json:"owner_id" gorm:"not null"`
	Owner     *Profiles  `json:"owner" gorm:"foreignKey:OwnerID"`
	Profiles  []Profiles `json:"profiles" gorm:"many2many:profile_publications;constraint:OnDelete:CASCADE"`
	Tags      []Tags     `json:"tags" gorm:"many2many:publication_tags;constraint:OnDelete:CASCADE"`
}

type Paginator struct {
	FirstID int `json:"first_id"`
	Count   int `json:"count"`
}

type SearchDataForPublications struct {
	Title string   `json:"title"`
	Tags  []uint64 `json:"tags"`
	Paginator
}

type TypeFile uint64

const (
	Word TypeFile = iota
	Exel
	LibraWord
	LibraExcel
)

type PublicationDownloadFiltres struct {
	CountPublications uint64
	DateStart         time.Time
	DateEnd           time.Time
	Type              TypeFile
}

func CreatePublication(pub *Publications, tagIDs []uint64, coauthorIDs []uint64) error {
	result := DB.Create(pub)
	if result.Error != nil {
		return result.Error
	}
	tags, err := GetTagsByID(tagIDs)
	if err != nil {
		return err
	}
	result = DB.Find(&tags, tagIDs)
	DB.Model(pub).Association("Tags").Append(tags)
	if result.Error != nil {
		return result.Error
	}
	var profiles []Profiles
	result = DB.Find(&profiles, coauthorIDs)
	DB.Model(pub).Association("Profiles").Append(profiles)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func DeletePublicationByID(ID uint64) error {
	result := DB.Delete(new(Publications), ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("Публикация с ID %d не найден", ID)
	}
	return nil
}

func UpdatePublication(pub *Publications, tagIDs []uint64, coauthorIDs []uint64) error {
	result := DB.Model(new(Publications)).Where("id = ?", pub.ID).Updates(pub)
	if result.Error != nil {
		return fmt.Errorf("Публикация с ID %d не найден", pub.ID)
	}
	var tags []Tags
	result = DB.Find(&tags, tagIDs)
	DB.Model(pub).Association("Tags").Replace(tags)
	if result.Error != nil {
		return result.Error
	}
	var profiles []Profiles
	result = DB.Find(&profiles, coauthorIDs)
	DB.Model(pub).Association("Profiles").Replace(profiles)
	if result.Error != nil {
		return result.Error
	}

	for _, coauthorID := range coauthorIDs {
		DeleteDataFromRedis(fmt.Sprintf("%d", coauthorID))
	}
	return nil
}

func GetPublicationByID(ID uint64) (*Publications, error) {
	pub := new(Publications)
	result := DB.Preload("Tags").Preload("Profiles", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name, middle_name")
	}).First(pub, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return pub, nil
}

func GetPublicationsByID(idList []uint64) ([]Publications, error) {
	publications := make([]Publications, 0)
	result := DB.Model(new(Publications)).Preload("Tags").Preload("Profiles", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name, middle_name")
	}).Where("id = ?", idList).Find(publications)
	return publications, result.Error
}

func GetPublicationsWithSearchParams(data SearchDataForPublications) ([]Publications, error) {
	publications := make([]Publications, 0)
	query := DB.Model(new(Publications)).Distinct("publications.*").
		Preload("Tags").
		Preload("Profiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		}).
		Where("id >= ?", data.FirstID).
		Order("created_at asc").
		Limit(data.Count)
	if query.Error != nil {
		return nil, query.Error
	}
	if data.Tags != nil && len(data.Tags) > 0 {
		query.Joins("left join publication_tags on publication_tags.publications_id = publications.id").
			Where("publication_tags.tags_id IN (?)", data.Tags)
	}
	if data.Title != "" {
		query.Where("title LIKE ?", fmt.Sprintf("%%%s%%", data.Title))
	}
	err := query.Find(&publications).Error
	return publications, err
}

func GetLastPublications(paginator Paginator) ([]Publications, error) {
	publications := make([]Publications, 0)
	result := DB.Model(new(Publications)).Preload("Tags").Preload("Profiles", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name, middle_name")
	}).Where("id >= ?", paginator.FirstID).Order("created_at asc").Limit(paginator.Count).Find(&publications)
	fmt.Println(publications)
	return publications, result.Error
}

func GetAllPublications() ([]Publications, error) {
	var pub []Publications
	result := DB.Preload("Tags").Preload("Profiles", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name, middle_name")
	}).Find(&pub)
	if result.Error != nil {
		return nil, result.Error
	}
	return pub, nil
}

func AddTagsToPublication(ID uint64, tagIDs []uint64) error {

	pub := new(Publications)
	var tags []Tags
	result := DB.First(pub, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&tags, tagIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(pub).Association("Tags").Append(tags)
	if err != nil {
		return err
	}
	return nil
}

func DeleteTagsFromPublication(ID uint64, tagIDs []uint64) error {
	pub := new(Publications)
	var tags []Tags
	result := DB.First(pub, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&tags, tagIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(pub).Association("Tags").Delete(tags)
	if err != nil {
		return err
	}
	return nil
}

func AddProfilesToPublication(ID uint64, profileIDs []uint64) error {
	pub := new(Publications)
	var profiles []Profiles
	result := DB.First(pub, ID)
	if result.Error != nil {
		return result.Error
	}
	result = DB.Find(&profiles, profileIDs)
	if result.Error != nil {
		return result.Error
	}
	err := DB.Model(pub).Association("Profiles").Append(profiles)
	if err != nil {
		return err
	}
	return nil
}

func DeleteProfileFromPublication(ID uint64, profileID uint64) error {
	pub := new(Publications)
	result := DB.First(pub, ID)
	if result.Error != nil {
		return result.Error
	}
	if pub.OwnerID == profileID {
		return fmt.Errorf("нельзя удалить владельца публикации")
	}
	profile := new(Profiles)
	result = DB.First(profile, profileID)
	err := DB.Model(pub).Association("Profiles").Delete(profile)
	if err != nil {
		return err
	}
	return nil
}

func GetPublicationListByFilters(userID uint64, filters PublicationDownloadFiltres) ([]Publications, error) {
	publications := make([]Publications, 0)
	query := DB.Model(new(Publications)).
		Joins("JOIN profile_publications ON profile_publications.publications_id = publications.id").
		Where("profile_publications.profiles_id = ?", userID).
		Preload("Profiles", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, first_name, last_name, middle_name")
		})
	if !filters.DateStart.IsZero() {
		query = query.Where("created_at >= ?", filters.DateStart)
	}
	if !filters.DateEnd.IsZero() {
		query = query.Where("created_at <= ?", filters.DateEnd)
	}
	if filters.CountPublications > 0 {
		query = query.Limit(int(filters.CountPublications))
	}
	result := query.Order("created_at desc").Find(&publications)
	if result.Error != nil {
		return nil, result.Error
	}
	return publications, nil
}
