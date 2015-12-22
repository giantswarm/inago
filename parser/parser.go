package parser

import (
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	tmplExt    = "tmpl"
	partialExt = "partial"
)

var (
	dynFlags = StringSlice{}
)

func DynFlags() *viper.Viper {
	return dynFlags.Viper()
}

func init() {
	pflag.Var(&dynFlags, "flag", "dynamically overwrite (multiple) template flags")
}

type NewTmplLoaderConfig struct {
	TmplPath string
}

type TmplLoader struct {
	NewTmplLoaderConfig

	collectionPath string
	flagPath       string
	servicePath    string
}

func NewTmplLoader(c NewTmplLoaderConfig) TmplLoader {
	tl := TmplLoader{
		NewTmplLoaderConfig: c,
	}

	tl.collectionPath = filepath.Join(tl.TmplPath, "collection")
	tl.flagPath = filepath.Join(tl.TmplPath, "flag")
	tl.servicePath = filepath.Join(tl.TmplPath, "service")

	return tl
}

type Tmpls struct {
	AllFlags map[string]interface{} `json:"all-flags"`

	Flags       Flags         `json:"flags"`
	Collections []Collection  `json:"collections,omitempty"`
	Services    []ServiceTmpl `json:"services,omitempty"`
}

func (c Tmpls) Viper() *viper.Viper {
	v := viper.New()
	for key, val := range c.AllFlags {
		v.Set(key, val)
	}
	return v
}

// TODO Tmpls.CollectionByName(name string) Tmpls
// TODO Tmpls.GroupByName(name string) Tmpls

func (cl TmplLoader) LoadTmpls(dynFlags *viper.Viper) (Tmpls, error) {
	// At first load flags and extend them using dynamic flags.
	flags, err := cl.LoadAllFlags(dynFlags)
	if err != nil {
		return Tmpls{}, maskAny(err)
	}
	allFlags := flags.OverwriteWith(dynFlags)

	// Now load all further configurations and inject flags and env configurations.
	collections, err := cl.LoadAllCollections(allFlags)
	if err != nil {
		return Tmpls{}, maskAny(err)
	}
	serviceTmpls, err := cl.LoadAllServiceTmpls(allFlags)
	if err != nil {
		return Tmpls{}, maskAny(err)
	}

	config := Tmpls{
		AllFlags:    allFlags.AllSettings(),
		Flags:       flags,
		Collections: collections,
		Services:    serviceTmpls,
	}

	return config, nil
}
