package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"strings"
)

func start(q string) {
	// parse
	tables, columns, err := getColumnsFromSQL(q)
	if err != nil {
		fmt.Println("parse select-stmt err")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// build tableMap
	tableMap := make(map[string]string)
	orphanColumns := []string{}

	// fmt.Println("== Tables")
	for _, tbl := range tables {
		tableMap[tbl] = ""
		// fmt.Println("- " + tbl)
	}
	for _, c := range columns {
		col := strings.Split(c, ".")
		if len(col) < 2 {
			fmt.Println("Something bad happen")
			os.Exit(1)
		}

		if col[0] == "" {
			orphanColumns = append(orphanColumns, col[1])
		} else {
			tableMap[col[0]] += "," + col[1]
			if tableMap[col[0]][0:1] == "," {
				tableMap[col[0]] = tableMap[col[0]][1:]
			}
		}
	}
	// assign orphan columns
	err = assignOrphanColumns(tableMap, tables, orphanColumns)
	if err != nil {
		fmt.Println("assignOrphanColumn err", err.Error())
		os.Exit(1)
	}

	// fmt.Println("= Tables")
	for _, tbl := range tables {
		fmt.Println("\n- " + tbl)

		// Index
		fmt.Println("  - Index")
		info, err := getKeys(database, tbl)
		if err != nil {
			fmt.Println("Get keys err", err.Error())
			os.Exit(1)
		}
		fmt.Printf("    - %-20s: (%s)\n", "PRIMARY", info.pkColumns)
		for k, v := range info.indexes {
			fmt.Printf("    - %-20s: (%s)\n", v, k)
		}

		// Column Cardinality
		fmt.Println("  - Cardinality")

		// tableColumnMapからcardinality取得
		cols := strings.Split(tableMap[tbl], ",")
		for _, c := range cols {
			cardinality, err := samplingColumnCardinality(tbl, info.pkColumns, c, 1000)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Printf("    - %s = %d\n", c, cardinality)
		}
	}

}
