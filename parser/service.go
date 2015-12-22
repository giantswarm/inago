package parser

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
)

type UnitTmpl struct {
	// general

	// UnitType is the GS specific unit type like lb-register or ambassador.
	GSType string `json:"gs-type,omitempty"`

	// SystemdType is the Type statement of a systemd unit file's Service
	// section. E.g. oneshot.
	SystemdType string `json:"systemd-type,omitempty"`
	Iptables    bool   `json:"iptables,omitempty"`

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

type ServiceTmpl struct {
	Name      string     `json:"name,omitempty"`
	After     string     `json:"after,omitempty"`
	Scale     int        `json:"scale,omitempty"`
	Conflicts []string   `json:"conflicts,omitempty"`
	Units     []UnitTmpl `json:"units,omitempty"`
}

func (tl TmplLoader) LoadAllServiceTmpls(allFlags *viper.Viper) ([]ServiceTmpl, error) {
	serviceTmpls := []ServiceTmpl{}

	tmplPaths, err := tmplPathsRecursive(tl.servicePath, tmplExt)
	if err != nil {
		return nil, maskAny(err)
	}
	partialPaths, err := tmplPathsRecursive(tl.servicePath, partialExt)
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
		var st ServiceTmpl
		err = unmarshalJSONFromBuffer(buffer, &st)
		if err != nil {
			return nil, maskAny(err)
		}

		serviceTmpls = append(serviceTmpls, st)
	}

	return serviceTmpls, nil
}
