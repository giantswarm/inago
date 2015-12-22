package infratmpl

import (
	"github.com/giantswarm/infra-tmpl-go/generator"
	"github.com/giantswarm/infra-tmpl-go/parser"
)

var (
	_ = generator.Service{}
	_ = parser.ServiceTmpl{}
)
