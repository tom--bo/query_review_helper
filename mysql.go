package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type IndexColumn struct {
	IndexName string `db:"index_name"`
	Columns   string `db:"columns"`
}

type IndexInfo struct {
	pkColumns string
	indexes   map[string]string
}

type TableColumn struct {
	TableName  string `db:"table_name"`
	ColumnName string `db:"column_name"`
}

type TableCardinality struct {
	Cardinality int `db:"cardinality"`
}

func connectMySQL() error {
	var err error
	mysqlHost := user + ":" + password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/" + database + "?loc=Local&parseTime=true"
	if socket != "" {
		mysqlHost = user + ":" + password + "@unix(" + socket + ")/" + database + "?loc=Local&parseTime=true"
	}

	db, err = sqlx.Open("mysql", mysqlHost)
	if err != nil {
		fmt.Println("Connection to MySQL fail.")
		return err
	}
	return nil
}

func getKeys(dbs, tbl string) (IndexInfo, error) {
	indexInfo := IndexInfo{}
	indexInfo.indexes = make(map[string]string)
	q := `
SELECT index_name as index_name, GROUP_CONCAT(column_name ORDER BY seq_in_index ASC) AS columns
FROM information_schema.statistics WHERE (table_schema, table_name) = (:sname, :tname) GROUP BY index_name ORDER BY columns
`
	rows, err := db.NamedQuery(q, map[string]interface{}{"sname": dbs, "tname": tbl})
	if err != nil {
		log.Fatalf("db.Query(): %s\n", err)
		return indexInfo, err
	}
	defer rows.Close()

	for rows.Next() {
		t := IndexColumn{}
		err = rows.StructScan(&t)
		if err != nil {
			fmt.Println(err.Error())
			return indexInfo, err
		}
		if t.IndexName == "PRIMARY" {
			indexInfo.pkColumns = t.Columns
		} else {
			indexInfo.indexes[t.Columns] = t.IndexName
		}
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("rows.Err(): %s\n", err)
		return indexInfo, err
	}
	return indexInfo, nil
}

func samplingColumnCardinality(table, pkColumn, column string, limit int) (int, error) {
	pkASC := strings.ReplaceAll(pkColumn, ",", " ASC, ") + " ASC"
	pkDESC := strings.ReplaceAll(pkColumn, ",", " DESC, ") + " DESC"
	if column == "" {
		fmt.Println(table, column)
		return 0, nil
	}

	q := fmt.Sprintf(`
SELECT COUNT(DISTINCT %s) as cardinality
FROM (
  (SELECT %s FROM %s ORDER BY %s LIMIT %d)
  UNION DISTINCT
  (SELECT %s FROM %s ORDER BY %s LIMIT %d)
) as tmp
`, column, column, table, pkASC, limit, column, table, pkDESC, limit)

	c := TableCardinality{}
	err := db.Get(&c, q)
	if err != nil {
		return 0, err
	}

	return c.Cardinality, nil
}

func assignOrphanColumns(tableMap map[string]string, tbls, cols []string) error {
	p := map[string]interface{}{
		"tablename":  tbls,
		"columnname": cols,
	}

	q := `SELECT table_name as table_name, column_name column_name
FROM information_schema.columns WHERE table_name IN (:tablename) AND column_name IN (:columnname)`
	query, args, err := sqlx.Named(q, p)
	if err != nil {
		return err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}
	query = db.Rebind(query)

	rows, err := db.Queryx(query, args...)
	if err != nil {
		return err
	}

	tableColumn := TableColumn{}
	for rows.Next() {
		err = rows.StructScan(&tableColumn)
		if err != nil {
			return err
		}
		if _, ok := tableMap[tableColumn.TableName]; ok {
			tableMap[tableColumn.TableName] += "," + tableColumn.ColumnName
		}
	}

	return nil
}

type ExplainResult struct {
	ID           int64          `db:"id"`
	SelectType   string         `db:"select_type"`
	Table        string         `db:"table"`
	Partitions   sql.NullString `db:"partitions"`
	Type         sql.NullString `db:"type"`
	PossibleKeys sql.NullString `db:"possible_keys"`
	Key          sql.NullString `db:"key"`
	KeyLen       sql.NullInt32  `db:"key_len"`
	Ref          sql.NullString `db:"ref"`
	Rows         int64          `db:"rows"`
	Filtered     float32        `db:"filtered"`
	Extra        sql.NullString `db:"Extra"`
}

func makeExplainQuery(query string) string {
	query = strings.TrimSpace(query)
	if len(query) < 7 {
		return "EXPLAIN " + query
	}
	if strings.ToUpper(query[0:7]) != "EXPLAIN" {
		return "EXPLAIN " + query
	}
	return query
}

func getExplainResult(query string) ([]ExplainResult, error) {
	query = makeExplainQuery(query)
	ret := []ExplainResult{}

	rows, err := db.Queryx(query)
	r := ExplainResult{}
	for rows.Next() {
		err = rows.StructScan(&r)
		if err != nil {
			return ret, err
		}
		ret = append(ret, r)
	}

	return ret, nil
}
