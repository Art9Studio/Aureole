package main

import (
	"encoding/json"
	"github.com/kr/pretty"
	"testing"
)

func Test_Session_RawExec(t *testing.T) {
	var pointsJSON = []byte(`{"id": "i1", "x":4, "y":-5}`)
	var pointsData interface{}
	err := json.Unmarshal(pointsJSON, &pointsData)
	if err != nil {
		t.Error(err)
	}

	str, err := GetJSONPath("{$.x}{$.y}", pointsData)
	_, _ = pretty.Print(str)
}
