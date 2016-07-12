package main

import (
	"fmt"
	"log"

	"github.com/achiku/dgw"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	connStr         = kingpin.Arg("conn", "PostgreSQL connection string in URL format").Required().String()
	schema          = kingpin.Flag("schema", "PostgreSQL schema name").Default("public").String()
	typeMapFilePath = kingpin.Flag("typemap", "type map file path").String()
)

func main() {
	kingpin.Parse()

	conn, err := dgw.OpenDB(*connStr)
	if err != nil {
		log.Fatal(err)
	}

	src, err := dgw.PgCreateStruct(conn, *schema, *typeMapFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", src)
}
