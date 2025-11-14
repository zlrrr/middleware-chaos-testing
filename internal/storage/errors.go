package storage

import "errors"

var (
	// ErrTaskNotFound is returned when a task is not found
	ErrTaskNotFound = errors.New("task not found")

	// ErrResultNotFound is returned when a result is not found
	ErrResultNotFound = errors.New("result not found")

	// ErrTaskAlreadyExists is returned when attempting to create a duplicate task
	ErrTaskAlreadyExists = errors.New("task already exists")
)
