package exploration

import "testing"

func Test_ServerFromStatusURI(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "https://sportsbots.xyz/users/Blitz_Burgh/statuses/1618752914499182593",
			out: "https://sportsbots.xyz",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			out, err := ServerFromStatusURI(tc.in)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if out != tc.out {
				t.Errorf("invalid out, expected %v got %v", tc.out, out)
			}
		})
	}
}
