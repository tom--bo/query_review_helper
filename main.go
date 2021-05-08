package main

import (
	"bufio"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	toml "github.com/pelletier/go-toml"
	"io"
	"os"
)

var (
	DEBUG           = false
	db              *sqlx.DB
	reader          = bufio.NewReaderSize(os.Stdin, 1000000)
	confFile        string
	host            string
	port            int
	limit           int
	user            string
	password        string
	database        string
	socket          string
	indexFlag       bool
	cardinalityFlag bool
	explainFlag     bool
	showCreateFlag  bool
)

func parseOptions() {
	flag.BoolVar(&DEBUG, "debug", false, "DEBUG mode")

	flag.BoolVar(&indexFlag, "i", false, "show indexes")
	flag.BoolVar(&cardinalityFlag, "c", false, "show cardinalities")
	flag.BoolVar(&explainFlag, "e", false, "show explain results")
	// flag.BoolVar(&showCreateFlag, "s", false, "show show create table results")
	flag.IntVar(&limit, "l", 5000, "limitation for cardinality sampling")

	flag.StringVar(&confFile, "f", "", "conf file for auth info")
	flag.StringVar(&host, "h", "localhost", "mysql host")
	flag.IntVar(&port, "P", 3306, "mysql port")
	flag.StringVar(&user, "u", "mysql", "mysql user")
	flag.StringVar(&password, "p", "password", "mysql password")
	flag.StringVar(&database, "d", "", "mysql database")
	flag.StringVar(&socket, "S", "", "mysql unix domain socket")

	flag.Parse()
}

func readConf() error {
	bytes, err := os.ReadFile(confFile)
	if err != nil {
		return err
	}
	conf, err := toml.Load(string(bytes))
	user = conf.Get("auth.user").(string)
	password = conf.Get("auth.password").(string)
	port = int(conf.Get("auth.port").(int64))

	return nil
}

func main() {
	fmt.Println("(Input query and ^D at the last line)")
	parseOptions()
	// if all view-flag is not specified, set all view-flags true (show all information in default)
	if !indexFlag && !cardinalityFlag && !explainFlag && !showCreateFlag {
		indexFlag = true
		cardinalityFlag = true
		explainFlag = true
		showCreateFlag = true
	}
	if confFile != "" {
		readConf()
	}

	// Get query
	q := readLine()

	// Connect MySQL
	err := connectMySQL()
	if err != nil {
		fmt.Println("Connect MySQL err", err.Error())
		os.Exit(1)
	}

	start(q)
}

func readLine() string {
	buf := make([]byte, 0, 100000)
	for {
		l, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		buf = append(buf, l...)
	}
	return string(buf)
}
