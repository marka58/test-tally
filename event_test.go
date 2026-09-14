package testtally

import "testing"

func TestParseEvent(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		want    Event
		wantErr bool
	}{
		{
			name: "pass",
			line: `{"Action":"pass","Package":"widget","Test":"TestBasic","Elapsed":0.04}`,
			want: Event{Action: "pass", Package: "widget", Test: "TestBasic", Elapsed: 0.04},
		},
		{
			name: "fail",
			line: `{"Action":"fail","Package":"widget","Test":"TestParseConfig","Elapsed":0.12}`,
			want: Event{Action: "fail", Package: "widget", Test: "TestParseConfig", Elapsed: 0.12},
		},
		{
			name: "output event has no Test field",
			line: `{"Action":"output","Package":"widget","Output":"=== RUN   TestBasic\n"}`,
			want: Event{Action: "output", Package: "widget", Output: "=== RUN   TestBasic\n"},
		},
		{
			name: "package-level event has no Test field",
			line: `{"Action":"pass","Package":"widget","Elapsed":1.28}`,
			want: Event{Action: "pass", Package: "widget", Elapsed: 1.28},
		},
		{
			name:    "empty line",
			line:    ``,
			wantErr: true,
		},
		{
			name:    "malformed json",
			line:    `{"Action":"pass",`,
			wantErr: true,
		},
		{
			name:    "not an object",
			line:    `"just a string"`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseEvent([]byte(tc.line))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseEvent(%q) = %+v, nil; want error", tc.line, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseEvent(%q) returned error: %v", tc.line, err)
			}
			if got != tc.want {
				t.Errorf("ParseEvent(%q) = %+v, want %+v", tc.line, got, tc.want)
			}
		})
	}
}
