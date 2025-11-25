package engine

import "testing"

func TestMatcher(t *testing.T) {
	m := newMatcher([]string{"Notepad++/backup/", "*.log", "*/cache/"})

	cases := []struct {
		path   string
		isDir  bool
		expect bool
	}{
		{"Notepad++/backup", true, true},
		{"Notepad++/backup/file.txt", false, true},
		{"Notepad++/config.xml", false, false},
		{"Any/cache", true, true},
		{"Any/cache/file.txt", false, true},
		{"logs/app.log", false, true},
		{"logs/app.txt", false, false},
	}

	for _, tc := range cases {
		if got := m.ShouldSkip(tc.path, tc.isDir); got != tc.expect {
			t.Fatalf("ShouldSkip(%q, %v) = %v, want %v", tc.path, tc.isDir, got, tc.expect)
		}
	}
}
