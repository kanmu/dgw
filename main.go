package main

import (
	"fmt"
	"log"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	connStr         = kingpin.Arg("conn", "PostgreSQL connection string in URL format").Required().String()
	schema          = kingpin.Flag("schema", "PostgreSQL schema name").Default("public").String()
	pkgName         = kingpin.Flag("package", "package name").Default("main").String()
	typeMapFilePath = kingpin.Flag("typemap", "column type and go type map file path").String()
	colMapFilePath  = kingpin.Flag("colmap", "table column name and go type map file path").String()
)

func main() {
	kingpin.Parse()

	conn, err := OpenDB(*connStr)
	if err != nil {
		log.Fatal(err)
	}

	src, err := PgCreateStruct(conn, *schema, *typeMapFilePath, *pkgName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", src)
}
