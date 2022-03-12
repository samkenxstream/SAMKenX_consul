//go:build consulent
// +build consulent

package acl

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEqualPartitions(t *testing.T) {
	type testcase struct {
		a, b   string
		expect bool
	}
	cases := []testcase{
		{"", "", true},
		{"default", "", true},
		{"DeFaUlT", "", true},
		{"dEfAuLt", "", true},
		{"DeFaUlT", "dEfAuLt", true},
		{"", "foo", false},
		{"foo", "foo", true},
		{"FoO", "fOo", true},
		{"bar", "foo", false},
	}
	for _, tc := range cases {
		t.Run(tc.a+" eq? "+tc.b, func(t *testing.T) {
			assert.Equal(t, tc.expect, EqualPartitions(tc.a, tc.b))
			assert.Equal(t, tc.expect, EqualPartitions(tc.b, tc.a))
		})
	}
}

func TestIsDefaultPartition(t *testing.T) {
	assert.True(t, IsDefaultPartition(""))
	assert.True(t, IsDefaultPartition("default"))
	assert.True(t, IsDefaultPartition("DeFaUlT"))
	assert.False(t, IsDefaultPartition("foo"))
	assert.False(t, IsDefaultPartition("bar"))
}
