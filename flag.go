package infraconfigparser

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
)

// rawFlag is a single flag describing what key-value pair can be configured.
type FlagDefinition struct {
	// Name is the flag name.
	Name string `json:"name"`

	// Value is the flag Value.
	Value interface{} `json:"value"`

	// Usage is the flag usage.
	Usage string `json:"usage"`

	// Required states whether this flag is allowed to be optional or not.
	Required bool `json:"required"`

	// Can be used to initialize a flags value the very first time. This can
	// either be a static well known or randomized value using available template
	// functions.
	OnSetup interface{} `json:"on-setup"`
}

// FlagTmpl is the structure defined in flag templates.
type FlagTmpl struct {
	// Overwrites defines what configuration flags this flags overwrite. It is an
	// error if the given configuration set cannot be found. It is also an error
	// if any flag supposed to be overwritten cannot be found in the given
	// configuration set.
	Overwrites string `json:"overwrites"`

	// FlagDefinition is a list of flag definitions. See FlagDefinition.
	FlagDefinitions []FlagDefinition `json:"flags"`
}

type Flag struct {
	Name string   `json:"name"`
	Tmpl FlagTmpl `json:"tmpl"`
}

type Flags []Flag

func (fs Flags) Viper() *viper.Viper {
	v := viper.New()
	for _, f := range fs {
		for _, flagDef := range f.Tmpl.FlagDefinitions {
			v.Set(flagDef.Name, flagDef.Value)
		}
	}
	return v
}

func (fs Flags) OverwriteWith(overwrite *viper.Viper) *viper.Viper {
	v := fs.Viper()
	for key, val := range overwrite.AllSettings() {
		v.Set(key, val)
	}
	return v
}

// LoadAllFlagTmpls tries to load all flags by the defined templates within a
// configuration set. There is a root template flag.tmpl. All other templates
// will be handled as partials. So they need to be included in the root
// template by the golang template statement. dynFlags is a viper containing
// dynamic flags given by --flag=foo:bar.
func (cl ConfigLoader) LoadAllFlags(dynFlags *viper.Viper) (Flags, error) {
	flags := Flags{}

	tmplPaths, err := tmplPathsRecursive(cl.FlagPath, tmplExt)
	if err != nil {
		return Flags{}, maskAny(err)
	}
	partialPaths, err := tmplPathsRecursive(cl.FlagPath, partialExt)
	if err != nil {
		return Flags{}, maskAny(err)
	}

	for _, tmplPath := range tmplPaths {
		// init template
		rootRaw, err := readFile(tmplPath)
		if err != nil {
			return Flags{}, maskAny(err)
		}
		rootTmpl, err := template.New(tmplPath).Funcs(tmplFuncs(dynFlags)).Parse(string(rootRaw))
		if err != nil {
			return Flags{}, maskAny(err)
		}

		// init partials
		for _, partialPath := range partialPaths {
			raw, err := readFile(partialPath)
			if err != nil {
				return Flags{}, maskAny(err)
			}
			t, err := template.New(partialPath).Funcs(tmplFuncs(dynFlags)).Parse(string(raw))
			if err != nil {
				return Flags{}, maskAny(err)
			}
			rootTmpl.AddParseTree(configNameByPath(partialPath), t.Tree)
		}

		// parse template
		buffer := new(bytes.Buffer)
		err = rootTmpl.Execute(buffer, nil)
		if err != nil {
			return Flags{}, maskAny(err)
		}
		var ft FlagTmpl
		err = unmarshalJSONFromBuffer(buffer, &ft)
		if err != nil {
			return Flags{}, maskAny(err)
		}

		f := Flag{
			Name: configNameByPath(tmplPath),
			Tmpl: ft,
		}
		flags = append(flags, f)
	}

	return flags, nil
}
