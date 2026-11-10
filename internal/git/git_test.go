package git

import "testing"

func TestPushRemote(t *testing.T) {
	cases := []struct {
		remotes []string
		want    string
		err     bool
	}{
		{nil, "", true},
		{[]string{"upstream"}, "upstream", false},
		{[]string{"fork", "origin"}, "origin", false},
		{[]string{"a", "b"}, "", true},
	}
	for _, c := range cases {
		got, err := pushRemote(c.remotes)
		if got != c.want || (err != nil) != c.err {
			t.Errorf("pushRemote(%v) = %q, %v", c.remotes, got, err)
		}
	}
}
