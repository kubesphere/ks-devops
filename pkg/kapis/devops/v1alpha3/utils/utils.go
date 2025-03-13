package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/klog/v2"
)

// GetChildFileName checks if file is recursively under given dir. If true, return the child file or directory name under dir. Otherwise, return empty string.
//
// filepath is a concrete file path which should never be a directory
//
// e.g.
//
// give {"/a/b/file.txt", "/a"}, return {"b", true}
//
// give {"/a/b/file.txt", "/a/b"}, return {"file.txt", false}
//
// give {"/a/b/file.txt", "/a/c"}, return {"", false}
func GetChildFileName(filepath string, dir string) (name string, isDir bool) {
	filepath = "/" + strings.TrimPrefix(filepath, "/")
	dir = "/" + strings.TrimSuffix(strings.TrimPrefix(dir, "/"), "/")
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	if strings.HasPrefix(filepath, dir) {
		right := strings.TrimPrefix(filepath, dir)
		parts := strings.Split(right, "/")
		switch len(parts) {
		case 0:
			return "", false
		case 1:
			return parts[0], false
		default:
			return parts[0], true
		}
	} else {
		return "", false
	}
}

// IsValidFilePattern checks if a pattern is a valid file path for the git system.
func IsValidFilePattern(pattern string) bool {
	// Check for empty pattern
	if pattern == "" {
		return false
	}

	// Normalize redundant slashes
	normalized := strings.ReplaceAll(pattern, "//", "/")

	// Invalid characters regex (add more based on filesystem if needed)
	invalidChars := regexp.MustCompile(`[<>:"|?*]`)
	if invalidChars.MatchString(normalized) {
		return false
	}

	// Ensure no leading or trailing whitespace
	if strings.TrimSpace(normalized) != normalized {
		return false
	}

	// Valid pattern with Unicode support
	return true
}

// Unpack uncompresses a given tar gzip file data (tgz) to a specified directory (toDir)
func Unpack(tgz []byte, toDir string, perm os.FileMode) error {
	err := os.MkdirAll(toDir, perm)
	if err != nil {
		klog.Errorf("Failed to create directory %s: %v", toDir, err)
		return err
	}

	// Create a new reader from the tgz byte slice
	gzReader, err := gzip.NewReader(bytes.NewReader(tgz))
	if err != nil {
		klog.Errorf("Failed to create gzip reader: %v", err)
		return err
	}
	defer gzReader.Close()

	// Create a tar reader from the decompressed gzip data
	tarReader := tar.NewReader(gzReader)

	// Iterate through the files in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			klog.Errorf("Failed to read tar header: %v", err)
			return err
		}

		// Resolve the path of the file to extract
		filePath := filepath.Join(toDir, header.Name)

		// Check the type of file
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory if it doesn't exist
			if err := os.MkdirAll(filePath, os.FileMode(header.Mode)); err != nil {
				klog.Errorf("Failed to create directory %s: %v", filePath, err)
				return err
			}
		case tar.TypeReg:
			// Create the file and any necessary directories
			if err := os.MkdirAll(filepath.Dir(filePath), perm); err != nil {
				klog.Errorf("Failed to create directory for file %s: %v", filePath, err)
				return err
			}

			// Create and write to the file
			outFile, err := os.Create(filePath)
			if err != nil {
				klog.Errorf("Failed to create file %s: %v", filePath, err)
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				klog.Errorf("Failed to write to file %s: %v", filePath, err)
				return err
			}
			outFile.Close()
			klog.Infof("Successfully extracted file: %s", filePath)
		default:
			// Log other file types as skipped
			klog.Warningf("Skipping unknown type %c in %s", header.Typeflag, header.Name)
		}
	}
	klog.Info("Unpack completed successfully")
	return nil
}

// GetPage returns a paginated subset of a slice.
func GetPage[T any](slice []T, page int, perPage int) ([]T, int) {
	total := len(slice)

	// Calculate starting and ending indices for the desired page.
	start := (page - 1) * perPage
	end := start + perPage

	// If start index is out of bounds, return an empty slice.
	if start >= len(slice) || start < 0 {
		return []T{}, total
	}

	// Adjust end index if it exceeds the slice length.
	if end > len(slice) {
		end = len(slice)
	}

	// Return the paginated slice.
	return slice[start:end], total
}

// SortRefsByShortName sorts a slice of git references by their short names.
// It prioritizes "main" and "master" branches, placing them at the beginning of the list.
// All other references are sorted alphabetically by their short names.
//
// Parameters:
//   - refs: A slice of pointers to plumbing.Reference objects to be sorted.
//
// Returns:
//   - A sorted slice of pointers to plumbing.Reference objects.
func SortRefsByShortName(refs []*plumbing.Reference) []*plumbing.Reference {
	sort.Slice(refs, func(i, j int) bool {
		nameI := refs[i].Name().Short()
		nameJ := refs[j].Name().Short()

		// Check if the branch is "main" or "master"
		isMainOrMasterI := strings.ToLower(nameI) == "main" || strings.ToLower(nameI) == "master"
		isMainOrMasterJ := strings.ToLower(nameJ) == "main" || strings.ToLower(nameJ) == "master"

		// If one is "main" or "master" and the other is not, prioritize "main" or "master"
		if isMainOrMasterI && !isMainOrMasterJ {
			return true
		}
		if !isMainOrMasterI && isMainOrMasterJ {
			return false
		}

		// Otherwise, sort alphabetically
		return nameI < nameJ
	})
	return refs
}
