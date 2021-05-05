package main

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type TableInfo struct {
	name    string
	index   IndexInfo
	columns map[string]int
}

func start(q string) {
	// parse
	tables, columns, err := getTableColumnsFromSQL(q)
	if err != nil {
		fmt.Println("parse select-stmt err", err.Error())
		os.Exit(1)
	}

	// Build tableMap
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

	// check columns in query
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
	if len(orphanColumns) > 0 {
		err = assignOrphanColumns(tableMap, tables, orphanColumns)
		if err != nil {
			fmt.Println("assignOrphanColumn err", err.Error())
			os.Exit(1)
		}
	}

	// Summarize result
	tableInfos := []TableInfo{}
	for _, tbl := range tables {
		// Get index info
		info, err := getKeys(database, tbl)
		if err != nil {
			fmt.Println("Get keys err", err.Error())
			os.Exit(1)
		}
		if info.pkColumns == "" {
			// Error and skip if PK not found
			fmt.Printf("[WARNING] %s table do not have Primary key, skip this table.", tbl)
			continue
		}
		tblInfo := TableInfo{
			name:    tbl,
			index:   info,
			columns: make(map[string]int),
		}

		// Column Cardinality
		cols := strings.Split(tableMap[tbl], ",")
		for _, c := range cols {
			if _, ok := tblInfo.columns[c]; ok || c == "" {
				continue
			}
			cardinality, err := samplingColumnCardinality(tbl, info.pkColumns, c, 5000)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			tblInfo.columns[c] = cardinality
		}
		tableInfos = append(tableInfos, tblInfo)
	}

	// print result
	for _, tbl := range tableInfos {
		// Table name
		fmt.Println("\n- " + tbl.name)

		// Index
		fmt.Println("  - Index")
		fmt.Printf("    - %-20s: (%s)\n", "PRIMARY", tbl.index.pkColumns)
		for k, v := range tbl.index.indexes {
			fmt.Printf("    - %-20s: (%s)\n", v, k)
		}

		// Column Cardinality
		fmt.Println("  - Cardinality")
		for col, cardi := range tbl.columns {
			fmt.Printf("    - %s = %d\n", col, cardi)
		}
	}
}
