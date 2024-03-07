package model

import "errors"

var (
	ErrBlankName     = errors.New("name is blank")
	ErrWrongPriority = errors.New("priority is less than 0")
	ErrGoodNotFound  = errors.New("good not found")
)
