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


type PimErrors []PimError


var pimErrors = PimErrors {
	 PimError{ Code:success,    Msg:"pim: success",                    Response:http.StatusOK},
    PimError{ Code:notFound,   Msg:"pim: requested taskid not found", Response:http.StatusNotFound},
    PimError{ Code:emptyList,  Msg:"pim: empty task list",            Response:http.StatusNotFound},
    PimError{ Code:badRequest, Msg:"pim: could not process request",  Response:http.StatusUnprocessableEntity},
}
