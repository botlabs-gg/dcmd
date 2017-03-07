package dcmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIntArg(t *testing.T) {
	part := "123"
	expected := int64(123)

	assert.True(t, Int.Matches(part), "Should match")

	v, err := Int.Parse(part, nil)
	assert.NoError(t, err, "Should parse sucessfully")
	assert.Equal(t, v, expected, "Should be equal")

	assert.False(t, Int.Matches("12hello21"), "Should not match")
}

func TestFloatArg(t *testing.T) {
	part := "12.3"
	expected := float64(12.3)

	assert.True(t, Float.Matches(part), "Should match")

	v, err := Float.Parse(part, nil)
	assert.NoError(t, err, "Should parse sucessfully")
	assert.Equal(t, v, expected, "Should be equal")

	assert.False(t, Float.Matches("1.2hello21"), "Should not match")
}
