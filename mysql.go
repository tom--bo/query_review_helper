package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
)

func connectMySQL() {
	var err error
	mysqlHost := user + ":" + password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/" + database + "?loc=Local&parseTime=true"
	if socket != "" {
		mysqlHost = user + ":" + password + "@unix(" + socket + ")/" + database + "?loc=Local&parseTime=true"
	}

	db, err = sqlx.Open("mysql", mysqlHost)
	if err != nil {
		fmt.Println("Connection to MySQL fail.")
	}
}


type TableInfo struct {
	IndexName string `db:"INDEX_NAME"`
	Columns string `db:"columns"`
}

func getKeys(dbs, tbl string) {
	q := `
SELECT index_name, GROUP_CONCAT(column_name ORDER BY seq_in_index ASC) AS columns
FROM information_schema.statistics WHERE (table_schema, table_name) = (:sname, :tname) GROUP BY index_name
`
	rows, err := db.NamedQuery(q, map[string]interface{}{"sname": dbs, "tname": tbl})
	if err != nil {
		log.Fatalf("db.Query(): %s\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		t := TableInfo{}
		err = rows.StructScan(&t)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		fmt.Printf("%v\n", t)
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("rows.Err(): %s\n", err)
	}
}


