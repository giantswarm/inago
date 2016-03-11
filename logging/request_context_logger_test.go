package logging

import (
	"reflect"
	"testing"

	"github.com/giantswarm/request-context"
)

// TestContextToCtx tests the contextToCtx function.
func TestContextToCtx(t *testing.T) {
	var tests = []struct {
		context Context
		ctx     requestcontext.Ctx
	}{
		{nil, nil},
		{
			Context{"a": 1},
			requestcontext.Ctx{"a": 1},
		},
		{
			Context{"a": 1, "b": 2},
			requestcontext.Ctx{"a": 1, "b": 2},
		},
	}

	for _, test := range tests {
		if !reflect.DeepEqual(contextToCtx(test.context), test.ctx) {
			t.Errorf("context %v did not map to ctx %v", test.context, test.ctx)
		}
	}
}
