package main

import (
	"bufio"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"io"
	"os"
)

var (
	db       *sqlx.DB
	rdr      = bufio.NewReaderSize(os.Stdin, 1000000)
	host     string
	port     int
	user     string
	password string
	database string
	socket   string
)

func parseOptions() {
	flag.StringVar(&host, "h", "localhost", "mysql host")
	flag.IntVar(&port, "P", 3306, "mysql port")
	flag.StringVar(&user, "u", "root", "mysql user")
	flag.StringVar(&password, "p", "", "mysql password")
	flag.StringVar(&database, "d", "", "mysql database")
	flag.StringVar(&socket, "S", "", "mysql unix domain socket")

	flag.Parse()
}

func main() {
	fmt.Println("(Input query and ^D at the last line)")
	parseOptions()

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
	buf := make([]byte, 0, 1000000)
	for {
		l, err := rdr.ReadString('\n')
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
