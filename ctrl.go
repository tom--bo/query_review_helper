package main

import (
	"fmt"
	"os"
	"strconv"
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
			cardinality, err := samplingColumnCardinality(tbl, info.pkColumns, c)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			tblInfo.columns[c] = cardinality
		}
		tableInfos = append(tableInfos, tblInfo)
	}

	// explain
	if explainFlag {
		fmt.Println("\n\n\n==== Explain Result ====")
		fmt.Printf("+----+-------------+----------------------+----------+----------------------+--------+------------------------------------------+----------+----------+------------+\n")
		fmt.Printf("| id | select_type |     table            |   type   |         key          | keylen |    ref                                   |   rows   | filtered |  extra     |\n")
		fmt.Printf("+----+-------------+----------------------+----------+----------------------+--------+------------------------------------------+----------+----------+------------+\n")
		explains, err := getExplainResult(q)
		if err != nil {
			fmt.Println("Failed to get EXPLAIN result", err.Error())
		}
		for _, e := range explains {
			id := e.ID
			selectType := e.SelectType
			table := e.Table
			accessType := "NULL"
			if e.Type.Valid {
				accessType = e.Type.String
			}
			key := "NULL"
			if e.Key.Valid {
				key = e.Key.String
			}
			keyLen := "NULL"
			if e.KeyLen.Valid {
				keyLen = strconv.Itoa(int(e.KeyLen.Int32))
			}
			ref := "NULL"
			if e.Ref.Valid {
				ref = e.Ref.String
			}
			extra := "NULL"
			if e.Extra.Valid {
				extra = e.Extra.String
			}
			fmt.Printf("| %2d | %-11s | %-20s | %-8s | %-20s | %-6s | %-40s | %-8d |   %-3.2f | %-10s |\n",
				id, selectType, table, accessType, key, keyLen, ref, e.Rows, e.Filtered, extra)
		}
		fmt.Printf("+----+-------------+----------------------+----------+----------------------+--------+------------------------------------------+----------+----------+------------+\n")
	}

	// print result
	fmt.Println("\n\n==== Tables in query ====")
	for _, tbl := range tableInfos {
		// Table name
		if indexFlag || cardinalityFlag {
			fmt.Println("")
		}
		fmt.Println("- " + tbl.name)

		// Index
		if indexFlag {
			fmt.Println("  - Index")
			fmt.Printf("    - %-20s: (%s)\n", "PRIMARY", tbl.index.pkColumns)
			for k, v := range tbl.index.indexes {
				fmt.Printf("    - %-20s: (%s)\n", v, k)
			}
		}

		// Column Cardinality
		if cardinalityFlag {
			fmt.Println("  - Cardinality")
			for col, cardi := range tbl.columns {
				fmt.Printf("    - %s = %d\n", col, cardi)
			}
		}
	}
}
