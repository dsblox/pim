package main

import (
)


type PimError struct {
	Code int    `json:"code"` // error code
	Msg	 string `json:"msg"`  // description of error
}

func (e *PimError) Error() string { return e.Msg }

type PimErrors []PimError

var pimErrors = PimErrors {
	PimError{ Code:0, Msg:"pim: success"},
    PimError{ Code:1, Msg:"pim: requested taskid not found"},
    PimError{ Code:2, Msg:"pim: empty task list"},
}
