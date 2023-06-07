package rabbit

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testMapForm struct {
	ID     uint    `json:"id" binding:"required"`
	Title  *string `json:"title"`
	Source *string `json:"source"`
}

func TestFormAsMap(t *testing.T) {
	title := "title"
	form := testMapForm{
		ID:    100,
		Title: &title,
	}
	{
		vals := StructAsMap(form, []string{"Title", "Target"}) // Title is valid
		assert.Equal(t, 1, len(vals))
		assert.Equal(t, title, vals["Title"])
	}
	{
		vals := StructAsMap(form, []string{"ID", "Source"}) // ID is valid
		assert.Equal(t, 1, len(vals))
		assert.Equal(t, uint(100), vals["ID"])
	}
	{
		vals := StructAsMap(&form, []string{"ID", "Title"}) // ID, Title are valid
		assert.Equal(t, 2, len(vals))
		assert.Equal(t, uint(100), vals["ID"])
		assert.Equal(t, title, vals["Title"])
	}
}

func TestRandText(t *testing.T) {
	v := RandText(10)
	assert.Equal(t, len(v), 10)

	v2 := RandNumberText(5)
	assert.Equal(t, len(v2), 5)
	_, err := strconv.ParseInt(v2, 10, 64)
	assert.Nil(t, err)
}
