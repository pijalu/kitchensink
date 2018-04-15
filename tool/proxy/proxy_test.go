package proxy

import (
	"context"
	"strings"
	"testing"
)

func _TestCopyConn(t *testing.T, expected string) {
	input := strings.NewReader(expected)
	output := strings.Builder{}

	ctx, cancel := context.WithCancel(context.Background())
	r := proxyRequest{
		ctx:    ctx,
		cancel: cancel,
	}
	r.copyConn(input, &output)

	actual := output.String()
	if expected != actual {
		t.Fatalf("Expected %s but got %s", expected, actual)
	}

	select {
	case <-ctx.Done():
	default:
		t.Fatal("Done should have been called")
	}
}

func TestCopyConn(t *testing.T) {
	for _, testCase := range []string{"Hello World", "", " "} {
		_TestCopyConn(t, testCase)
	}
}
