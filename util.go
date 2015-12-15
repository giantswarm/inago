package infraconfigparser

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/viper"
)

func configNameByPath(p string) string {
	base := filepath.Base(p)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func readFile(p string) ([]byte, error) {
	raw, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, maskAny(err)
	}
	if filepath.Ext(p) == ".gpg" {
		// For now .gpg need to be included excplicitly.
		raw, err = decryptGPGBytes(raw)
		if err != nil {
			return nil, maskAny(err)
		}
	}
	return raw, nil
}

func tmplPathsRecursive(dir, ext string) ([]string, error) {
	paths := []string{}

	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ext) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, maskAny(err)
	}

	return paths, nil
}

func parseAndExecuteTmpl(flags *viper.Viper, p string, v interface{}) error {
	// read file
	raw, err := readFile(p)
	if err != nil {
		return maskAny(err)
	}

	// parse template
	t, err := template.New(p).Funcs(tmplFuncs(flags)).Parse(string(raw))
	if err != nil {
		return maskAny(err)
	}
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, nil)
	if err != nil {
		return maskAny(err)
	}

	// parse json
	err = unmarshalJSONFromBuffer(buffer, v)
	if err != nil {
		return maskAny(err)
	}

	return nil
}
