package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringInSliceFound(t *testing.T) {

	var slice = []string{"abc", "def", "ghi"}

	for _, s := range slice {
		found := StringInSlice(s, slice)
		assert.True(t, found, fmt.Sprintf("String %#v not found in slice %#v", s, slice))
	}
}

func TestStringInSliceNotFound(t *testing.T) {

	var slice = []string{"abc", "def", "ghi"}
	var other = []string{"", "xyz"}

	for _, s := range other {
		found := StringInSlice(s, slice)
		assert.False(t, found, fmt.Sprintf("String %#v found in slice %#v", s, slice))
	}

}
