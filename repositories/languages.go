package repositories

import (
	"golang-rest-api/db"
	"golang-rest-api/errs"
	"golang-rest-api/models"
)

type LanguagesRepository interface {
	GetTTSLanguages(genderFiltler bool) ([]models.Language, error)
	GetTTSGendersByLangCode(langCode string) ([]models.GenderVoice, error)
}

type languagesRepository struct {
	*db.BDData
}

func NewLanguagesRepository(db *db.BDData) LanguagesRepository {
	return &languagesRepository{db}
}

func (tg *languagesRepository) GetTTSLanguages(genderFiltler bool) ([]models.Language, error) {
	const op errs.Op = "repositories/LanguagesTTSRepository.GetTTSLanguages"
	var languages []models.Language

	query := tg.DB

	if genderFiltler {
		query = tg.DB.Preload("GenderVoices")
	}

	err := query.Find(&languages).Error
	if err != nil {
		return languages, errs.E(op, errs.ERR_DB, err)
	}

	return languages, nil
}

func (tg *languagesRepository) GetTTSGendersByLangCode(langCode string) ([]models.GenderVoice, error){
	const op errs.Op = "repositories/LanguagesTTSRepository.GetTTSGendersByLangCode"
	var genders []models.GenderVoice

	err := tg.DB.Where("language_id = (?)", tg.DB.Table("languages").Select("id").Where("code = ?", langCode)).Find(&genders).Error
	if err != nil {
		return genders, errs.E(op, errs.ERR_DB, err)
	}

	return genders, nil
}