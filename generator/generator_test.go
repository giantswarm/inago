package generator

import (
	"testing"

	"github.com/giantswarm/infra-tmpl-go/parser"
	"github.com/spf13/viper"
)

func Test_Generator_Unitfiles(t *testing.T) {
	c := parser.NewTmplLoaderConfig{
		TmplPath: "../fixture/simple",
	}
	tl := parser.NewTmplLoader(c)

	v := viper.New()
	v.Set("flavor", "x")

	tmpls, err := tl.LoadTmpls(v)
	if err != nil {
		t.Fatalf("LoadTmpls failed: %#v", err)
	}

	unitfiles, err := GenerateUnitFiles(tmpls)
	if err != nil {
		t.Fatalf("GenerateUnitFiles failed: %#v", err)
	}

	t.Logf(">>>> tmpls: %#v", tmpls)
	t.Logf("")
	for _, ufs := range unitfiles {
		t.Logf("%s", ufs.Content)
		t.Logf("")
	}

	t.Fatalf("end")
}
