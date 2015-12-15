package infraconfigparser

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
)

type Unit struct {
	Name string   `json:"name,omitempty"`
	Tmpl UnitTmpl `json:"tmpl,omitempty"`
}

type UnitTmpl struct {
	// general
	Type string `json:"type,omitempty"`

	// service
	Name      string   `json:"name,omitempty"`
	Image     string   `json:"image,omitempty"`
	Version   string   `json:"version,omitempty"`
	ExecStart []string `json:"exec-start,omitempty"`

	// lb-register
	Dependency string `json:"dependency,omitempty"`
	Port       int    `json:"port,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

func (cl ConfigLoader) LoadAllUnits(allFlags *viper.Viper) ([]Unit, error) {
	units := []Unit{}

	tmplPaths, err := tmplPathsRecursive(cl.UnitPath, tmplExt)
	if err != nil {
		return nil, maskAny(err)
	}
	partialPaths, err := tmplPathsRecursive(cl.UnitPath, partialExt)
	if err != nil {
		return nil, maskAny(err)
	}

	for _, tmplPath := range tmplPaths {
		// init template
		rootRaw, err := readFile(tmplPath)
		if err != nil {
			return nil, maskAny(err)
		}
		rootTmpl, err := template.New(tmplPath).Funcs(tmplFuncs(allFlags)).Parse(string(rootRaw))
		if err != nil {
			return nil, maskAny(err)
		}

		// init partials
		for _, partialPath := range partialPaths {
			raw, err := readFile(partialPath)
			if err != nil {
				return nil, maskAny(err)
			}
			t, err := template.New(partialPath).Funcs(tmplFuncs(allFlags)).Parse(string(raw))
			if err != nil {
				return nil, maskAny(err)
			}
			rootTmpl.AddParseTree(configNameByPath(partialPath), t.Tree)
		}

		// parse template
		buffer := new(bytes.Buffer)
		err = rootTmpl.Execute(buffer, nil)
		if err != nil {
			return nil, maskAny(err)
		}
		var ut UnitTmpl
		err = unmarshalJSONFromBuffer(buffer, &ut)
		if err != nil {
			return nil, maskAny(err)
		}

		g := Unit{
			Name: configNameByPath(tmplPath),
			Tmpl: ut,
		}
		units = append(units, g)
	}

	return units, nil
}
