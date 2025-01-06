package migrate

import (
	"time"
)

type Migrations struct {
	ID      int64  `db:"id"`
	Name    string `db:"name"`
	Up      string `db:"up"`
	Down    string `db:"down"`
	Status  string `db:"status"`
	Version int    `db:"version"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
