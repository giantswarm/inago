package parser

import (
	"testing"

	"github.com/spf13/viper"
)

func Test_Collection_Simple(t *testing.T) {
	c := NewTmplLoaderConfig{
		TmplPath: "../fixture/simple",
	}
	tl := NewTmplLoader(c)

	collections, err := tl.LoadAllCollections(nil)
	if err != nil {
		t.Fatalf("LoadAllFlags failed: %#v", err)
	}
	if len(collections) != 1 {
		t.Fatalf("len(collections) != 1: %#v", len(collections))
	}
	if collections[0].Name != "collection" {
		t.Fatalf("collections[0].Name != \"collection\": %#v", collections[0].Name)
	}
	if len(collections[0].Tmpl.Groups) != 3 {
		t.Fatalf("len(collections[0].Tmpl.Groups) != 3: %#v", len(collections[0].Tmpl.Groups))
	}
	if collections[0].Tmpl.Groups[0] != "group" {
		t.Fatalf("collections[0].Tmpl.Groups[0] != \"group\": %#v", collections[0].Tmpl.Groups[0])
	}
	if collections[0].Tmpl.Groups[1] != "group" {
		t.Fatalf("collections[0].Tmpl.Groups[1] != \"group\": %#v", collections[0].Tmpl.Groups[1])
	}
	if collections[0].Tmpl.Groups[1] != "group" {
		t.Fatalf("collections[0].Tmpl.Groups[1] != \"group\": %#v", collections[0].Tmpl.Groups[1])
	}
}

func Test_Collection_Subdir(t *testing.T) {
	c := NewTmplLoaderConfig{
		TmplPath: "../fixture/subdir",
	}
	tl := NewTmplLoader(c)

	v := viper.New()
	v.Set("flavor", "y")

	collections, err := tl.LoadAllCollections(v)
	if err != nil {
		t.Fatalf("LoadAllFlags failed: %#v", err)
	}
	if len(collections) != 1 {
		t.Fatalf("len(collections) != 1: %#v", len(collections))
	}

	if collections[0].Name != "collection" {
		t.Fatalf("collections[0].Name != \"collection\": %#v", collections[0].Name)
	}
	if len(collections[0].Tmpl.Groups) != 3 {
		t.Fatalf("len(collections[0].Tmpl.Groups) != 3: %#v", len(collections[0].Tmpl.Groups))
	}
	if collections[0].Tmpl.Groups[0] != "subdir" {
		t.Fatalf("collections[0].Tmpl.Groups[0] != \"subdir\": %#v", collections[0].Tmpl.Groups[0])
	}
	if collections[0].Tmpl.Groups[1] != "subdir" {
		t.Fatalf("collections[0].Tmpl.Groups[1] != \"subdir\": %#v", collections[0].Tmpl.Groups[1])
	}
	if collections[0].Tmpl.Groups[2] != "y" {
		t.Fatalf("collections[0].Tmpl.Groups[2] != \"y\": %#v", collections[0].Tmpl.Groups[2])
	}
}
