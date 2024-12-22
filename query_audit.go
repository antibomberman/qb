package dblayer

import (
	"encoding/json"
	"reflect"
	"time"
)

// AuditLog представляет запись аудита
type AuditLog struct {
	ID        int64     `db:"id"`
	TableName string    `db:"table_name"`
	RecordID  int64     `db:"record_id"`
	Action    string    `db:"action"`
	OldData   []byte    `db:"old_data"`
	NewData   []byte    `db:"new_data"`
	UserID    int64     `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

// WithAudit включает аудит для запроса
func (qb *QueryBuilder) WithAudit(userID int64) *QueryBuilder {
	qb.On(BeforeUpdate, func(data interface{}) error {
		oldData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		_, err = qb.Create(&AuditLog{
			TableName: qb.table,
			RecordID:  reflect.ValueOf(data).Elem().FieldByName("ID").Int(),
			Action:    "update",
			OldData:   oldData,
			UserID:    userID,
			CreatedAt: time.Now(),
		})
		return err
	})

	qb.On(AfterUpdate, func(data interface{}) error {
		newData, err := json.Marshal(data)
		if err != nil {
			return err
		}

		return qb.Where("table_name = ? AND record_id = ?", qb.table,
			reflect.ValueOf(data).Elem().FieldByName("ID").Int()).
			OrderBy("id", "DESC").
			Limit(1).
			UpdateMap(map[string]interface{}{
				"new_data": newData,
			})
	})

	return qb
}
