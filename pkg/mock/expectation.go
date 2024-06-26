package mock

import (
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/khulnasoft/harbor-scanner-tunnel/pkg/tunnel"
)

const Anything = mock.Anything

// Expectation represents an expectation of a method being called and its return values.
type Expectation struct {
	Method     string
	Args       []interface{}
	ReturnArgs []interface{}
}

// ApplyExpectations applies the specified expectations on a given mock.
func ApplyExpectations(t *testing.T, mock interface{}, expectations ...*Expectation) {
	t.Helper()
	if len(expectations) == 0 || expectations[0] == nil {
		return
	}
	switch v := mock.(type) {
	case *Enqueuer:
		m := mock.(*Enqueuer)
		for _, e := range expectations {
			m.On(e.Method, e.Args...).Return(e.ReturnArgs...)
		}
	case *Store:
		m := mock.(*Store)
		for _, e := range expectations {
			m.On(e.Method, e.Args...).Return(e.ReturnArgs...)
		}
	case *tunnel.MockWrapper:
		m := mock.(*tunnel.MockWrapper)
		for _, e := range expectations {
			m.On(e.Method, e.Args...).Return(e.ReturnArgs...)
		}
	case *Transformer:
		m := mock.(*Transformer)
		for _, e := range expectations {
			m.On(e.Method, e.Args...).Return(e.ReturnArgs...)
		}
	default:
		t.Fatalf("Unrecognized mock type: %T!", v)
	}
}
