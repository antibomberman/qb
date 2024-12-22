package dblayer

import (
	"encoding/json"
	"time"
)

// QueuedOperation представляет отложенную операцию
type QueuedOperation struct {
	ID        int64     `db:"id"`
	Operation string    `db:"operation"`
	Data      []byte    `db:"data"`
	Status    string    `db:"status"`
	RunAt     time.Time `db:"run_at"`
}

// Queue добавляет операцию в очередь
func (qb *QueryBuilder) Queue(operation string, data interface{}, runAt time.Time) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = qb.Create(&QueuedOperation{
		Operation: operation,
		Data:      jsonData,
		Status:    "pending",
		RunAt:     runAt,
	})
	return err
}

// ProcessQueue обрабатывает очередь
func (qb *QueryBuilder) ProcessQueue(handler func(QueuedOperation) error) error {
	var operations []QueuedOperation

	_, err := qb.Where("status = ? AND run_at <= ?", "pending", time.Now()).
		Get(&operations)
	if err != nil {
		return err
	}

	for _, op := range operations {
		if err := handler(op); err != nil {
			return err
		}

		err = qb.Where("id = ?", op.ID).
			UpdateMap(map[string]interface{}{
				"status": "completed",
			})
		if err != nil {
			return err
		}
	}

	return nil
}
