package unarchive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnarchive(t *testing.T) {
	tests := []string{"zip", "tar", "tar.gz"}
	for _, extension := range tests {
		t.Run(extension, func(t *testing.T) {
			// Create temp directory
			tmpDir, createTempDirCallback := createTempDirWithCallbackAndAssert(t)
			defer createTempDirCallback()
			// Run unarchive on archive created on Unix
			err := runUnarchive(t, "unix."+extension, "archives", filepath.Join(tmpDir, "unix"))
			assert.NoError(t, err)
			assert.FileExists(t, filepath.Join(tmpDir, "unix", "link"))
			assert.FileExists(t, filepath.Join(tmpDir, "unix", "dir", "file"))

			// Run unarchive on archive created on Windows
			err = runUnarchive(t, "win."+extension, "archives", filepath.Join(tmpDir, "win"))
			assert.NoError(t, err)
			assert.FileExists(t, filepath.Join(tmpDir, "win", "link.lnk"))
			assert.FileExists(t, filepath.Join(tmpDir, "win", "dir", "file.txt"))
		})
	}
}

var unarchiveSymlinksCases = []struct {
	prefix        string
	expectedFiles []string
}{
	{prefix: "softlink-rel", expectedFiles: []string{filepath.Join("softlink-rel", "a", "softlink-rel"), filepath.Join("softlink-rel", "b", "c", "d", "file")}},
	{prefix: "softlink-cousin", expectedFiles: []string{filepath.Join("a", "b", "softlink-cousin"), filepath.Join("a", "c", "d")}},
	{prefix: "softlink-uncle-file", expectedFiles: []string{filepath.Join("a", "b", "softlink-uncle"), filepath.Join("a", "c")}},
}

func TestUnarchiveSymlink(t *testing.T) {
	testExtensions := []string{"zip", "tar", "tar.gz"}
	for _, extension := range testExtensions {
		t.Run(extension, func(t *testing.T) {
			for _, testCase := range unarchiveSymlinksCases {
				t.Run(testCase.prefix, func(t *testing.T) {
					// Create temp directory
					tmpDir, createTempDirCallback := createTempDirWithCallbackAndAssert(t)
					defer createTempDirCallback()

					// Run unarchive
					err := runUnarchive(t, testCase.prefix+"."+extension, "archives", tmpDir)
					assert.NoError(t, err)

					// Assert the all expected files were extracted
					for _, expectedFiles := range testCase.expectedFiles {
						assert.FileExists(t, filepath.Join(tmpDir, expectedFiles))
					}
				})
			}
		})
	}
}

func TestUnarchiveZipSlip(t *testing.T) {
	tests := []struct {
		testType    string
		archives    []string
		errorSuffix string
	}{
		{"rel", []string{"zip", "tar", "tar.gz"}, "illegal path in archive: '../file'"},
		{"abs", []string{"tar", "tar.gz"}, "illegal path in archive: '/tmp/bla/file'"},
		{"softlink-abs", []string{"zip", "tar", "tar.gz"}, "illegal link path in archive: '/tmp/bla/file'"},
		{"softlink-rel", []string{"zip", "tar", "tar.gz"}, "illegal link path in archive: '../../file'"},
		{"softlink-loop", []string{"tar"}, "a link can't lead to an ancestor directory"},
		{"softlink-uncle", []string{"zip", "tar", "tar.gz"}, "a link can't lead to an ancestor directory"},
		{"hardlink-tilde", []string{"tar", "tar.gz"}, "walking hardlink: illegal link path in archive: '~/../../../../../../../../../Users/Shared/sharedFile.txt'"},
	}
	for _, test := range tests {
		t.Run(test.testType, func(t *testing.T) {
			// Create temp directory
			tmpDir, createTempDirCallback := createTempDirWithCallbackAndAssert(t)
			defer createTempDirCallback()
			for _, archive := range test.archives {
				// Unarchive and make sure an error returns
				err := runUnarchive(t, test.testType+"."+archive, "zipslip", tmpDir)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.errorSuffix)
			}
		})
	}
}

func runUnarchive(t *testing.T, archiveFileName, sourceDir, targetDir string) error {
	uarchiver := Unarchiver{}
	archivePath := filepath.Join("testdata", sourceDir, archiveFileName)
	assert.True(t, uarchiver.IsSupportedArchive(archivePath))
	return uarchiver.Unarchive(filepath.Join("testdata", sourceDir, archiveFileName), archiveFileName, targetDir)
}

func createTempDirWithCallbackAndAssert(t *testing.T) (string, func()) {
	tempDirPath, err := os.MkdirTemp("", "archiver_test")
	assert.NoError(t, err, "Couldn't create temp dir")
	return tempDirPath, func() {
		assert.NoError(t, os.RemoveAll(tempDirPath), "Couldn't remove temp dir")
	}
}
