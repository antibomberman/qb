package dblayer

import (
	"fmt"
	"strings"
	"time"
)

// getDateFunctions возвращает функции для текущей СУБД
func (qb *QueryBuilder) getDateFunctions() DateFunctions {
	if qb.getDriverName() == "postgres" {
		return DateFunctions{
			DateDiff:    "DATE_PART('day', %s::timestamp - %s::timestamp)",
			DateTrunc:   "DATE_TRUNC",
			DateFormat:  "TO_CHAR",
			TimeZone:    "AT TIME ZONE",
			Extract:     "EXTRACT",
			DateAdd:     "% + INTERVAL '% %'",
			CurrentDate: "CURRENT_DATE",
		}
	}
	return DateFunctions{
		DateDiff:    "DATEDIFF(%s, %s)",
		DateTrunc:   "DATE_FORMAT", // MySQL не имеет прямого аналога DATE_TRUNC
		DateFormat:  "DATE_FORMAT",
		TimeZone:    "CONVERT_TZ",
		Extract:     "EXTRACT",
		DateAdd:     "DATE_ADD(%, INTERVAL % %)",
		CurrentDate: "CURDATE()",
	}
}

// convertToPostgresFormat преобразует формат даты из MySQL в PostgreSQL
func convertToPostgresFormat(mysqlFormat string) string {
	replacer := strings.NewReplacer(
		"%Y", "YYYY",
		"%m", "MM",
		"%d", "DD",
		"%H", "HH24",
		"%i", "MI",
		"%s", "SS",
	)
	return replacer.Replace(mysqlFormat)
}

// getMySQLDateFormat преобразует части даты в формат MySQL
func getMySQLDateFormat(part string) string {
	switch strings.ToLower(part) {
	case "year":
		return "%Y"
	case "month":
		return "%Y-%m"
	case "day":
		return "%Y-%m-%d"
	case "hour":
		return "%Y-%m-%d %H"
	case "minute":
		return "%Y-%m-%d %H:%i"
	default:
		return "%Y-%m-%d %H:%i:%s"
	}
}

// WhereDate добавляет условие по дате
func (qb *QueryBuilder) WhereDate(column string, operator string, value time.Time) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) %s ?", column, operator),
		args:     []interface{}{value},
	})
	return qb
}

// WhereBetweenDates добавляет условие между датами
func (qb *QueryBuilder) WhereBetweenDates(column string, start time.Time, end time.Time) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) BETWEEN ? AND ?", column),
		args:     []interface{}{start.Format("2006-01-02"), end.Format("2006-01-02")},
	})
	return qb
}

// WhereDateTime добавляет условие по дате и времени
func (qb *QueryBuilder) WhereDateTime(column string, operator string, value time.Time) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s %s ?", column, operator),
		args:     []interface{}{value.Format("2006-01-02 15:04:05")},
	})
	return qb
}

// WhereBetweenDateTime добавляет условие между датами и временем
func (qb *QueryBuilder) WhereBetweenDateTime(column string, start time.Time, end time.Time) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("%s BETWEEN ? AND ?", column),
		args: []interface{}{
			start.Format("2006-01-02 15:04:05"),
			end.Format("2006-01-02 15:04:05"),
		},
	})
	return qb
}

// WhereYear добавляет условие по году
func (qb *QueryBuilder) WhereYear(column string, operator string, year int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(YEAR FROM %s) %s ?", column, operator),
		args:     []interface{}{year},
	})
	return qb
}

// WhereMonth добавляет условие по месяцу
func (qb *QueryBuilder) WhereMonth(column string, operator string, month int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(MONTH FROM %s) %s ?", column, operator),
		args:     []interface{}{month},
	})
	return qb
}

// WhereDay добавляет условие по дню
func (qb *QueryBuilder) WhereDay(column string, operator string, day int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(DAY FROM %s) %s ?", column, operator),
		args:     []interface{}{day},
	})
	return qb
}

// WhereTime добавляет условие по времени (без учета даты)
func (qb *QueryBuilder) WhereTime(column string, operator string, value time.Time) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("TIME(%s) %s ?", column, operator),
		args:     []interface{}{value.Format("15:04:05")},
	})
	return qb
}

// WhereDateIsNull проверяет является ли дата NULL
func (qb *QueryBuilder) WhereDateIsNull(column string) *QueryBuilder {
	return qb.WhereNull(column)
}

// WhereDateIsNotNull проверяет является ли дата NOT NULL
func (qb *QueryBuilder) WhereDateIsNotNull(column string) *QueryBuilder {
	return qb.WhereNotNull(column)
}

// WhereCurrentDate добавляет условие на текущую дату
func (qb *QueryBuilder) WhereCurrentDate(column string, operator string) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) %s CURRENT_DATE", column, operator),
	})
	return qb
}

// WhereLastDays добавляет условие за последние n дней
func (qb *QueryBuilder) WhereLastDays(column string, days int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) >= CURRENT_DATE - INTERVAL ? DAY", column),
		args:     []interface{}{days},
	})
	return qb
}

// WhereWeekday добавляет условие по дню недели (1 = Понедельник, 7 = Воскресенье)
func (qb *QueryBuilder) WhereWeekday(column string, operator string, weekday int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(DOW FROM %s) %s ?", column, operator),
		args:     []interface{}{weekday},
	})
	return qb
}

