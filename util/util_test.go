package util

import (
	"github.com/4thel00z/pcopy/test"
	"os"
	"testing"
	"time"
)

func TestExpandHome_WithTilde(t *testing.T) {
	test.StrEquals(t, os.Getenv("HOME")+"/this/is/a/path", ExpandHome("~/this/is/a/path"))
}

func TestExpandHome_NoTilde(t *testing.T) {
	test.StrEquals(t, "/this/is/an/absolute/path", ExpandHome("/this/is/an/absolute/path"))
}

func TestCollapseHome_HasHomePrefix(t *testing.T) {
	test.StrEquals(t, "~/this/is/a/path", CollapseHome(os.Getenv("HOME")+"/this/is/a/path"))
}

func TestCollapseHome_NoHomePrefix(t *testing.T) {
	test.StrEquals(t, "/this/is/an/absolute/path", CollapseHome("/this/is/an/absolute/path"))
}

func TestBytesToHuman_Small(t *testing.T) {
	test.StrEquals(t, "10 B", BytesToHuman(10))
}

func TestBytesToHuman_Large(t *testing.T) {
	test.StrEquals(t, "10.1 MB", BytesToHuman(10590617))
}

func TestCommonPrefix_Empty(t *testing.T) {
	var paths []string
	test.StrEquals(t, "", commonPrefix(paths))
}

func TestCommonPrefix_1(t *testing.T) {
	paths := []string{
		"/home/phil/code/pcopy/go.mod",
		"/home/phil/code/pcopy/go.sum",
	}
	test.StrEquals(t, "/home/phil/code/pcopy", commonPrefix(paths))
}

func TestCommonPrefix_2(t *testing.T) {
	paths := []string{
		"/home/phil/code/pcopy/go.mod",
		"/home/phil/file.txt",
	}
	test.StrEquals(t, "/home/phil", commonPrefix(paths))
}

func TestCommonPrefix_NoCommonPrefix(t *testing.T) {
	paths := []string{
		"/home/phil/code/pcopy/go.mod",
		"/etc/file.txt",
	}
	test.StrEquals(t, "", commonPrefix(paths))
}

func TestCommonPrefix_SingleFile(t *testing.T) {
	paths := []string{
		"/home/phil/code/pcopy",
	}
	test.StrEquals(t, "/home/phil/code/pcopy", commonPrefix(paths))
}

func TestRelativizePaths_Empty(t *testing.T) {
	var files []string
	baseDir, relativeFiles, err := relativizeFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	test.StrEquals(t, "", baseDir)
	test.Int64Equals(t, 0, int64(len(relativeFiles)))
}

func TestRelativizePaths_AbsFilesOnly(t *testing.T) {
	files := []string{
		"/home/phil/code/pcopy/go.mod",
		"/home/phil/code/pcopy/go.sum",
		"/home/phil/code/fsdup/main.go",
	}
	baseDir, relativeFiles, err := relativizeFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	test.StrEquals(t, "/home/phil/code", baseDir)
	test.StrEquals(t, "pcopy/go.mod", relativeFiles[0])
	test.StrEquals(t, "pcopy/go.sum", relativeFiles[1])
	test.StrEquals(t, "fsdup/main.go", relativeFiles[2])
}

func TestRelativizePaths_AbsFilesNoCommonPrefix(t *testing.T) {
	files := []string{
		"/home/phil/code/pcopy/go.mod",
		"/etc/file.txt",
	}
	baseDir, relativeFiles, err := relativizeFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	test.StrEquals(t, "", baseDir)
	test.StrEquals(t, "home/phil/code/pcopy/go.mod", relativeFiles[0])
	test.StrEquals(t, "etc/file.txt", relativeFiles[1])
}

func TestRelativizePaths_OnlyRelFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	files := []string{
		"some/file.txt",
		"other/file2.txt",
		"file3.txt",
	}
	baseDir, relativeFiles, err := relativizeFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	test.StrEquals(t, tmpDir, baseDir)
	test.StrEquals(t, "some/file.txt", relativeFiles[0])
	test.StrEquals(t, "other/file2.txt", relativeFiles[1])
	test.StrEquals(t, "file3.txt", relativeFiles[2])
}

func TestRelativizePaths_RelAndAbsFiles(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	files := []string{
		"some/file.txt",
		"other/file2.txt",
		"/etc/pcopy/server.conf",
	}
	baseDir, relativeFiles, err := relativizeFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	test.StrEquals(t, "", baseDir)
	test.StrEquals(t, tmpDir[1:]+"/some/file.txt", relativeFiles[0])
	test.StrEquals(t, tmpDir[1:]+"/other/file2.txt", relativeFiles[1])
	test.StrEquals(t, "etc/pcopy/server.conf", relativeFiles[2])
}

