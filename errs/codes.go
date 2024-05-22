package errs

type Code struct{
	Code string
	Info string
}

var AB_ERR_001 Code = Code{Code: "AB_ERR_001", Info: "audiobook not found"}
var AB_ERR_002 Code = Code{Code: "AB_ERR_002", Info: "tag query path parameter not interger"}
var AB_ERR_003 Code = Code{Code: "AB_ERR_003", Info: "page query path parameter not interger"}
var AB_ERR_004 Code = Code{Code: "AB_ERR_004", Info: "itemsPerPage query path parameter not interger"}
var AB_ERR_005 Code = Code{Code: "AB_ERR_005", Info: "can't parse UUID from path"}
var AB_ERR_006 Code = Code{Code: "AB_ERR_006", Info: "convertion query path parameter not boolean"}
var AB_ERR_007 Code = Code{Code: "AB_ERR_007", Info: "can't connect to minio service"}
var AB_ERR_008 Code = Code{Code: "AB_ERR_008", Info: "error checking or creating bucket in minio service"}
var AB_ERR_009 Code = Code{Code: "AB_ERR_009", Info: "file bigger than 10 MB"}
var AB_ERR_010 Code = Code{Code: "AB_ERR_010", Info: "can't get info from the file"}
var AB_ERR_011 Code = Code{Code: "AB_ERR_011", Info: "audiobook name already exist"}
var AB_ERR_012 Code = Code{Code: "AB_ERR_012", Info: "external api: minio api error"}
var AB_ERR_013 Code = Code{Code: "AB_ERR_013", Info: "account not authorized for audiobook"}
var AB_ERR_014 Code = Code{Code: "AB_ERR_014", Info: "can't delete, audiobook in use"}
var AB_ERR_015 Code = Code{Code: "AB_ERR_015", Info: "genders query path parameter not boolean"}
var AB_ERR_016 Code = Code{Code: "AB_ERR_016", Info: "langCode query path parameter empty"}

var ERR_DB Code = Code{Code: "ERR_DB", Info: "database error"}
var ERR_ENCODEJSON Code = Code{Code: "ERR_ENCODEJSON", Info: "error encoding json response"}
var ERR_DECODEJSON Code = Code{Code: "ERR_ENCODEJSON", Info: "error decoding json from request"}
var ERR_UNANTICIPATED Code = Code{Code: "ERR_UNANTICIPATED", Info: "unexpected error - contact support"}