// WhereQuarter добавляет условие по кварталу (1-4)
func (qb *QueryBuilder) WhereQuarter(column string, operator string, quarter int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(QUARTER FROM %s) %s ?", column, operator),
		args:     []interface{}{quarter},
	})
	return qb
}

// WhereWeek добавляет условие по номеру недели в году (1-53)
func (qb *QueryBuilder) WhereWeek(column string, operator string, week int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(WEEK FROM %s) %s ?", column, operator),
		args:     []interface{}{week},
	})
	return qb
}

// WhereDateRange добавляет условие по диапазону дат с включением/исключением границ
func (qb *QueryBuilder) WhereDateRange(column string, start time.Time, end time.Time, inclusive bool) *QueryBuilder {
	if inclusive {
		return qb.WhereBetweenDates(column, start, end)
	}

	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) > ? AND DATE(%s) < ?", column, column),
		args:     []interface{}{start.Format("2006-01-02"), end.Format("2006-01-02")},
	})
	return qb
}

// WhereNextDays добавляет условие на следующие n дней
func (qb *QueryBuilder) WhereNextDays(column string, days int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) <= CURRENT_DATE + INTERVAL ? DAY AND DATE(%s) >= CURRENT_DATE", column, column),
		args:     []interface{}{days},
	})
	return qb
}

// WhereDateBetweenColumns проверяет, что дата находится между значениями двух других колонок
func (qb *QueryBuilder) WhereDateBetweenColumns(dateColumn string, startColumn string, endColumn string) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("DATE(%s) BETWEEN DATE(%s) AND DATE(%s)", dateColumn, startColumn, endColumn),
	})
	return qb
}

// WhereAge добавляет условие по возрасту (для дат рождения)
func (qb *QueryBuilder) WhereAge(column string, operator string, age int) *QueryBuilder {
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf("EXTRACT(YEAR FROM AGE(%s)) %s ?", column, operator),
		args:     []interface{}{age},
	})
	return qb
}

// WhereDateDiff добавляет условие по разнице между датами
func (qb *QueryBuilder) WhereDateDiff(column1 string, column2 string, operator string, days int) *QueryBuilder {
	df := qb.getDateFunctions()
	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   fmt.Sprintf(df.DateDiff+" %s ?", column1, column2, operator),
		args:     []interface{}{days},
	})
	return qb
}

// WhereDateTrunc добавляет условие с усечением даты
func (qb *QueryBuilder) WhereDateTrunc(part string, column string, operator string, value time.Time) *QueryBuilder {
	df := qb.getDateFunctions()
	var clause string
	var args []interface{}

	if qb.getDriverName() == "postgres" {
		clause = fmt.Sprintf("%s(?, %s) %s ?", df.DateTrunc, column, operator)
		args = []interface{}{part, value}
	} else {
		// Преобразуем part в формат MySQL
		format := getMySQLDateFormat(part)
		clause = fmt.Sprintf("%s(%s, ?) %s ?", df.DateTrunc, column, operator)
		args = []interface{}{format, value.Format(format)}
	}

	qb.conditions = append(qb.conditions, Condition{
		operator: "AND",
		clause:   clause,
		args:     args,
	})
	return qb
}

// WhereTimeWindow добавляет условие попадания времени в окно
func (qb *QueryBuilder) WhereTimeWindow(column string, startTime, endTime time.Time) *QueryBuilder {
	if qb.getDriverName() == "postgres" {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("EXTRACT(HOUR FROM %s) * 60 + EXTRACT(MINUTE FROM %s) BETWEEN ? AND ?", column, column),
			args: []interface{}{
				startTime.Hour()*60 + startTime.Minute(),
				endTime.Hour()*60 + endTime.Minute(),
			},
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("TIME(%s) BETWEEN ? AND ?", column),
			args: []interface{}{
				startTime.Format("15:04:05"),
				endTime.Format("15:04:05"),
			},
		})
	}
	return qb
}

// WhereBusinessDays добавляет условие только по рабочим дням
func (qb *QueryBuilder) WhereBusinessDays(column string) *QueryBuilder {
	if qb.getDriverName() == "postgres" {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("EXTRACT(DOW FROM %s) BETWEEN 1 AND 5", column),
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("WEEKDAY(%s) < 5", column),
		})
	}
	return qb
}

// WhereDateFormat добавляет условие по отформатированной дате
func (qb *QueryBuilder) WhereDateFormat(column string, format string, operator string, value string) *QueryBuilder {
	df := qb.getDateFunctions()

	if qb.getDriverName() == "postgres" {
		// Преобразуем формат из MySQL в PostgreSQL
		pgFormat := convertToPostgresFormat(format)
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s(%s, ?) %s ?", df.DateFormat, column, operator),
			args:     []interface{}{pgFormat, value},
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s(%s, ?) %s ?", df.DateFormat, column, operator),
			args:     []interface{}{format, value},
		})
	}
	return qb
}

// WhereTimeZone добавляет условие с учетом временной зоны
func (qb *QueryBuilder) WhereTimeZone(column string, operator string, value time.Time, timezone string) *QueryBuilder {
	df := qb.getDateFunctions()

	if qb.getDriverName() == "postgres" {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s %s ? %s ?", column, df.TimeZone, operator),
			args:     []interface{}{timezone, value},
		})
	} else {
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause:   fmt.Sprintf("%s(%s, 'UTC', ?) %s ?", df.TimeZone, column, operator),
			args:     []interface{}{timezone, value},
		})
	}
	return qb
}
