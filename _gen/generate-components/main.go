package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/quickfixgo/quickfix/_gen"
	"github.com/quickfixgo/quickfix/datadictionary"
)

var (
	pkg     string
	fixSpec *datadictionary.DataDictionary
	imports map[string]bool
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: generate-components [flags] <path to data dictionary>\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func initPackage() {
	pkg = strings.ToLower(fixSpec.FIXType) + strconv.Itoa(fixSpec.Major) + strconv.Itoa(fixSpec.Minor)

	if fixSpec.ServicePack != 0 {
		pkg += "sp" + strconv.Itoa(fixSpec.ServicePack)
	}
}

func packageString() (s string) {
	s = fmt.Sprintf("package %v\n", pkg)
	return
}

func genComponentImports() (fileOut string) {

	if len(imports) == 0 {
		return
	}

	fileOut += "import(\n"
	for i := range imports {
		fileOut += fmt.Sprintf("\"%v\"\n", i)
	}
	fileOut += ")\n"

	return
}

type group struct {
	parent string
	field  *datadictionary.FieldDef
}

func collectGroups(parent string, part datadictionary.MessagePart, groups []group) []group {
	switch field := part.(type) {
	case *datadictionary.FieldDef:
		if !field.IsGroup() {
			return groups
		}

		groups = append(groups, group{parent, field})
		for _, childField := range field.Parts {
			groups = collectGroups(field.Name(), childField, groups)
		}
	}

	return groups
}

func writeFieldDeclaration(part datadictionary.MessagePart, componentName string) string {
	switch field := part.(type) {
	case datadictionary.Component:
		imports[fmt.Sprintf("github.com/quickfixgo/quickfix/%v/%v", pkg, strings.ToLower(field.Name()))] = true
	case *datadictionary.FieldDef:
		if !field.IsGroup() {
			goType := gen.FixFieldTypeToGoType(field.Type)
			if goType == "time.Time" {
				imports["time"] = true
			}
		}
	}

	return gen.WriteFieldDeclaration(fixSpec.Major, fixSpec.Minor, part, componentName)
}

func genGroupDeclaration(field *datadictionary.FieldDef, parent string) (fileOut string) {
	fileOut += fmt.Sprintf("//%v is a repeating group in %v\n", field.Name(), parent)
	fileOut += fmt.Sprintf("type %v struct {\n", field.Name())
	for _, groupField := range field.Parts {
		fileOut += writeFieldDeclaration(groupField, field.Name())
	}

	fileOut += "}\n"

	return
}

func genGroupDeclarations(name string, fields []datadictionary.MessagePart) (fileOut string) {
	groups := []group{}
	for _, field := range fields {
		groups = collectGroups(name, field, groups)
	}

	for _, group := range groups {
		fileOut += genGroupDeclaration(group.field, group.parent)
	}

	return
}

func genHeader(header *datadictionary.MessageDef) {
	imports = make(map[string]bool)

	//delay field output to determine imports
	delayOut := genGroupDeclarations("Header", header.Parts)
	delayOut += fmt.Sprintf("//Header is the %v Header type\n", pkg)
	delayOut += "type Header struct {\n"
	for _, field := range header.Parts {
		delayOut += writeFieldDeclaration(field, "Header")
	}
	delayOut += "}\n"

	fileOut := packageString()
	fileOut += genComponentImports()
	fileOut += delayOut

	writer := new(bytes.Buffer)
	if err := gen.WriteFieldSetters(writer, "Header", header.Parts); err != nil {
		panic(err)
	}
	fileOut += writer.String()

	gen.WriteFile(path.Join(pkg, "header.go"), fileOut)
}

func genTrailer(trailer *datadictionary.MessageDef) {
	imports = make(map[string]bool)
	fileOut := packageString()
	fileOut += fmt.Sprintf("//Trailer is the %v Trailer type\n", pkg)
	fileOut += "type Trailer struct {\n"
	for _, field := range trailer.Parts {
		fileOut += writeFieldDeclaration(field, "Trailer")
	}
	fileOut += "}\n"

	writer := new(bytes.Buffer)
	if err := gen.WriteFieldSetters(writer, "Trailer", trailer.Parts); err != nil {
		panic(err)
	}
	fileOut += writer.String()

	gen.WriteFile(path.Join(pkg, "trailer.go"), fileOut)
}

func genComponent(name string, component *datadictionary.ComponentType) {
	imports = make(map[string]bool)

	//delay output to determine imports
	delayOut := genGroupDeclarations(name, component.Parts)
	delayOut += fmt.Sprintf("//%v is a %v Component\n", name, pkg)
	delayOut += fmt.Sprintf("type %v struct {\n", name)
	for _, field := range component.Parts {
		delayOut += writeFieldDeclaration(field, name)
	}
	delayOut += "}\n"

	fileOut := fmt.Sprintf("package %v\n", strings.ToLower(name))
	fileOut += genComponentImports()
	fileOut += delayOut

	fileOut += genComponentSetters(component)

	gen.WriteFile(path.Join(pkg, strings.ToLower(name), name+".go"), fileOut)
}

func genComponentSetters(component *datadictionary.ComponentType) string {
	writer := new(bytes.Buffer)
	if err := gen.WriteFieldSetters(writer, component.Name(), component.Parts); err != nil {
		panic(err)
	}

	return writer.String()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}

	dataDict := flag.Arg(0)

	if spec, err := datadictionary.Parse(dataDict); err != nil {
		panic(err)
	} else {
		fixSpec = spec
	}

	initPackage()
	if fi, err := os.Stat(pkg); os.IsNotExist(err) {
		if err := os.Mkdir(pkg, os.ModePerm); err != nil {
			panic(err)
		}
	} else if !fi.IsDir() {
		panic(pkg + "/ is not a directory")
	}

	switch pkg {
	//uses fixt11 header/trailer
	case "fix50", "fix50sp1", "fix50sp2":
	default:
		genHeader(fixSpec.Header)
		genTrailer(fixSpec.Trailer)
	}

	for name, component := range fixSpec.ComponentTypes {
		genComponent(name, component)
	}
}
