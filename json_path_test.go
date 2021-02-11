package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GetJSONPath_Simple(t *testing.T) {
	var pointsJSON = []byte(`{"id": "i1", "x":4, "y":-5}`)
	var pointsData interface{}

	err := json.Unmarshal(pointsJSON, &pointsData)
	assert.NoError(t, err)

	x, err := GetJSONPath("{$.x}", pointsData)
	assert.NoError(t, err)
	assert.Equal(t, 4.0, x)

	y, err := GetJSONPath("{$.y}", pointsData)
	assert.NoError(t, err)
	assert.Equal(t, -5.0, y)

	arr, err := GetJSONPath("{$.x}{$.y}", pointsData)
	assert.NoError(t, err)
	assert.Len(t, arr, 2)
}

func Test_GetJSONPath_Advanced(t *testing.T) {
	type m = map[string]interface{}
	type ar = []interface{}

	in := m{
		"store": m{
			"book": ar{
				m{
					"category": "reference",
					"author":   "Nigel Rees",
					"title":    "Sayings of the Century",
					"price":    8.95,
				},
				m{
					"category": "fiction",
					"author":   "Evelyn Waugh",
					"title":    "Sword of Honour",
					"price":    12.99,
				},
				m{
					"category": "fiction",
					"author":   "Herman Melville",
					"title":    "Moby Dick",
					"isbn":     "0-553-21311-3",
					"price":    8.99,
				},
				m{
					"category": "fiction",
					"author":   "J. R. R. Tolkien",
					"title":    "The Lord of the Rings",
					"isbn":     "0-395-19395-8",
					"price":    22.99,
				},
			},
			"bicycle": m{
				"color": "red",
				"price": 19.95,
			},
		},
	}

	out, err := GetJSONPath("{.store.bicycle.color}", in)
	assert.NoError(t, err)
	assert.Equal(t, "red", out)

	out, err = GetJSONPath("{.store.bicycle.price}", in)
	assert.NoError(t, err)
	assert.Equal(t, 19.95, out)

	_, err = GetJSONPath("{.store.bogus}", in)
	assert.Error(t, err)

	_, err = GetJSONPath("{.store.unclosed}", in)
	assert.Error(t, err)

	out, err = GetJSONPath("{.store}", in)
	assert.NoError(t, err)
	assert.EqualValues(t, in["store"], out)

	out, err = GetJSONPath("{$.store.book[*].author}", in)
	assert.NoError(t, err)
	assert.Len(t, out, 4)
	assert.Contains(t, out, "Nigel Rees")
	assert.Contains(t, out, "Evelyn Waugh")
	assert.Contains(t, out, "Herman Melville")
	assert.Contains(t, out, "J. R. R. Tolkien")

	out, err = GetJSONPath("{$..book[?( @.price < 10.0 )]}", in)
	assert.NoError(t, err)
	expected := ar{
		m{
			"category": "reference",
			"author":   "Nigel Rees",
			"title":    "Sayings of the Century",
			"price":    8.95,
		},
		m{
			"category": "fiction",
			"author":   "Herman Melville",
			"title":    "Moby Dick",
			"isbn":     "0-553-21311-3",
			"price":    8.99,
		},
	}
	assert.EqualValues(t, expected, out)

	in = m{
		"a": m{
			"aa": m{
				"foo": m{
					"aaa": m{
						"aaaa": m{
							"bar": 1234,
						},
					},
				},
			},
			"ab": m{
				"aba": m{
					"foo": m{
						"abaa": true,
						"abab": "baz",
					},
				},
			},
		},
	}
	out, err = GetJSONPath("{..foo.*}", in)
	assert.NoError(t, err)
	assert.Len(t, out, 3)
	assert.Contains(t, out, m{"aaaa": m{"bar": 1234}})
	assert.Contains(t, out, true)
	assert.Contains(t, out, "baz")

	type bicycleType struct {
		Color string
	}
	type storeType struct {
		Bicycle *bicycleType
		safe    interface{}
	}

	structIn := &storeType{
		Bicycle: &bicycleType{
			Color: "red",
		},
		safe: "hidden",
	}

	out, err = GetJSONPath("{.Bicycle.Color}", structIn)
	assert.NoError(t, err)
	assert.Equal(t, "red", out)

	_, err = GetJSONPath("{.safe}", structIn)
	assert.Error(t, err)

	_, err = GetJSONPath("{.*}", structIn)
	assert.Error(t, err)
}
