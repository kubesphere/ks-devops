package utils

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
)

func TestGetChildFileName(t *testing.T) {
	tests := []struct {
		filepath string
		dir      string
		expected string
		isDir    bool
	}{
		// Basic cases
		{"/a/b/file.txt", "/a", "b", true},
		{"/a/b/file.txt", "/a/b", "file.txt", false},
		{"/a/b/file.txt", "/a/c", "", false},

		// Case with trailing slashes in dir
		{"/a/b/file.txt", "/a/", "b", true},
		{"/a/b/file.txt", "/a/b/", "file.txt", false},

		// Edge case where filepath is at the root of dir
		{"/a/b", "/a", "b", false},
		{"/a/b", "/a/b", "", false}, // filepath matches dir exactly

		// Case where filepath is a subdirectory of dir
		{"/a/b/c/file.txt", "/a", "b", true},
		{"/a/b/c/file.txt", "/a/b", "c", true},
		{"/a/b/c/file.txt", "/a/b/c", "file.txt", false},

		// Cases with file directly under directory
		{"/a/file.txt", "/a", "file.txt", false},
		{"/a/file.txt", "/a/", "file.txt", false},

		// Edge cases with empty filepath and/or dir
		{"", "/a", "", false},            // empty filepath
		{"/a/b/file.txt", "", "", false}, // empty dir
		{"", "", "", false},              // both empty

		// Cases with overlapping directory names
		{"/a/bb/file.txt", "/a/b", "", false}, // directory name not exactly matching
		{"/a/bb", "/a", "bb", false},

		// Case where directory is deeper than filepath
		{"/a", "/a/b", "", false},
		{"/a/b", "/a/b/c", "", false},

		// Cases with multi-level directory checks
		{"/a/b/c/d/e.txt", "/a/b", "c", true},
		{"/a/b/c/d/e.txt", "/a/b/c", "d", true},
		{"/a/b/c/d/e.txt", "/a/b/c/d", "e.txt", false},

		// Case where file has unusual extension
		{"/a/b/.hiddenfile", "/a", "b", true},
		{"/a/b/.hiddenfile", "/a/b", ".hiddenfile", false},
	}

	for _, tt := range tests {
		t.Run(tt.filepath+"_"+tt.dir, func(t *testing.T) {
			gotName, gotIsDir := GetChildFileName(tt.filepath, tt.dir)
			if gotName != tt.expected || gotIsDir != tt.isDir {
				t.Errorf("GetChildFileName(%q, %q) = (%q, %v), want (%q, %v)", tt.filepath, tt.dir, gotName, gotIsDir, tt.expected, tt.isDir)
			}
		})
	}
}

func TestIsValidFilePattern(t *testing.T) {
	patterns := []struct {
		pattern     string
		expected    bool
		description string
	}{
		// Valid pattern with non-English characters
		{"a/文件/", true, "Pattern with Chinese characters should be valid"},
		// Valid pattern with spaces
		{"a/b/my file.txt", true, "Pattern with spaces should be valid"},
		// Valid pattern with emoji
		{"a/📁/file.txt", true, "Pattern with emoji should be valid"},
		// Invalid pattern with special character |
		{"a/b/|file.txt", false, "Pattern with invalid character | should be invalid"},
		// Empty pattern
		{"", false, "Empty pattern should be invalid"},
		// Pattern with leading/trailing whitespace
		{" a/b/", false, "Pattern with leading space should be invalid"},
		{"a/b/ ", false, "Pattern with trailing space should be invalid"},
	}

	for _, tt := range patterns {
		valid := IsValidFilePattern(tt.pattern)
		if valid != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.description, tt.expected, valid)
		}
	}
}

func TestSortRefsByShortName(t *testing.T) {
	tests := []struct {
		name     string
		input    []*plumbing.Reference
		expected []*plumbing.Reference
	}{
		{
			name: "main and master are prioritized",
			input: []*plumbing.Reference{
				plumbing.NewReferenceFromStrings("refs/heads/feature/login", ""),
				plumbing.NewReferenceFromStrings("refs/heads/main", ""),
				plumbing.NewReferenceFromStrings("refs/heads/bugfix/issue-123", ""),
				plumbing.NewReferenceFromStrings("refs/heads/master", ""),
			},
			expected: []*plumbing.Reference{
				plumbing.NewReferenceFromStrings("refs/heads/main", ""),
				plumbing.NewReferenceFromStrings("refs/heads/master", ""),
				plumbing.NewReferenceFromStrings("refs/heads/bugfix/issue-123", ""),
				plumbing.NewReferenceFromStrings("refs/heads/feature/login", ""),
			},
		},
		{
			name: "alphabetical sorting when main/master is not present",
			input: []*plumbing.Reference{
				plumbing.NewReferenceFromStrings("refs/heads/feature/login", ""),
				plumbing.NewReferenceFromStrings("refs/heads/develop", ""),
				plumbing.NewReferenceFromStrings("refs/heads/bugfix/issue-123", ""),
			},
			expected: []*plumbing.Reference{
				plumbing.NewReferenceFromStrings("refs/heads/bugfix/issue-123", ""),
				plumbing.NewReferenceFromStrings("refs/heads/develop", ""),
				plumbing.NewReferenceFromStrings("refs/heads/feature/login", ""),
			},
		},
		{
			name: "case insensitivity for main/master",
			input: []*plumbing.Reference{
				plumbing.NewReferenceFromStrings("refs/heads/MAIN", ""),
				plumbing.NewReferenceFromStrings("refs/heads/feature/login", ""),
				plumbing.NewReferenceFromStrings("refs/heads/Master", ""),
			},
			expected: []*plumbing.Reference{
				plumbing.NewReferenceFromStrings("refs/heads/MAIN", ""),
				plumbing.NewReferenceFromStrings("refs/heads/Master", ""),
				plumbing.NewReferenceFromStrings("refs/heads/feature/login", ""),
			},
		},
		{
			name:     "empty slice",
			input:    []*plumbing.Reference{},
			expected: []*plumbing.Reference{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortRefsByShortName(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
			}

			for i, ref := range result {
				if ref.Name().Short() != tt.expected[i].Name().Short() {
					t.Errorf("at index %d: expected %s, got %s", i, tt.expected[i].Name().Short(), ref.Name().Short())
				}
			}
		})
	}
}
