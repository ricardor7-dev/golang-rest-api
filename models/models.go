package models

import (
	"time"
	"github.com/google/uuid"
)

type Audiobook struct{
	ID        				uuid.UUID 				`gorm:"primarykey;type:uuid;column:id;constraint:OnDelete:CASCADE;default:gen_random_uuid()"`
    CreatedAt 				time.Time
    UpdatedAt 				time.Time
    DeletedAt 				*time.Time 				`gorm:"index"`
	AccountID				string					`gorm:"not null;"`       
	Name					string					`gorm:"not null;"`
	Description 			string
	FilePath				string					`gorm:"not null;"`
	Tag 					[]*Tag 					`gorm:"many2many:audiobook_tags;constraint:OnDelete:CASCADE"`
}
//TODO: create one to one conection to Language

type Tag struct{
	ID						uint					`gorm:"primaryKey;constraint:OnDelete:CASCADE"`
	Name 					string					`gorm:"not null;"`
	Audiobook 				[]*Audiobook 			`gorm:"many2many:audiobook_tags;constraint:OnDelete:CASCADE"`
}


//////////////////////////////////////////
type Language struct{
	ID				uint			`gorm:"primaryKey;constraint:OnDelete:CASCADE"`
	Name			string			`gorm:"not null;"`
	Code			string			`gorm:"not null;"`
	GenderVoices	[]GenderVoice
}

type GenderVoice struct{
	ID				uint			`gorm:"primaryKey;constraint:OnDelete:CASCADE"`
	LanguageID		uint			`gorm:"not null;"`
	Gender			string			`gorm:"not null;"`
	VoiceName		string			`gorm:"not null;"`
}


////CHANGE TABLENAME
func (GenderVoice) TableName() string {
	return "gender_language_voice"
}