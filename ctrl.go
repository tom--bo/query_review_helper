package main

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func start(q string) {
	// parse
	tables, columns, err := getTableColumnsFromSQL(q)
	if err != nil {
		fmt.Println("parse select-stmt err")
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// build tableMap
	tableMap := make(map[string]string)
	orphanColumns := []string{}

	if DEBUG {
		fmt.Println("== Tables\n", tables)
	}

	for _, tbl := range tables {
		tableMap[tbl] = ""
	}
	if DEBUG {
		fmt.Println("== Columns\n", columns)
	}
	for _, col := range columns {
		c := strings.Split(col, ".")
		if len(c) < 2 {
			fmt.Println("Something bad happen")
			os.Exit(1)
		}

		_, ok := tableMap[c[0]]
		if !ok || c[0] == "" {
			orphanColumns = append(orphanColumns, c[1])
		} else {
			tableMap[c[0]] += "," + c[1]
		}
	}
	if DEBUG {
		fmt.Println("tableMap\n", tableMap, "\n======")
	}

	// assign orphan columns
	err = assignOrphanColumns(tableMap, tables, orphanColumns)
	if err != nil {
		fmt.Println("assignOrphanColumn err", err.Error())
		os.Exit(1)
	}

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
		cols := strings.Split(tableMap[tbl], ",")
		for _, c := range cols {
			if c == "" {
				continue
			}
			cardinality, err := samplingColumnCardinality(tbl, info.pkColumns, c, 1000)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Printf("    - %s = %d\n", c, cardinality)
		}
	}

}
