package server

import (
	"time"

	"golang-rest-api/models"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type AudiobookJSON struct{
	ID                	uuid.UUID              		`json:"id,omitempty"`
	CreatedOn         	*time.Time        			`json:"createdOn,omitempty"`
	ModifiedOn        	*time.Time        			`json:"modifiedOn,omitempty"`
	Name             	string            			`json:"name,omitempty"`// validate:"required,min=1,max=32"`
	Description			string						`json:"description,omitempty"`
	FilePath			string						`json:"filePath,omitempty"`
	UserID				string						`json:"userID,omitempty"`
	UserName			string						`json:"userName,omitempty"`
	InUseList			*[]string					`json:"inUseList,omitempty"`
	Tag					[]TagJSON					`json:"tag,omitempty"`
	InUseBool			*bool						`json:"inUse,omitempty"`
}

type TagJSON struct{
	ID					uint						`json:"id,omitempty"`
	Name				string						`json:"name,omitempty"`
}

type RecordsJSON struct {
	Records      interface{} 	`json:"records"`
}
type PagedResults struct {
	Page         int       		`json:"page"`
	ItemsPerPage int		 	`json:"itemsPerPage"`
	Records      interface{} 	`json:"records"`
	TotalRecords int         	`json:"total_records"`
}

func NewAudiobooksResponse(audiobookBD *[]models.Audiobook) []AudiobookJSON{
	audiobookResponse := []AudiobookJSON{}

	for i := range *audiobookBD{
		audio := AudiobookJSON{
			ID: (*audiobookBD)[i].ID,
			CreatedOn: &(*audiobookBD)[i].CreatedAt,
			ModifiedOn: &(*audiobookBD)[i].UpdatedAt,
			Name: (*audiobookBD)[i].Name,
			Description: (*audiobookBD)[i].Description,
			FilePath: (*audiobookBD)[i].FilePath,
		}
		
		audio.Tag = []TagJSON{}
		for _, tag := range (*audiobookBD)[i].Tag{
			tagJson := TagJSON{}
			tagJson.ID	=	tag.ID
			tagJson.Name	= tag.Name
			audio.Tag =append(audio.Tag, tagJson)
		}
		
		audiobookResponse = append(audiobookResponse, audio)
	}

	return audiobookResponse
}


func NewAudiobookResponse(audiobookDB *models.Audiobook) AudiobookJSON{
	audiobookResponse := AudiobookJSON{
		ID: audiobookDB.ID,
		CreatedOn: &audiobookDB.CreatedAt,
		ModifiedOn: &audiobookDB.UpdatedAt,
		Name: audiobookDB.Name,
		Description: audiobookDB.Description,
		FilePath: audiobookDB.FilePath,
	}

	audiobookResponse.Tag = []TagJSON{}
	for _, tag := range audiobookDB.Tag{
		tagJson := TagJSON{
			ID: tag.ID,
			Name: tag.Name,
		}

		audiobookResponse.Tag = append(audiobookResponse.Tag, tagJson)
	}

	return audiobookResponse
}

func NewTagResponse(tagDB *[]models.Tag) []TagJSON{
	tags := []TagJSON{}

	for _, t := range *tagDB{
		tag := TagJSON{
			ID: t.ID,
			Name: t.Name,
		}
		tags = append(tags, tag)
	}

	return tags
}

type LanguageJSON struct{
	Alias			string				`json:"alias,omitempty"`
	Code			string				`json:"code,omitempty"`
	Gender			*[]AliasCodeJSON	`json:"gender,omitempty"`
}

type AliasCodeJSON struct{
	Alias			string		`json:"alias,omitempty"`
	Code			string		`json:"code,omitempty"`	
}


func NewLanguagesResponse(languages *[]models.Language) []LanguageJSON{
	ttsLanguages := []LanguageJSON{}
	
	for i := range *languages{
		lang := &LanguageJSON{
			Alias: (*languages)[i].Name,
			Code: (*languages)[i].Code,
		}

		genders := []AliasCodeJSON{}

		for j := range (*languages)[i].GenderVoices{
			gender := &AliasCodeJSON{}
			if len(genders)>1{
				break
			}else if (*languages)[i].GenderVoices[j].Gender =="male" && slices.ContainsFunc((*languages)[i].GenderVoices, func( c models.GenderVoice) bool { return c.Gender =="male"}){
				gender.Alias = (*languages)[i].GenderVoices[j].Gender
				gender.Code = "1"
			}else if (*languages)[i].GenderVoices[j].Gender =="female" && 
			slices.ContainsFunc((*languages)[i].GenderVoices, func( c models.GenderVoice) bool { return c.Gender =="female"}){
				gender.Alias = (*languages)[i].GenderVoices[j].Gender
				gender.Code = "0"
			}

			genders = append(genders, *gender)
		}

		if len(genders)> 0{
			lang.Gender =  &genders
		}

		ttsLanguages = append(ttsLanguages, *lang)
	}

	return ttsLanguages
}

func NewTTSGendersResponse(genders *[]models.GenderVoice) []AliasCodeJSON{
	ttsGenders := []AliasCodeJSON{}

	for i := range *genders{
		gender := &AliasCodeJSON{}
		if len(ttsGenders)>1{
			break
		}else if (*genders)[i].Gender =="male" && 
		slices.ContainsFunc(*genders, func( c models.GenderVoice) bool { return c.Gender =="male"}){
			gender.Alias = (*genders)[i].Gender
			gender.Code = "1"
		}else if (*genders)[i].Gender =="female" && 
		slices.ContainsFunc(*genders, func( c models.GenderVoice) bool { return c.Gender =="female"}){
			gender.Alias = (*genders)[i].Gender
			gender.Code = "0"
		}

		ttsGenders = append(ttsGenders, *gender)
	}

	return ttsGenders
}


