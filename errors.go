package main

import (
    "net/http"
)

type PimErrId int
const (
	success PimErrId = iota 
	notFound
	emptyList
	badRequest
	undoEmpty
	authNoToken
	authSig
	authFail
	authToken
	authErr
	authTaken
	authBadEmail
	authBadPW
)

type PimError struct {
	Code     PimErrId    `json:"code"`     // error code
	Msg	     string      `json:"msg"`      // description of error
	Response int         `json:"response"` // http response code (if applicable)
}

func (e *PimError) Error() string { return e.Msg }

func (e *PimError) AppendMessage(additionalText string) {
	e.Msg += ": "
	e.Msg += additionalText
}


func pimErr(id PimErrId) PimError {
	return pimErrors[id]
}
func pimSuccess() PimError {
	return pimErrors[success]
}



type PimErrors []PimError


var pimErrors = PimErrors {
	PimError{ Code:success,     Msg:"pim: success",                    Response:http.StatusOK},
    PimError{ Code:notFound,    Msg:"pim: requested taskid not found", Response:http.StatusNotFound},
    PimError{ Code:emptyList,   Msg:"pim: empty task list",            Response:http.StatusNotFound},
    PimError{ Code:badRequest,  Msg:"pim: could not process request",  Response:http.StatusUnprocessableEntity},
    PimError{ Code:undoEmpty,   Msg:"pim: nothing to undo",  		   Response:http.StatusOK},
    PimError{ Code:authNoToken, Msg:"pim: no auth token",              Response:http.StatusUnauthorized},
    PimError{ Code:authSig,     Msg:"pim: invalid auth signature",     Response:http.StatusUnauthorized},
    PimError{ Code:authFail,    Msg:"pim: authentication failed",      Response:http.StatusUnauthorized},
    PimError{ Code:authToken,   Msg:"pim: invalid auth token",         Response:http.StatusUnauthorized},
    PimError{ Code:authErr,     Msg:"pim: authentication error",       Response:http.StatusInternalServerError},
    PimError{ Code:authTaken,   Msg:"pim: request username taken",     Response:http.StatusOK},
    PimError{ Code:authBadEmail,Msg:"pim: invalid email provided",     Response:http.StatusOK},
    PimError{ Code:authBadPW   ,Msg:"pim: insecure PW provided",       Response:http.StatusOK},    
}
