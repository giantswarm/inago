package infraconfigparser

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/DisposaBoy/JsonConfigReader"
)

func unmarshalJSONFromBuffer(buf *bytes.Buffer, val interface{}) error {
	r := JsonConfigReader.New(buf)
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return maskAny(err)
	}
	err = json.Unmarshal(b, &val)
	if err != nil {
		inspectJSONError(err, b)
		return maskAny(err)
	}

	return nil
}