func TestRelativizePaths_SingleRelFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	files := []string{
		"dir/file.txt",
	}
	baseDir, relativeFiles, err := relativizeFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	test.StrEquals(t, tmpDir+"/dir", baseDir)
	test.StrEquals(t, "file.txt", relativeFiles[0])
}

func TestDurationToHuman_SevenDays(t *testing.T) {
	d := 7 * 24 * time.Hour
	test.StrEquals(t, "7d", DurationToHuman(d))
}

func TestDurationToHuman_MoreThanOneDay(t *testing.T) {
	d := 49 * time.Hour
	test.StrEquals(t, "2d1h", DurationToHuman(d))
}

func TestDurationToHuman_LessThanOneDay(t *testing.T) {
	d := 17*time.Hour + 15*time.Minute
	test.StrEquals(t, "17h15m", DurationToHuman(d))
}

func TestDurationToHuman_TenOfThings(t *testing.T) {
	d := 10*time.Hour + 10*time.Minute + 10*time.Second
	test.StrEquals(t, "10h10m10s", DurationToHuman(d))
}

func TestDurationToHuman_Zero(t *testing.T) {
	test.StrEquals(t, "0", DurationToHuman(0))
}

func TestParseDuration_ZeroSuccess(t *testing.T) {
	d, err := ParseDuration("0")
	if err != nil {
		t.Fatal(err)
	}
	if d != 0 {
		t.Fatalf("expected %d, got %d", 0, d)
	}
}

func TestParseDuration_SecondsOnly(t *testing.T) {
	d, err := ParseDuration("3600")
	if err != nil {
		t.Fatal(err)
	}
	if d != time.Hour {
		t.Fatalf("expected %d, got %d", time.Hour, d)
	}
}

func TestParseDuration_WithDaysSuccess(t *testing.T) {
	d, err := ParseDuration("10d")
	if err != nil {
		t.Fatal(err)
	}
	if d != 10*24*time.Hour {
		t.Fatalf("expected %d, got %d", 10*24*time.Hour, d)
	}
}

func TestParseDuration_WithoutDaysSuccess(t *testing.T) {
	d, err := ParseDuration("10h5m")
	if err != nil {
		t.Fatal(err)
	}
	if d != 10*time.Hour+5*time.Minute {
		t.Fatalf("expected %d, got %d", 10*time.Hour+5*time.Minute, d)
	}
}

func TestParseDuration_WithDaysAndHoursFailure(t *testing.T) {
	_, err := ParseDuration("10d1h") // not supported
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}

func TestParseDuration_WithWeeksSuccess(t *testing.T) {
	d, err := ParseDuration("2w")
	if err != nil {
		t.Fatal(err)
	}
	if d != 2*7*24*time.Hour {
		t.Fatalf("expected %d, got %d", 2*7*24*time.Hour, d)
	}
}

func TestParseDuration_WithMonthsSuccess(t *testing.T) {
	d, err := ParseDuration("2mo")
	if err != nil {
		t.Fatal(err)
	}
	if d != 2*30*24*time.Hour {
		t.Fatalf("expected %d, got %d", 2*30*24*time.Hour, d)
	}
}

func TestParseDuration_WithYearsSuccess(t *testing.T) {
	d, err := ParseDuration("2y")
	if err != nil {
		t.Fatal(err)
	}
	if d != 2*365*24*time.Hour {
		t.Fatalf("expected %d, got %d", 2*365*24*time.Hour, d)
	}
}

func TestParseSize_10GSuccess(t *testing.T) {
	s, err := ParseSize("10G")
	if err != nil {
		t.Fatal(err)
	}
	test.Int64Equals(t, 10*1024*1024*1024, s)
}

func TestParseSize_10MUpperCaseSuccess(t *testing.T) {
	s, err := ParseSize("10M")
	if err != nil {
		t.Fatal(err)
	}
	test.Int64Equals(t, 10*1024*1024, s)
}

func TestParseSize_10kLowerCaseSuccess(t *testing.T) {
	s, err := ParseSize("10k")
	if err != nil {
		t.Fatal(err)
	}
	test.Int64Equals(t, 10*1024, s)
}

func TestParseSize_FailureInvalid(t *testing.T) {
	_, err := ParseSize("not a size")
	if err == nil {
		t.Fatalf("expected error, but got none")
	}
}
