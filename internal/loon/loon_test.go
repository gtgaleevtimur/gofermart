package loon

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoon(t *testing.T) {
	tests := []struct {
		name   string
		number string
		flag   bool
	}{
		{
			name:   "Negative",
			number: "4561 2612 1234 5464",
			flag:   false,
		},
		{
			name:   "Positive",
			number: "4561 2612 1234 5467",
			flag:   true,
		},
	}
	for _, tt := range tests {
		t.Run(t.Name(), func(t *testing.T) {
			res := IsValid(tt.number)
			require.Equal(t, tt.flag, res)
		})
	}
}
