package infraconfigparser

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
)

type Group struct {
	Name string    `json:"name,omitempty"`
	Tmpl GroupTmpl `json:"tmpl,omitempty"`
}

type GroupTmpl struct {
	After     string   `json:"after,omitempty"`
	Scale     int      `json:"scale,omitempty"`
	Conflicts string   `json:"conflicts,omitempty"`
	Iptables  bool     `json:"iptables,omitempty"`
	Units     []string `json:"units,omitempty"`
}

func (cl ConfigLoader) LoadAllGroups(allFlags *viper.Viper) ([]Group, error) {
	groups := []Group{}

	tmplPaths, err := tmplPathsRecursive(cl.GroupPath, tmplExt)
	if err != nil {
		return nil, maskAny(err)
	}
	partialPaths, err := tmplPathsRecursive(cl.GroupPath, partialExt)
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
		var gt GroupTmpl
		err = unmarshalJSONFromBuffer(buffer, &gt)
		if err != nil {
			return nil, maskAny(err)
		}

		g := Group{
			Name: configNameByPath(tmplPath),
			Tmpl: gt,
		}
		groups = append(groups, g)
	}

	return groups, nil
}
