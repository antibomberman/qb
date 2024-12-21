package dblayer

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Update обновляет записи используя структуру
func (qb *QueryBuilder) Update(data interface{}, fields ...string) error {
	var sets []string
	if len(fields) > 0 {
		// Обновляем только указанные поля
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	} else {
		// Обновляем все поля
		fields, _ := qb.getStructInfo(data)
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	}

	whereSQL := buildConditions(qb.conditions)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	_, err := qb.getExecutor().NamedExec(query, data)
	return err
}

// UpdateContext обновляет записи с контекстом
func (qb *QueryBuilder) UpdateContext(ctx context.Context, data interface{}, fields ...string) error {
	var sets []string
	if len(fields) > 0 {
		// Обновляем только указанные поля
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	} else {
		// Обновляем все поля
		fields, _ := qb.getStructInfo(data)
		for _, field := range fields {
			sets = append(sets, fmt.Sprintf("%s = :%s", field, field))
		}
	}

	whereSQL := buildConditions(qb.conditions)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	_, err := qb.getExecutor().NamedExecContext(ctx, query, data)
	return err
}

// UpdateMap обновляет записи используя map
func (qb *QueryBuilder) UpdateMap(data map[string]interface{}) error {
	sets := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		sets = append(sets, col+" = ?")
		values = append(values, val)
	}

	whereSQL := buildConditions(qb.conditions)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	return qb.execExec(query, values...)
}

func (qb *QueryBuilder) UpdateMapContext(ctx context.Context, data map[string]interface{}) error {
	sets := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		sets = append(sets, col+" = ?")
		values = append(values, val)
	}

	whereSQL := buildConditions(qb.conditions)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		qb.table,
		strings.Join(sets, ", "),
		whereSQL)

	return qb.execExecContext(ctx, query, values...)
}

// BulkUpdate выполняет массовое обновление записей
func (qb *QueryBuilder) BulkUpdate(records []map[string]interface{}, keyColumn string) error {
	if len(records) == 0 {
		return nil
	}

	// Получаем все колонки из первой записи
	columns := make([]string, 0)
	for column := range records[0] {
		if column != keyColumn {
			columns = append(columns, column)
		}
	}
	sort.Strings(columns)

	// Формируем CASE выражения для каждой колонки
	cases := make([]string, len(columns))
	keyValues := make([]interface{}, 0, len(records))
	valueArgs := make([]interface{}, 0, len(records)*len(columns))

	for i, column := range columns {
		whenClauses := make([]string, 0, len(records))
		for _, record := range records {
			if i == 0 {
				keyValues = append(keyValues, record[keyColumn])
			}
			whenClauses = append(whenClauses, "WHEN ? THEN ?")
			valueArgs = append(valueArgs, record[keyColumn], record[column])
		}
		cases[i] = fmt.Sprintf("%s = CASE %s %s END",
			column,
			keyColumn,
			strings.Join(whenClauses, " "),
		)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s IN (%s)",
		qb.table,
		strings.Join(cases, ", "),
		keyColumn,
		strings.Repeat("?,", len(records)-1)+"?",
	)

	// Объединяем все аргументы
	args := make([]interface{}, 0, len(valueArgs)+len(keyValues))
	args = append(args, valueArgs...)
	args = append(args, keyValues...)

	return qb.execExec(query, args...)
}

// BulkUpdateContext выполняет массовое обновление записей с контекстом
func (qb *QueryBuilder) BulkUpdateContext(ctx context.Context, records []map[string]interface{}, keyColumn string) error {
	if len(records) == 0 {
		return nil
	}

	// Получаем все колонки из первой записи
	columns := make([]string, 0)
	for column := range records[0] {
		if column != keyColumn {
			columns = append(columns, column)
		}
	}
	sort.Strings(columns)

	// Формируем CASE выражения для каждой колонки
	cases := make([]string, len(columns))
	keyValues := make([]interface{}, 0, len(records))
	valueArgs := make([]interface{}, 0, len(records)*len(columns))

	for i, column := range columns {
		whenClauses := make([]string, 0, len(records))
		for _, record := range records {
			if i == 0 {
				keyValues = append(keyValues, record[keyColumn])
			}
			whenClauses = append(whenClauses, "WHEN ? THEN ?")
			valueArgs = append(valueArgs, record[keyColumn], record[column])
		}
		cases[i] = fmt.Sprintf("%s = CASE %s %s END",
			column,
			keyColumn,
			strings.Join(whenClauses, " "),
		)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s IN (%s)",
		qb.table,
		strings.Join(cases, ", "),
		keyColumn,
		strings.Repeat("?,", len(records)-1)+"?",
	)

	// Объединяем все аргументы
	args := make([]interface{}, 0, len(valueArgs)+len(keyValues))
	args = append(args, valueArgs...)
	args = append(args, keyValues...)

	return qb.execExecContext(ctx, query, args...)
}

// BatchUpdate обновляет записи пакетами указанного размера
func (qb *QueryBuilder) BatchUpdate(records []map[string]interface{}, keyColumn string, batchSize int) error {
	if len(records) == 0 {
		return nil
	}

	// Разбиваем записи на пакеты
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		// Обновляем текущий пакет
		if err := qb.BulkUpdate(batch, keyColumn); err != nil {
			return err
		}
	}

	return nil
}

// BatchUpdateContext обновляет записи пакетами с поддержкой контекста
func (qb *QueryBuilder) BatchUpdateContext(ctx context.Context, records []map[string]interface{}, keyColumn string, batchSize int) error {
	if len(records) == 0 {
		return nil
	}

	// Разбиваем записи на пакеты
	for i := 0; i < len(records); i += batchSize {
		// Проверяем контекст
		if err := ctx.Err(); err != nil {
			return err
		}

		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}
		batch := records[i:end]

		// Обновляем текущий пакет
		if err := qb.BulkUpdateContext(ctx, batch, keyColumn); err != nil {
			return err
		}
	}

	return nil
}
