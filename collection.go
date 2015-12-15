package infraconfigparser

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
)

// CollectionTmpl is the structure defined in collection templates.
type CollectionTmpl struct {
	Groups []string `json:"groups,omitempty"`
}

type Collection struct {
	Name string         `json:"name,omitempty"`
	Tmpl CollectionTmpl `json:"tmpl,omitempty"`
}

func (cl ConfigLoader) LoadAllCollections(allFlags *viper.Viper) ([]Collection, error) {
	collections := []Collection{}

	tmplPaths, err := tmplPathsRecursive(cl.CollectionPath, tmplExt)
	if err != nil {
		return nil, maskAny(err)
	}
	partialPaths, err := tmplPathsRecursive(cl.CollectionPath, partialExt)
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
		var ct CollectionTmpl
		err = unmarshalJSONFromBuffer(buffer, &ct)
		if err != nil {
			return nil, maskAny(err)
		}

		c := Collection{
			Name: configNameByPath(tmplPath),
			Tmpl: ct,
		}
		collections = append(collections, c)
	}

	return collections, nil
}
