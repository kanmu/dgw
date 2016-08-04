package main

import (
	"fmt"
	"html/template"
	"strconv"
)

var tmplFuncMap = template.FuncMap{
	"shortname":      shortname,
	"fieldnames":     fieldnames,
	"fieldparams":    fieldparams,
	"colnames":       colnames,
	"colname":        colname,
	"colcount":       colcount,
	"colprefixnames": colprefixnames,
	"colvals":        colvals,
}

func colname(col *PgColumn) string {
	return col.Name
}

func colprefixnames(fields []*StructField, prefix string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + f.Name
		i++
	}

	return str
}

func colcount(fields []*StructField, ignoreNames ...string) int {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	i := 1
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		i++
	}
	return i
}

func colvals(fields []*StructField, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + fmt.Sprintf("$%d", i)
		i++
	}

	return str
}

func colnames(fields []*StructField, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + f.Name
		i++
	}

	return str
}

func fieldparams(fields []*StructField, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + "$" + strconv.Itoa(i)
		i++
	}
	return str
}

func shortname(name string) string {
	return "rcv"
}

func fieldnames(fields []*StructField, prefix string, ignoreNames ...string) string {
	ignore := map[string]bool{}
	for _, n := range ignoreNames {
		ignore[n] = true
	}

	str := ""
	i := 0
	for _, f := range fields {
		if ignore[f.Name] {
			continue
		}

		if i != 0 {
			str = str + ", "
		}
		str = str + prefix + "." + f.Name
		i++
	}
	return str
}
