package dblayer

import "fmt"

type UniqueViolationError struct {
	Column string
}

func (e UniqueViolationError) Error() string {
	return fmt.Sprintf("unique constraint violation on column: %s", e.Column)
}
