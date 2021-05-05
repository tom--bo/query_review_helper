package main

import (
	"fmt"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/parser/test_driver"
)

type colVisitor struct {
	colNames []string
	tblNames []string
}

func (v *colVisitor) Enter(in ast.Node) (ast.Node, bool) {
	if c, ok := in.(*ast.ColumnName); ok {
		v.colNames = append(v.colNames, c.Table.L+"."+c.Name.O)
	}
	if t, ok := in.(*ast.TableName); ok {
		v.tblNames = append(v.tblNames, t.Name.O)
	}
	return in, false
}

func (v *colVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}

func extract(rootNode *ast.StmtNode) ([]string, []string) {
	v := &colVisitor{}
	(*rootNode).Accept(v)
	return v.tblNames, v.colNames
}

func parse(sql string) (*ast.StmtNode, error) {
	p := parser.New()

	stmtNodes, _, err := p.Parse(sql, "", "")
	if err != nil {
		return nil, err
	}

	return &stmtNodes[0], nil
}

func getColumnsFromSQL(sql string) ([]string, []string, error) {
	astNode, err := parse(sql)
	if err != nil {
		fmt.Printf("parse error: %v\n", err.Error())
		return []string{}, []string{}, err
	}
	tbls, cols := extract(astNode)

	return tbls, cols, nil
}
