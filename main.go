package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	connStr          = kingpin.Arg("conn", "PostgreSQL connection string in URL format").Required().String()
	schema           = kingpin.Flag("schema", "PostgreSQL schema name").Default("public").Short('s').String()
	pkgName          = kingpin.Flag("package", "package name").Default("main").Short('p').String()
	typeMapFilePath  = kingpin.Flag("typemap", "column type and go type map file path").Short('t').String()
	excludeTableName = kingpin.Flag("exclude", "table names to exclude").Short('x').Strings()
	outFile          = kingpin.Flag("output", "output file path").Short('o').String()
)

func main() {
	kingpin.Parse()

	conn, err := OpenDB(*connStr)
	if err != nil {
		log.Fatal(err)
	}

	src, err := PgCreateStruct(conn, *schema, *typeMapFilePath, *pkgName, *excludeTableName)
	if err != nil {
		log.Fatal(err)
	}

	var out os.File
	if *outFile != "" {
		out, err := os.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		if _, err := out.Write(src); err != nil {
			log.Fatal(err)
		}
		out.Sync()
		return
	}
	out = *os.Stdout
	if _, err := out.Write(src); err != nil {
		log.Fatal(err)
	}
}
