package generator

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"text/template"

	"github.com/giantswarm/infra-tmpl-go/parser"
)

const (
	unitTmpl = "template/unit.tmpl"
)

var (
	trimSpaceRegex = regexp.MustCompile(`(?m)(^\s+)`)
)

type UnitFile struct {
	Name    string
	Service string
	Content string
}

type UnitFiles []UnitFile

func (ufs UnitFiles) ForService(name string) (UnitFiles, error) {
	list := UnitFiles{}

	for _, uf := range ufs {
		if uf.Service == name {
			list = append(list, uf)
		}
	}

	return list, nil
}

func GenerateUnitFiles(c parser.Tmpls) (UnitFiles, error) {
	ufs := UnitFiles{}

	for _, s := range c.Services {
		genService := mapParserServiceToGeneratorService(s)

		for _, genUnit := range genService.Units {
			sections := NewSections(genService, genUnit)
			raw, err := parseUnitTemplate(sections)
			if err != nil {
				return UnitFiles{}, maskAny(err)
			}

			uf := UnitFile{
				Name:    genUnit.Name,
				Service: genService.Name,
				Content: string(raw),
			}

			ufs = append(ufs, uf)
		}
	}

	return ufs, nil
}

func parseUnitTemplate(sections Sections) ([]byte, error) {
	// read file
	raw, err := ioutil.ReadFile(unitTmpl)
	if err != nil {
		return nil, maskAny(err)
	}

	// parse template
	t, err := template.New(unitTmpl).Parse(string(raw))
	if err != nil {
		return nil, maskAny(err)
	}
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, sections)
	if err != nil {
		return nil, maskAny(err)
	}

	b := trimSpaceRegex.ReplaceAll(buffer.Bytes(), []byte{})

	return b, nil
}
