package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/google/uuid"

	"golang-rest-api/db"
	"golang-rest-api/errs"
	"golang-rest-api/models"
)

// Pagination
const (
	PAGE_DEFAULT           = 1
	ITEMS_PER_PAGE_DEFAULT = 20
)

// AudiobookQuery holds the query string filters for GetAudiobooks.
// The schema tags are read by the HTTP decoder and the validate tags by the validator.
type AudiobookQuery struct {
	Name         string `schema:"name"`
	SortBy       string `schema:"sortBy"       validate:"omitempty,oneof=name modifiedOn"`
	OrderBy      string `schema:"orderBy"      validate:"omitempty,oneof=asc desc"`
	Date         string `schema:"date"         validate:"omitempty,datetime=2006-01-02"`
	Tag          int    `schema:"tag"          validate:"gte=0"`
	Page         int    `schema:"page"         validate:"gte=0"`
	ItemsPerPage int    `schema:"itemsPerPage" validate:"gte=0,lte=100"`
}

// audiobookSortColumns maps the sortBy values accepted by the API to DB columns.
// Only values from this map may reach ORDER BY, never the raw user input.
var audiobookSortColumns = map[string]string{
	"name":       "name",
	"modifiedOn": "updated_at",
}

type AudiobookRepository interface {
	GetAudiobooks(ctx context.Context, accountID string, q AudiobookQuery) ([]models.Audiobook, int, int, int, error)
	GetAudiobook(ctx context.Context, accountID string, id uuid.UUID, tag bool) (models.Audiobook, error)
	CreateAudiobook(ctx context.Context, audiobook *models.Audiobook) error
	DeleteAudiobook(ctx context.Context, accountID string, id uuid.UUID) error
	EditAudiobook(ctx context.Context, audiobook *models.Audiobook) error
}

type audiobookRepository struct {
	*db.BDData
}

func NewAudiobookRepository(db *db.BDData) AudiobookRepository {
	return &audiobookRepository{db}
}

func (ann *audiobookRepository) GetAudiobooks(ctx context.Context, accountID string, q AudiobookQuery) ([]models.Audiobook, int, int, int, error) {
	const op errs.Op = "repositories/AudiobookRepository.GetAudiobook"
	var audiosDB []models.Audiobook
	var totalRecords int64

	db := ann.DB.WithContext(ctx)

	db = db.Where("account_id = ?", accountID)

	if q.Name != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?)", "%"+q.Name+"%")
	}

	if q.Date != "" {
		day, err := time.Parse("2006-01-02", q.Date)
		if err != nil {
			return audiosDB, 0, PAGE_DEFAULT, ITEMS_PER_PAGE_DEFAULT, errs.E(op, errs.AB_ERR_019, err)
		}
		db = db.Where("updated_at >= ? AND updated_at < ?", day, day.AddDate(0, 0, 1))

	}

	if q.Tag > 0 {
		db = db.Where("id IN (?)", ann.DB.Table("audiobook_tags").
			Select("audiobook_id").Where("tag_id = ?", q.Tag))
	}

	page := q.Page
	itemsPerPage := q.ItemsPerPage

	offset := (page - 1) * itemsPerPage
	if page <= 1 {
		offset = 0
		if page == 0 {
			page = PAGE_DEFAULT
		}
	}
	if itemsPerPage <= 0 {
		itemsPerPage = ITEMS_PER_PAGE_DEFAULT
	}

	db = db.Model(&models.Audiobook{}).Session(&gorm.Session{})
	if err := db.Count(&totalRecords).Error; err != nil {
		return audiosDB, 0, PAGE_DEFAULT, ITEMS_PER_PAGE_DEFAULT, errs.E(op, errs.ERR_DB, err)

	}
	sortColumn := "name" //default or updated_at

	if q.SortBy != "" {
		col, ok := audiobookSortColumns[q.SortBy]
		if !ok {
			return audiosDB, 0, PAGE_DEFAULT, ITEMS_PER_PAGE_DEFAULT, errs.E(op, errs.AB_ERR_017, fmt.Errorf("invalid sortBy %q", q.SortBy))
		}
		sortColumn = col
	}

	if q.OrderBy == "desc" {
		sortColumn += " " + q.OrderBy
	}

	err := db.Order(sortColumn).Offset(offset).Limit(itemsPerPage).Preload("Tag").Find(&audiosDB).Error
	if err != nil {
		return audiosDB, 0, PAGE_DEFAULT, ITEMS_PER_PAGE_DEFAULT, errs.E(op, errs.ERR_DB, err)
	}

	return audiosDB, int(totalRecords), page, itemsPerPage, nil
}

