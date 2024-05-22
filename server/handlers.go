package server

import (
	"fmt"

	"net/http"
	"strconv"
	"encoding/json"
	"os"

	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"

	"golang-rest-api/errs"
	"golang-rest-api/models"

)

//Request Header Claims
const (
	claim      = "teste"
	Account            = "serviceaccount"
	HeaderAccount      = claim + Account
	USERID         = "userid"
	HeaderUserID   = claim + USERID
	USERNAME       = "username"
	HeaderUserName = claim + USERNAME
)

func getHeaderParams(r *http.Request ) map[string]string{
	params := map[string]string{}	//or make(map[string]string)

	params[HeaderAccount] = r.Header.Get(HeaderAccount)
	params[HeaderUserID] = r.Header.Get(HeaderUserID)
	params[HeaderUserName] = r.Header.Get(HeaderUserName)

	return params
}
//reads integers from the request header parameter, if it doesnt exist puts the v input value
func readInt(r *http.Request, param string, v int) (int, error) {
	p := r.FormValue(param)
	if p == "" {
		return v, nil
	}

	return strconv.Atoi(p)
}
//////////////////////////////////////////////////////////////////////////////////
func (s *Server) getAudiobooks(w http.ResponseWriter, r *http.Request) {
	var err error
	logger := *hlog.FromRequest(r)

	filters := make(map[string]interface{})

	filters["headerParams"] = getHeaderParams(r)

	filters["name"] = r.FormValue("name")
	filters["sortBy"] = r.FormValue("sortBy")
	filters["orderBy"] = r.FormValue("orderBy")
	filters["date"]		= r.FormValue("date")

	filters["tag"], err = readInt(r, "tag", 0)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_002, err))
		return
	}

	filters["page"], err = readInt(r, "page", 0)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_003, err))
		return
	}
	filters["itemsPerPage"], err = readInt(r, "itemsPerPage", 0)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_004, err))
		return
	}

	audiobooksDB, totalRecords, page, itemsPerPage, err := s.Repositories.AudiobookRepository.GetAudiobooks(filters)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	records := NewAudiobooksResponse(&audiobooksDB)

	result := &PagedResults{Page: page, ItemsPerPage: itemsPerPage, Records: records, TotalRecords: totalRecords}

	if err = json.NewEncoder(w).Encode(result); err != nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.ERR_ENCODEJSON, err))
		return
	}	
}

func (s *Server) getAudiobookByID(w http.ResponseWriter, r *http.Request){
	logger := *hlog.FromRequest(r)
	param := mux.Vars(r)["id"]

	//id in UUID format
	id, err := uuid.Parse(param)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_005, err))
		return
	}

	audio, err := s.Repositories.AudiobookRepository.GetAudiobook(id, true)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	audioResponse := NewAudiobookResponse(&audio)

	if err = json.NewEncoder(w).Encode(audioResponse); err != nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.ERR_ENCODEJSON, err))
		return
	}
}

func (s *Server) postAudiobook(w http.ResponseWriter, r *http.Request){
	var err error
	logger := *hlog.FromRequest(r)

	convertion:= true

	headerParams := getHeaderParams(r)
	account := headerParams[HeaderAccount]

	convertionString :=r.FormValue("convertion")
	if convertionString !=""{
		convertion, err = strconv.ParseBool(convertionString)
		if err != nil {
			errs.HTTPErrorResponse(w, logger, errs.E( errs.AB_ERR_006, err))
			return
		}
	}
	
	//generate UUID for fileName
	id := uuid.New()

	// Maximum upload of 10 MB files
	err = r.ParseMultipartForm(10 << 20)
	if err !=nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_009, err))
		return
	}

	file, handler, err := r.FormFile("objectName")
	if err !=nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_010, err))
		return
	}
	defer file.Close()
	logger.Info().Msg("Uploaded File:" + handler.Filename + " File Size:" + fmt.Sprintf("%#v", handler.Size) + "MIME Header: " + fmt.Sprintf("%#v", handler.Header))
	
	fileName := handler.Filename

	//get name for db data insertion
	name := r.FormValue("name")
	if name ==""{
		name = fileName
		if convertion{
			name = strings.TrimSuffix(fileName, filepath.Ext(fileName))+"."+os.Getenv("AUDIO_FORMAT")
		}
	}

	//add info to DB
	audio := models.Audiobook{
		ID: id,
		Name: name,
		Description: r.FormValue("description"),
		AccountID: account,
	}

	if err = s.Repositories.AudiobookRepository.CreateAudiobook(&audio); err != nil{
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	audiobookResponse := NewAudiobookResponse(&audio)
	//Change statusCode to 201???
	if err = json.NewEncoder(w).Encode(audiobookResponse); err != nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.ERR_ENCODEJSON, err))
		return
	}	
}

func (s *Server) editAudiobook(w http.ResponseWriter, r *http.Request){
	logger := *hlog.FromRequest(r)

	var audioJson AudiobookJSON
	var err error

	param := mux.Vars(r)["id"]

	id, err := uuid.Parse(param)
	if err != nil {

		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_005, err))
		return
	}

	err = json.NewDecoder(r.Body).Decode(&audioJson)
	defer r.Body.Close()
	if err = decoderErr(err); err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	//TODO: make body validation

	//////////////

	audioBD, err := s.Repositories.AudiobookRepository.GetAudiobook(id, false)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	if audioBD.Name != ""{
		audioBD.Name = audioJson.Name
	}
	if audioJson.Description != ""{
		audioBD.Description = audioJson.Description
	}

	for i := range audioJson.Tag{
		tag := models.Tag{
			ID: audioJson.Tag[i].ID,
		}
		audioBD.Tag = append(audioBD.Tag, &tag)
	}

	if err = s.Repositories.AudiobookRepository.EditAudiobook(&audioBD); err != nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.ERR_ENCODEJSON, err))
		return
	}

}


func (s *Server) deleteAudiobookByID(w http.ResponseWriter, r *http.Request){
	var err error
	
	logger := *hlog.FromRequest(r)
	
	param := mux.Vars(r)["id"]

	id, err := uuid.Parse(param)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_005, err))
		return
	}

	if err = s.Repositories.AudiobookRepository.DeleteAudiobook(id); err != nil{
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

}

func (s *Server) getTags(w http.ResponseWriter, r *http.Request){
	logger := *hlog.FromRequest(r)

	tags, err :=s.Repositories.TagRepository.GetTags()
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	result := NewTagResponse(&tags)

	if err = json.NewEncoder(w).Encode(result); err != nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.ERR_ENCODEJSON, err))
		return
	}
}


func (s *Server) getLanguages(w http.ResponseWriter, r *http.Request){
	var err error
	
	logger := *hlog.FromRequest(r)

	gender := false 
	genderFilter := r.FormValue("genders")

	if genderFilter !=""{
		gender, err =strconv.ParseBool(genderFilter)
		if err != nil {
			errs.HTTPErrorResponse(w, logger, errs.E(errs.AB_ERR_015, err))
			return
		}
	}

	languages, err :=s.Repositories.LanguageRepository.GetTTSLanguages(gender)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}
	result:= &RecordsJSON{Records: NewLanguagesResponse(&languages)}

	if err = json.NewEncoder(w).Encode(result); err != nil{
		errs.HTTPErrorResponse(w, logger, errs.E(errs.ERR_ENCODEJSON, err))
		return
	}
}