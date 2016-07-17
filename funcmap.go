package main

import (
	"html/template"
	"strconv"
)

var tmplFuncMap = template.FuncMap{
	"shortname":   shortname,
	"fieldnames":  fieldnames,
	"fieldparams": fieldparams,
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
