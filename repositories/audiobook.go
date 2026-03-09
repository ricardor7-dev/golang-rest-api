package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/google/uuid"

	"golang-rest-api/db"
	"golang-rest-api/errs"
	"golang-rest-api/models"
)

//Pagination
const (
	PAGE_DEFAULT           = 1
	ITEMS_PER_PAGE_DEFAULT = 20
)

type AudiobookRepository interface{
	GetAudiobooks(ctx context.Context, filters map[string]interface{}) ([]models.Audiobook, int, int, int, error)
	GetAudiobook(ctx context.Context, id uuid.UUID, tag bool) (models.Audiobook, error)
	CreateAudiobook(ctx context.Context, audiobook *models.Audiobook) error
	DeleteAudiobook(ctx context.Context, id uuid.UUID) error
	EditAudiobook(ctx context.Context, audiobook *models.Audiobook) error
}

type audiobookRepository struct {
	*db.BDData
}

func NewAudiobookRepository(db *db.BDData) AudiobookRepository {
	return &audiobookRepository{db}
}

func (ann * audiobookRepository) GetAudiobooks(ctx context.Context, filters map[string]interface{}) ([]models.Audiobook, int, int, int, error) {
	const op errs.Op = "repositories/AudiobookRepository.GetAudiobook"
	var audiosDB []models.Audiobook
	var totalRecords int64

	headerParams := filters["headerParams"].(map[string]string)

	db := ann.DB.WithContext(ctx)

	db = db.Where(&models.Audiobook{AccountID: headerParams["x-kong-jwt-claim-serviceaccount"]})

	if name, ok := filters["name"].(string); ok && name != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%")
	}

	if date, ok := filters["date"].(string); ok && date != ""{
		dateTime, _ := time.Parse("2006-01-02", date)
		db = db.Where("updated_at = ?", dateTime)
	}

	if tag:=filters["tag"].(int); tag>0{
		db.Where("id IN (?)", ann.DB.Table("audiobook_tags").
		Select("audiobook_id").Where("tag_id = ?", tag))
	}

	page := filters["page"].(int)
	itemsPerPage := filters["itemsPerPage"].(int)

	offset := (page-1) * itemsPerPage
	if page <= 1 {
		offset = 0
		if page == 0 {
			page = PAGE_DEFAULT
		}
	}
	if itemsPerPage <= 0 {
		itemsPerPage = ITEMS_PER_PAGE_DEFAULT
	}
	db.Model(&models.Audiobook{}).Count(&totalRecords)
	db.Offset(offset).Limit(itemsPerPage)


	sortOrder := "name"	//default or updated_at
	if sortBy := filters["sortBy"].(string); sortBy !="" && sortBy != "name" {
		//make a dictionary for this
		if sortBy == "modifiedOn"{
			sortOrder = "updated_at"
		}else{
			sortOrder = sortBy
		}
	}

	if orderBy := filters["orderBy"].(string); orderBy == "desc" {
		sortOrder += " " + orderBy 
	}


	err := db.Order(sortOrder).Preload("Tag").Find(&audiosDB).Error
	if err != nil {
		return audiosDB, 0, PAGE_DEFAULT, ITEMS_PER_PAGE_DEFAULT, errs.E(op, errs.ERR_DB, err)
	}


	return audiosDB, int(totalRecords), page, itemsPerPage, nil
}

func (ann * audiobookRepository) GetAudiobook(ctx context.Context, id uuid.UUID, tag bool) (models.Audiobook, error){
	const op errs.Op = "repositories/AudiobookRepository.GetAudiobook"

	var audio models.Audiobook
	var err error

	db := ann.DB.WithContext(ctx)
	if tag{
		db.Preload("Tag")
	}
	err = db.First(&audio, id).Error
	if errs.Is(err, gorm.ErrRecordNotFound) {
		return audio, errs.E(op, errs.AB_ERR_001, err)
	}
	if err != nil {
		return audio, errs.E(op, errs.ERR_DB, err)
	}

	return audio, nil

}

func (ann *audiobookRepository) DeleteAudiobook(ctx context.Context, id uuid.UUID) error{
	const op errs.Op = "repositories/AudiobookRepository.DeleteAudio"
	var err error

	err = ann.WithContext(ctx).Unscoped().Select(clause.Associations).Delete(&models.Audiobook{ID: id}).Error
	if err !=nil{
		return errs.E(op, errs.ERR_DB, err)
	}

	return nil
}

func (ann *audiobookRepository) CreateAudiobook(ctx context.Context, audiobook *models.Audiobook) error{
	const op errs.Op = "repositories/AudiobookRepository.CreateAudiobook"
	/*
	create := func(tx *gorm.DB) error{
		return tx.Create(&audiobook).Error
	}
	*/
	if err := doTransactionSlice(ann.DB.WithContext(ctx), []func(tx *gorm.DB) error{ 
		func(tx *gorm.DB) error{
			return tx.Create(&audiobook).Error
		},
	}); err != nil{
		return errs.E(op, errs.ERR_DB, err)
	}
	/*
	if err := ann.DB.Create(&audiobook).Error; err != nil{
		return errs.E(errs.Type(errs.EDATABASE), err)
	}
	*/

	return nil
}

func  (ann *audiobookRepository) EditAudiobook(ctx context.Context, audio *models.Audiobook) error{
	const op errs.Op = "repositories/AudiobookRepository.EditAudiobook"

	db := ann.DB.WithContext(ctx)
	//if new tags, remove the old
	if audio.Tag !=nil{
		db.Model(&models.Audiobook{ID: audio.ID}).Association("Tag").Clear()
	}

	if err := db.Save(audio).Error; err != nil {
		return errs.E(op, errs.ERR_DB, err)
	}

	return nil
}


func (ann *audiobookRepository) IsAudiobookInUse(id uuid.UUID) (bool, error){
	const op errs.Op = "repositories/AudiobookRepository.IsAudiobookInUse"
	var exist bool

	err := ann.DB.Model(&models.Audiobook{}).Select("count(*) > 0").Where("id", id).Find(&exist).Error
	if err != nil{
		return false, errs.E(op, errs.ERR_DB, err)
	}

	return exist, nil
}


func doTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error{
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	err := fn(tx)
	if err != nil {
		if e := tx.Rollback().Error; e != nil {
			return e
		}
		return err
	}
	if err = tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func doTransactionSlice(db *gorm.DB, fn [](func(tx *gorm.DB) error)) error{
	var err error

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for i:= range fn{
		if err = fn[i](tx); err != nil{
			if e := tx.Rollback().Error; e != nil {
				return e
			}
			return err
		}
	}

	if err = tx.Commit().Error; err != nil {
		return err
	}

	return nil
}