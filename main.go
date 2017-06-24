package main

import (
	"io"
	"log"
	"os"
	"os/exec"

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

	var out io.Writer
	if *outFile != "" {
		out, err = os.Create(*outFile)
		if err != nil {
			log.Fatalf("failed to create output file %s: %s", *outFile, err)
		}
	} else {
		out = os.Stdout
	}
	if _, err := out.Write(src); err != nil {
		log.Fatal(err)
	}
	if *outFile != "" {
		params := []string{"-w", *outFile}
		if err := exec.Command("goimports", params...).Run(); err != nil {
			log.Fatalf("failed to goimports: %s", err)
		}
	}
}
