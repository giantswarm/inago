package infraconfigparser

import (
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

type NewConfigLoaderConfig struct {
	FlagPath       string
	CollectionPath string
	GroupPath      string
	UnitPath       string
}

type ConfigLoader struct {
	NewConfigLoaderConfig
}

func NewConfigLoader(c NewConfigLoaderConfig) ConfigLoader {
	return ConfigLoader{c}
}

type Config struct {
	AllFlags map[string]interface{} `json:"all-flags"`

	Flags       Flags        `json:"flags"`
	Collections []Collection `json:"collections,omitempty"`

	// At some point this should just be a list of services, and not groups and
	// units.
	Groups []Group `json:"groups,omitempty"`
	Units  []Unit  `json:"units,omitempty"`
}

func (c Config) Viper() *viper.Viper {
	v := viper.New()
	for key, val := range c.AllFlags {
		v.Set(key, val)
	}
	return v
}

// TODO Config.CollectionByName(name string) Config
// TODO Config.GroupByName(name string) Config

func (cl ConfigLoader) LoadConfig(dynFlags *viper.Viper) (Config, error) {
	// At first load flags and extend them using dynamic flags.
	flags, err := cl.LoadAllFlags(dynFlags)
	if err != nil {
		return Config{}, maskAny(err)
	}
	allFlags := flags.OverwriteWith(dynFlags)

	// Now load all further configurations and inject flags and env configurations.
	collections, err := cl.LoadAllCollections(allFlags)
	if err != nil {
		return Config{}, maskAny(err)
	}
	groups, err := cl.LoadAllGroups(allFlags)
	if err != nil {
		return Config{}, maskAny(err)
	}
	units, err := cl.LoadAllUnits(allFlags)
	if err != nil {
		return Config{}, maskAny(err)
	}

	config := Config{
		AllFlags:    allFlags.AllSettings(),
		Flags:       flags,
		Collections: collections,
		Groups:      groups,
		Units:       units,
	}

	return config, nil
}