func (ann *audiobookRepository) GetAudiobook(ctx context.Context, accountID string, id uuid.UUID, tag bool) (models.Audiobook, error) {
	const op errs.Op = "repositories/AudiobookRepository.GetAudiobook"

	var audio models.Audiobook
	var err error

	db := ann.DB.WithContext(ctx)
	if tag {
		db = db.Preload("Tag")
	}
	// scoped to the account: another account's audiobook is reported as not found
	err = db.Where("id = ? AND account_id = ?", id, accountID).First(&audio).Error
	if errs.Is(err, gorm.ErrRecordNotFound) {
		return audio, errs.E(op, errs.AB_ERR_001, err)
	}
	if err != nil {
		return audio, errs.E(op, errs.ERR_DB, err)
	}

	return audio, nil

}

func (ann *audiobookRepository) DeleteAudiobook(ctx context.Context, accountID string, id uuid.UUID) error {
	const op errs.Op = "repositories/AudiobookRepository.DeleteAudio"

	err := ann.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// check ownership first: deleting the associations only uses the ID,
		// so a scoped Delete alone would still remove another account's tag links
		var audio models.Audiobook
		if err := tx.Where("id = ? AND account_id = ?", id, accountID).First(&audio).Error; err != nil {
			return err
		}

		return tx.Unscoped().Select(clause.Associations).Delete(&audio).Error
	})
	if errs.Is(err, gorm.ErrRecordNotFound) {
		return errs.E(op, errs.AB_ERR_001, err)
	}
	if err != nil {
		return errs.E(op, errs.ERR_DB, err)
	}

	return nil
}

func (ann *audiobookRepository) CreateAudiobook(ctx context.Context, audiobook *models.Audiobook) error {
	const op errs.Op = "repositories/AudiobookRepository.CreateAudiobook"
	/*
		create := func(tx *gorm.DB) error{
			return tx.Create(&audiobook).Error
		}
	*/
	if err := doTransactionSlice(ann.DB.WithContext(ctx), []func(tx *gorm.DB) error{
		func(tx *gorm.DB) error {
			return tx.Create(&audiobook).Error
		},
	}); err != nil {
		return errs.E(op, errs.ERR_DB, err)
	}
	/*
		if err := ann.DB.Create(&audiobook).Error; err != nil{
			return errs.E(errs.Type(errs.EDATABASE), err)
		}
	*/

	return nil
}

func (ann *audiobookRepository) EditAudiobook(ctx context.Context, audio *models.Audiobook) error {
	const op errs.Op = "repositories/AudiobookRepository.EditAudiobook"

	db := ann.DB.WithContext(ctx)
	//if new tags, remove the old
	if audio.Tag != nil {
		if err := doTransactionSlice(ann.DB, []func(tx *gorm.DB) error{
			func(tx *gorm.DB) error { return tx.Model(&models.Audiobook{ID: audio.ID}).Association("Tag").Clear() },
			func(tx *gorm.DB) error { return tx.Save(audio).Error },
		},
		); err != nil {
			return errs.E(op, errs.ERR_DB, err)
		}
		return nil
	}

	if err := db.Save(audio).Error; err != nil {
		return errs.E(op, errs.ERR_DB, err)
	}

	return nil
}

func (ann *audiobookRepository) IsAudiobookInUse(id uuid.UUID) (bool, error) {
	const op errs.Op = "repositories/AudiobookRepository.IsAudiobookInUse"
	var exist bool

	err := ann.DB.Model(&models.Audiobook{}).Select("count(*) > 0").Where("id", id).Find(&exist).Error
	if err != nil {
		return false, errs.E(op, errs.ERR_DB, err)
	}

	return exist, nil
}

func doTransactionSlice(db *gorm.DB, fn [](func(tx *gorm.DB) error)) error {
	var err error

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for i := range fn {
		if err = fn[i](tx); err != nil {
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
