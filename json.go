package dblayer

import (
	"context"
	"encoding/json"
	"fmt"
)

func (r *DBLayer) SaveJSON(ctx context.Context, tableName, idColumn string, id interface{}, jsonColumn string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", tableName, jsonColumn, idColumn)
	_, err = r.db.ExecContext(ctx, query, jsonData, id)
	if err != nil {
		return fmt.Errorf("failed to save JSON data: %w", err)
	}
	return nil
}

func (r *DBLayer) LoadJSON(ctx context.Context, tableName, idColumn string, id interface{}, jsonColumn string, result interface{}) error {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", jsonColumn, tableName, idColumn)
	var jsonData []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(&jsonData)
	if err != nil {
		return fmt.Errorf("failed to load JSON data: %w", err)
	}

	err = json.Unmarshal(jsonData, result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}
