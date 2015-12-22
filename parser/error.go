package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/juju/errgo"
)

var (
	maskAny = errgo.MaskFunc(errgo.Any)
)

var (
	flagNotFoundError = errgo.New("flag not found")
)

func exitOnError(err error, annotations ...interface{}) {
	if err != nil {
		inspectJSONError(err, annotations...)
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func inspectJSONError(err error, annotations ...interface{}) {
	if err == nil || len(annotations) == 0 {
		return
	}

	fmt.Printf("---- error ----\n")
	fmt.Printf("\n")

	serr, ok := errgo.Cause(err).(*json.SyntaxError)
	if ok {
		if len(annotations) != 1 {
			return
		}

		raw := []byte{}
		if r, ok := annotations[0].([]byte); ok {
			raw = r
		} else {
			return
		}
		data := string(raw)

		start, end := strings.LastIndex(data[:serr.Offset], "\n")+1, len(data)
		if idx := strings.Index(data[start:], "\n"); idx >= 0 {
			end = start + idx
		}

		fmt.Printf("%s\n", data[indexBoundary(data, start-100):indexBoundary(data, end+100)])
	} else {
		fmt.Printf("  no information available\n")
	}

	fmt.Printf("\n")
	fmt.Printf("---- error ----\n")
	fmt.Printf("\n")
}

func indexBoundary(data string, boundary int) int {
	if boundary < 0 {
		// When data length is 90 but we want to show boundaries up to length of
		// -100 we need to return the minimum of data.
		return 0
	}

	if len(data) < boundary {
		// When data length is 90 but we want to show boundaries up to length of
		// 100 we need to return the maximum of data.
		return len(data) - 1
	}

	// We did
	return boundary
}
