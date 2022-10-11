package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPaginator(t *testing.T) {
	p := newPaginator(0, 100)

	data := p.MustEncode()

	p2, err := parsePage(data)
	assert.NoError(t, err)

	assert.Equal(t, p.Offset, p2.Offset)
	assert.Equal(t, p.Size, p2.Size)
}
