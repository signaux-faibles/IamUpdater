package main

import (
	"fmt"
)

//type MisconfiguredUserError struct {
//	user  User
//	cause string
//}
//
//func (e MisconfiguredUserError) Error() string {
//	return fmt.Sprintf("l'utilisateur %s est mal configuré : %s", e.user.email, e.cause)
//}

type PipelineRunError struct {
	stage PipelineStage
	err   error
}

func (e PipelineRunError) Error() string {
	return fmt.Sprintf("Un étape du pipeline a rencontré une erreur: %s", e.stage.id)
}

func (e PipelineRunError) Unwrap() error {
	return e.err
}

type InvalidExcelFileError struct {
	msg string
	err error
}

func (e InvalidExcelFileError) Error() string {
	return e.msg
}

func (e InvalidExcelFileError) Unwrap() error {
	return e.err
}
