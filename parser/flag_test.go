package parser

import (
	"testing"

	"github.com/spf13/viper"
)

func Test_Flag_Simple(t *testing.T) {
	c := NewTmplLoaderConfig{
		TmplPath: "../fixture/simple",
	}
	tl := NewTmplLoader(c)

	v := viper.New()
	v.Set("flavor", "x")

	flags, err := tl.LoadAllFlags(v)
	if err != nil {
		t.Fatalf("LoadAllFlags failed: %#v", err)
	}
	if len(flags.Viper().AllKeys()) != 1 {
		t.Fatalf("len(flags.Viper().AllKeys()) != 1: %#v", len(flags.Viper().AllKeys()))
	}
	if flags.Viper().GetString("x") != "x" {
		t.Fatalf("flags.Viper().GetString(\"x\") != \"x\": %#v", flags.Viper().GetString("x"))
	}
}

func Test_Flag_Subdir(t *testing.T) {
	c := NewTmplLoaderConfig{
		TmplPath: "../fixture/subdir",
	}
	tl := NewTmplLoader(c)

	v := viper.New()
	v.Set("flavor", "y")

	flags, err := tl.LoadAllFlags(v)
	if err != nil {
		t.Fatalf("LoadAllFlags failed: %#v", err)
	}
	if len(flags.Viper().AllKeys()) != 1 {
		t.Fatalf("len(flags.Viper().AllKeys()) != 1: %#v", len(flags.Viper().AllKeys()))
	}
	if flags.Viper().GetString("y") != "y" {
		t.Fatalf("flags.Viper().GetString(\"y\") != \"y\": %#v", flags.Viper().GetString("y"))
	}
}
