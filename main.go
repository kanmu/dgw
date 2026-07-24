package main

import (
	"io"
	"log"
	"os"

	"golang.org/x/tools/imports"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	connStr = kingpin.Arg(
		"conn", "PostgreSQL connection string in URL format").Required().String()
	schema = kingpin.Flag(
		"schema", "PostgreSQL schema name").Default("public").Short('s').String()
	pkgName          = kingpin.Flag("package", "package name").Default("main").Short('p').String()
	typeMapFilePath  = kingpin.Flag("typemap", "column type and go type map file path").Short('t').String()
	autGenKeyList    = kingpin.Flag("autogenkey", "auto generate key list").Short('k').Strings()
	exTbls           = kingpin.Flag("exclude", "table names to exclude").Short('x').Strings()
	customTmpl       = kingpin.Flag("template", "custom template path").String()
	outFile          = kingpin.Flag("output", "output file path").Short('o').String()
	noQueryInterface = kingpin.Flag("no-interface", "output without Queryer interface").Bool()
	deprecated       = kingpin.Flag("deprecated", "deprecated table names").Strings()
	queryer          = kingpin.Flag("queryer", "Queryer type name").String()
	contextOnly      = kingpin.Flag("context-only", "output only Context-aware APIs (suppress non-Context Create/GetXByPk)").Bool()
	version          string
)

func init() {
	kingpin.Version(version)
}

func main() {
	kingpin.Parse()

	conn, err := OpenDB(*connStr)
	if err != nil {
		log.Fatal(err)
	}

	st, err := PgCreateStruct(conn, *schema, *typeMapFilePath, *pkgName, *customTmpl, *exTbls, *autGenKeyList, *deprecated, *queryer, *contextOnly)
	if err != nil {
		log.Fatal(err)
	}

	var src []byte
	if *noQueryInterface {
		src = st
	} else {
		q := []byte(queryInterface)
		src = append(st, q...)
	}

	if *outFile != "" {
		src, err = imports.Process(*outFile, src, nil)
		if err != nil {
			log.Fatalf("failed to goimports: %s", err)
		}
	}

	var out io.Writer
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			log.Fatalf("failed to create output file %s: %s", *outFile, err)
		}
		defer f.Close() //nolint:errcheck
		out = f
	} else {
		out = os.Stdout
	}

	if _, err := out.Write(src); err != nil {
		log.Fatal(err)
	}
}
