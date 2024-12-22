package dblayer

import (
	"fmt"
	"strings"
)

// FullTextSearch добавляет поддержку полнотекстового поиска
type FullTextSearch struct {
	SearchRank float64 `db:"search_rank"`
}

// Search выполняет полнотекстовый поиск
func (qb *QueryBuilder) Search(columns []string, query string) *QueryBuilder {
	if qb.getDriverName() == "postgres" {
		// Для PostgreSQL используем ts_vector и ts_query
		vectorExpr := make([]string, len(columns))
		for i, col := range columns {
			vectorExpr[i] = fmt.Sprintf("to_tsvector(%s)", col)
		}

		qb.columns = append(qb.columns,
			fmt.Sprintf("ts_rank_cd(to_tsvector(concat_ws(' ', %s)), plainto_tsquery(?)) as search_rank",
				strings.Join(columns, ", ")))

		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf("to_tsvector(concat_ws(' ', %s)) @@ plainto_tsquery(?)",
				strings.Join(columns, ", ")),
			args: []interface{}{query},
		})

		qb.OrderBy("search_rank", "DESC")
	} else {
		// Для MySQL используем MATCH AGAINST
		qb.columns = append(qb.columns,
			fmt.Sprintf("MATCH(%s) AGAINST(? IN BOOLEAN MODE) as search_rank",
				strings.Join(columns, ",")))

		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf("MATCH(%s) AGAINST(? IN BOOLEAN MODE)",
				strings.Join(columns, ",")),
			args: []interface{}{query},
		})

		qb.OrderBy("search_rank", "DESC")
	}

	return qb
}
