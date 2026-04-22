package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "orgs.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp CSV: %v", err)
	}
	return path
}

func TestReadOrganizationsFromCSV_HappyPath(t *testing.T) {
	path := writeTempCSV(t, "org-one\norg-two\norg-three\n")
	got, err := ReadOrganizationsFromCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"org-one", "org-two", "org-three"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadOrganizationsFromCSV_TrimsWhitespace(t *testing.T) {
	path := writeTempCSV(t, "  org-one  \n\torg-two\t\n")
	got, err := ReadOrganizationsFromCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"org-one", "org-two"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadOrganizationsFromCSV_SkipsBlankAndInvalid(t *testing.T) {
	// Blank name, names with spaces, and names with slashes should be skipped.
	path := writeTempCSV(t, "org-one\n\n   \nbad name\nbad/name\norg-two\n")
	got, err := ReadOrganizationsFromCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"org-one", "org-two"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadOrganizationsFromCSV_MultiColumnUsesFirst(t *testing.T) {
	// CSV reader requires consistent field counts, so each row must have the same
	// number of columns. Only the first column should be used as the org name.
	path := writeTempCSV(t, "org-one,note-a\norg-two,note-b\n")
	got, err := ReadOrganizationsFromCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"org-one", "org-two"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestReadOrganizationsFromCSV_MissingFile(t *testing.T) {
	_, err := ReadOrganizationsFromCSV(filepath.Join(t.TempDir(), "does-not-exist.csv"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestReadOrganizationsFromCSV_EmptyFile(t *testing.T) {
	path := writeTempCSV(t, "")
	got, err := ReadOrganizationsFromCSV(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}
