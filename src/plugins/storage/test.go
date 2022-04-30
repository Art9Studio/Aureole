package storage

import (
	"aureole/internal/plugins"
	"math/rand"
	"strconv"
	"testing"

	"github.com/go-test/deep"
)

type Foo struct {
	Bar        string
	privateBar string
}

type privateFoo struct {
	Bar        string
	privateBar string
}

func TestStore(storage plugins.Storage, t *testing.T) {
	k := strconv.FormatInt(rand.Int63(), 10)

	// Initially the k shouldn't exist
	found, err := storage.Get(k, new(Foo))
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("A value was found, but no value was expected")
	}

	// Deleting a non-existing k-value pair should NOT lead to an error
	err = storage.Delete(k)
	if err != nil {
		t.Error(err)
	}

	// Store an object
	v := Foo{
		Bar: "baz",
	}
	err = storage.Set(k, v, 0)
	if err != nil {
		t.Error(err)
	}

	// Storing it again should not lead to an error but just overwrite it
	err = storage.Set(k, v, 0)
	if err != nil {
		t.Error(err)
	}

	// Retrieve the object
	expected := v
	actualPtr := new(Foo)
	found, err = storage.Get(k, actualPtr)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Error("No value was found, but should have been")
	}
	actual := *actualPtr
	if actual != expected {
		t.Errorf("Expected: %v, but was: %v", expected, actual)
	}

	// Delete
	err = storage.Delete(k)
	if err != nil {
		t.Error(err)
	}
	// Key-value pair shouldn't exist anymore
	found, err = storage.Get(k, new(Foo))
	if err != nil {
		t.Error(err)
	}
	if found {
		t.Error("A value was found, but no value was expected")
	}
}

func TestTypes(storage plugins.Storage, t *testing.T) {
	boolVar := true
	// Omit byte
	// Omit error - it's a Go builtin type but marshalling and then unmarshalling doesn't lead to equal objects
	floatVar := 1.2
	intVar := 1
	runeVar := '⚡'
	stringVar := "foo"

	structVar := Foo{
		Bar: "baz",
	}
	structWithPrivateFieldVar := Foo{
		Bar:        "baz",
		privateBar: "privBaz",
	}
	// The differing expected var for structWithPrivateFieldVar
	structWithPrivateFieldExpectedVar := Foo{
		Bar: "baz",
	}
	privateStructVar := privateFoo{
		Bar: "baz",
	}
	privateStructWithPrivateFieldVar := privateFoo{
		Bar:        "baz",
		privateBar: "privBaz",
	}
	// The differing expected var for privateStructWithPrivateFieldVar
	privateStructWithPrivateFieldExpectedVar := privateFoo{
		Bar: "baz",
	}

	sliceOfBool := []bool{true, false}
	sliceOfByte := []byte("foo")
	// Omit slice of float
	sliceOfInt := []int{1, 2}
	// Omit slice of rune
	sliceOfString := []string{"foo", "bar"}

	sliceOfSliceOfString := [][]string{{"foo", "bar"}}

	sliceOfStruct := []Foo{{Bar: "baz"}}
	sliceOfPrivateStruct := []privateFoo{{Bar: "baz"}}

	testVals := []struct {
		testName string
		v        interface{}
		expected interface{}
		testGet  func(*testing.T, plugins.Storage, string, interface{})
	}{
		{"bool", boolVar, boolVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(bool)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"float", floatVar, floatVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(float64)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"int", intVar, intVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(int)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"rune", runeVar, runeVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(rune)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"string", stringVar, stringVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(string)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"struct", structVar, structVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(Foo)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"struct with private field", structWithPrivateFieldVar, structWithPrivateFieldExpectedVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(Foo)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"private struct", privateStructVar, privateStructVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(privateFoo)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"private struct with private field", privateStructWithPrivateFieldVar, privateStructWithPrivateFieldExpectedVar, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new(privateFoo)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if actual != expected {
				t.Errorf("Expected: %v, but was: %v", expected, actual)
			}
		}},
		{"slice of bool", sliceOfBool, sliceOfBool, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new([]bool)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if diff := deep.Equal(actual, expected); diff != nil {
				t.Error(diff)
			}
		}},
		{"slice of byte", sliceOfByte, sliceOfByte, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new([]byte)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if diff := deep.Equal(actual, expected); diff != nil {
				t.Error(diff)
			}
		}},
		{"slice of int", sliceOfInt, sliceOfInt, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new([]int)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if diff := deep.Equal(actual, expected); diff != nil {
				t.Error(diff)
			}
		}},
		{"slice of string", sliceOfString, sliceOfString, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new([]string)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if diff := deep.Equal(actual, expected); diff != nil {
				t.Error(diff)
			}
		}},
		{"slice of slice of string", sliceOfSliceOfString, sliceOfSliceOfString, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new([][]string)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if diff := deep.Equal(actual, expected); diff != nil {
				t.Error(diff)
			}
		}},
		{"slice of struct", sliceOfStruct, sliceOfStruct, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new([]Foo)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if diff := deep.Equal(actual, expected); diff != nil {
				t.Error(diff)
			}
		}},
		{"slice of private struct", sliceOfPrivateStruct, sliceOfPrivateStruct, func(t *testing.T, s plugins.Storage, k string, expected interface{}) {
			actualPtr := new([]privateFoo)
			ok, err := s.Get(k, actualPtr)
			handleGetError(t, err, ok)
			actual := *actualPtr
			if diff := deep.Equal(actual, expected); diff != nil {
				t.Error(diff)
			}
		}},
	}

	for _, testVal := range testVals {
		t.Run(testVal.testName, func(t2 *testing.T) {
			key := strconv.FormatInt(rand.Int63(), 10)
			err := storage.Set(key, testVal.v, 0)
			if err != nil {
				t.Error(err)
			}
			testVal.testGet(t, storage, key, testVal.expected)
		})
	}
}

func handleGetError(t *testing.T, err error, ok bool) {
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("No value was ok, but should have been")
	}
}
