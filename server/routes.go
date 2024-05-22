package server

import (
	"net/http"

)

const (
	// Content Type header key
	contentTypeHeaderKey string = "Content-Type"
	// application/json header value for Content-Type header key
	appJSONContentTypeHeaderVal string = "application/json"
	// id is used to represent an external id. This is a common
	// enough pattern in these services
	idPathDir string = "/{id}"
	//audio V1 Path root
	audiobookV1PathRoot string = "/v1/audiobook"
	//tags V1 Path root
	tagsV1PathRoot string = "/v1/tags"
	//languages path dir
	languagesPathDir string = "/languages"
)

func (app *Server)registerRoutes() {
	//
	app.router.Handle(audiobookV1PathRoot,
		app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.getAudiobooks)).Methods(http.MethodGet)

	app.router.Handle(audiobookV1PathRoot,
		app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.postAudiobook)).Methods(http.MethodPost)

	app.router.Handle(audiobookV1PathRoot+idPathDir,
		app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.getAudiobookByID)).Methods(http.MethodGet)

	app.router.Handle(audiobookV1PathRoot+idPathDir,
		app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.deleteAudiobookByID)).Methods(http.MethodDelete)

	app.router.Handle(audiobookV1PathRoot+idPathDir,
		app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.editAudiobook)).Methods(http.MethodPut).Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	app.router.Handle(tagsV1PathRoot,
		app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.getTags)).Methods(http.MethodGet)

	// app.router.Handle(ttsV1PathRoot+languagesPathDir,
	// 	app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.getLanguages)).Methods(http.MethodGet)
		
	//app.router.Handle(ttsV1PathRoot+languagesPathDir+"/{langCode}/genders",
	//	app.LoggerChain().Append(app.JsonContentTypeResponseHandler).ThenFunc(app.getTTSGendersByLangCode)).Methods(http.MethodGet)

}
