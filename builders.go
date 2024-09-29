package dblayer

import (
	"context"
	"fmt"
	"strings"
)

func (r *DBLayer) buildGroupByQuery(tableName string, groupColumns []string, aggregations map[string]string, conditions []Condition) (string, []interface{}) {
	selectClauses := append([]string{}, groupColumns...)
	for alias, agg := range aggregations {
		selectClauses = append(selectClauses, fmt.Sprintf("%s AS %s", agg, alias))
	}

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectClauses, ", "), tableName)
	where, args := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
	}
	query += " GROUP BY " + strings.Join(groupColumns, ", ")

	return query, args
}

func (r *DBLayer) buildUpdateQuery(tableName string, updates map[string]interface{}, conditions []Condition) (string, []interface{}) {
	setStatements := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+len(conditions))

	for field, value := range updates {
		setStatements = append(setStatements, field+" = ?")
		args = append(args, value)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setStatements, ", "))
	where, whereArgs := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
		args = append(args, whereArgs...)
	}

	return query, args
}

func (r *DBLayer) buildDeleteQuery(tableName string, conditions []Condition) (string, []interface{}) {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	where, args := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
	}
	return query, args
}

func (r *DBLayer) buildWhereClause(conditions []Condition) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	clauses := make([]string, 0, len(conditions))
	args := make([]interface{}, 0, len(conditions))

	for _, condition := range conditions {
		clauses = append(clauses, fmt.Sprintf("%s %s ?", condition.Column, condition.Operator))
		args = append(args, condition.Value)
	}

	return strings.Join(clauses, " AND "), args
}

func (r *DBLayer) buildExistsQuery(tableName string, conditions []Condition) (string, []interface{}) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s", tableName)
	where, args := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
	}
	query += ")"
	return query, args
}
func (r *DBLayer) buildSelectQuery(tableName string, conditions []Condition) (string, []interface{}) {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	where, args := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
	}
	return query, args
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (r *DBLayer) aggregateValue(ctx context.Context, aggregateFunc, tableName, column string, conditions []Condition) (interface{}, error) {
	query, args := r.buildAggregateQuery(fmt.Sprintf("%s(%s)", aggregateFunc, column), tableName, conditions)
	var result interface{}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate %s of %s in %s: %w", aggregateFunc, column, tableName, err)
	}
	return result, nil
}

func (r *DBLayer) buildAggregateQuery(aggregation, tableName string, conditions []Condition) (string, []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s", aggregation, tableName)
	where, args := r.buildWhereClause(conditions)
	if where != "" {
		query += " WHERE " + where
	}
	return query, args
}
