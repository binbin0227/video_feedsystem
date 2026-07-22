package db

import "testing"

func TestEscapeLikeKeyword(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		want    string
	}{
		{name: "plain text", keyword: "alice", want: "alice"},
		{name: "percent", keyword: "100%", want: "100!%"},
		{name: "underscore", keyword: "a_b", want: "a!_b"},
		{name: "escape marker", keyword: "a!b", want: "a!!b"},
		{name: "combined", keyword: "!_%", want: "!!!_!%"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := escapeLikeKeyword(test.keyword); got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}
