package repositories

import (

	"golang-rest-api/db"
	"golang-rest-api/errs"
	"golang-rest-api/models"
)

type TagRepository interface{
	GetTags() ([]models.Tag, error)
}

type tagRepository struct {
	*db.BDData
}


func NewTagRepository(db *db.BDData) TagRepository {
	return &tagRepository{db}
}

func (tg *tagRepository) GetTags() ([]models.Tag, error){
	const op errs.Op = "repositories/TagRepository.GetTags"
	var tags []models.Tag
	var err error

	err = tg.DB.Find(&tags).Error
	if err != nil {
		return tags, errs.E(op, errs.ERR_DB, err)
	}

	return tags, nil
}
