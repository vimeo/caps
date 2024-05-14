package webvtt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVTTDetection(t *testing.T) {
	assert.True(t, Reader{}.Detect(SampleVTT))
	assert.False(t, Reader{}.Detect(InvalidVTT1))
	assert.False(t, Reader{}.Detect(InvalidVTT2))
}

func TestVTTtoVTT(t *testing.T) {
	type test struct {
		input    []byte
		expected []byte
	}
	tests := []test{
		// TODO: add tests for utf8 and unicode
		{input: SampleVTT, expected: SampleVTT},
	}
	for _, test := range tests {
		captions, err := NewReader(false).Read(test.input)
		assert.Nil(t, err)
		result, _ := NewWriter().Write(captions)
		assert.Equal(t, string(test.expected), string(result))
	}
}
