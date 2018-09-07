package db

import (
	"reflect"
	"testing"
)

func TestCloneStringMapHelper(t *testing.T) {
	source := map[string]interface{}{"field1": 1}
	results := cloneStringMap(source)
	if !reflect.DeepEqual(source, results) {
		t.Fatalf("Expected %v but got %v", source, results)
	}
}
