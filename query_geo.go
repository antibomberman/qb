package dblayer

import "fmt"

// Point представляет географическую точку
type Point struct {
	Lat float64
	Lng float64
}

// GeoSearch добавляет геопространственные запросы
func (qb *QueryBuilder) GeoSearch(column string, point Point, radius float64) *QueryBuilder {
	if qb.getDriverName() == "postgres" {
		// Для PostgreSQL с PostGIS
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf(
				"ST_DWithin(ST_SetSRID(ST_MakePoint(%s), 4326), ST_SetSRID(ST_MakePoint(?, ?), 4326), ?)",
				column),
			args: []interface{}{point.Lng, point.Lat, radius},
		})
	} else {
		// Для MySQL
		qb.conditions = append(qb.conditions, Condition{
			operator: "AND",
			clause: fmt.Sprintf(
				"ST_Distance_Sphere(%s, POINT(?, ?)) <= ?",
				column),
			args: []interface{}{point.Lng, point.Lat, radius},
		})
	}
	return qb
}
