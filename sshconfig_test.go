package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertSSHHosts(t *testing.T, got, want []SSHHost) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d records, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d]: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestSortRecs(t *testing.T) {
	input := []SSHHost{
		{Priority: 2, Host: "b", HostName: "z.example.com", User: "u", Port: 22},
		{Priority: 1, Host: "a", HostName: "z.example.com", User: "u", Port: 22},
		{Priority: 1, Host: "c", HostName: "a.example.com", User: "u", Port: 22},
	}
	// First by HostName, then Host, then Priority
	want := []SSHHost{
		{Priority: 1, Host: "c", HostName: "a.example.com", User: "u", Port: 22},
		{Priority: 1, Host: "a", HostName: "z.example.com", User: "u", Port: 22},
		{Priority: 2, Host: "b", HostName: "z.example.com", User: "u", Port: 22},
	}
	assertSSHHosts(t, sortRecs(input), want)
}

func TestSortRecsByPriority(t *testing.T) {
	input := []SSHHost{
		{Priority: 3, Host: "a", HostName: "x.example.com", User: "u", Port: 22},
		{Priority: 1, Host: "a", HostName: "x.example.com", User: "u", Port: 22},
		{Priority: 2, Host: "a", HostName: "x.example.com", User: "u", Port: 22},
	}
	want := []SSHHost{
		{Priority: 1, Host: "a", HostName: "x.example.com", User: "u", Port: 22},
		{Priority: 2, Host: "a", HostName: "x.example.com", User: "u", Port: 22},
		{Priority: 3, Host: "a", HostName: "x.example.com", User: "u", Port: 22},
	}
	assertSSHHosts(t, sortRecs(input), want)
}

func TestLoadCSV(t *testing.T) {
	dir := t.TempDir()
	csvfile := filepath.Join(dir, "sshconfig.csv")
	content := "priority;host;hostname;user;identityfile;port\n" +
		"1;myhost;192.168.1.1;alice;id_rsa;22\n" +
		"2;otherhost;10.0.0.1;bob;id_ed25519;2222\n"
	if err := os.WriteFile(csvfile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	recs, err := loadCSV(csvfile)
	if err != nil {
		t.Fatalf("loadCSV: unexpected error: %v", err)
	}
	// sorted by HostName: 10.0.0.1 < 192.168.1.1
	want := []SSHHost{
		{Priority: 2, Host: "otherhost", HostName: "10.0.0.1", User: "bob", IdentityFile: "id_ed25519", Port: 2222},
		{Priority: 1, Host: "myhost", HostName: "192.168.1.1", User: "alice", IdentityFile: "id_rsa", Port: 22},
	}
	assertSSHHosts(t, recs, want)
}

func TestLoadCSVMissingFile(t *testing.T) {
	_, err := loadCSV(filepath.Join(t.TempDir(), "missing.csv"))
	if err == nil {
		t.Fatal("loadCSV: expected error for missing file, got nil")
	}
}

func TestLoadCSVErrors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"invalid priority", "priority;host;hostname;user;identityfile;port\nnotanumber;myhost;192.168.1.1;alice;id_rsa;22\n"},
		{"invalid port", "priority;host;hostname;user;identityfile;port\n1;myhost;192.168.1.1;alice;id_rsa;notaport\n"},
		{"wrong columns", "priority;host;hostname;user;identityfile;port\n1;myhost;192.168.1.1\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csvfile := filepath.Join(t.TempDir(), "sshconfig.csv")
			if err := os.WriteFile(csvfile, []byte(tt.content), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadCSV(csvfile); err == nil {
				t.Fatalf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestWriteCSVRoundtrip(t *testing.T) {
	dir := t.TempDir()
	csvfile := filepath.Join(dir, "sshconfig.csv")
	content := "priority;host;hostname;user;identityfile;port\n" +
		"1;myhost;192.168.1.1;alice;id_rsa;22\n"
	if err := os.WriteFile(csvfile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	recs, err := loadCSV(csvfile)
	if err != nil {
		t.Fatalf("loadCSV: %v", err)
	}
	if err := writeCSV(csvfile, recs); err != nil {
		t.Fatalf("writeCSV: %v", err)
	}

	recs2, err := loadCSV(csvfile)
	if err != nil {
		t.Fatalf("loadCSV after writeCSV: %v", err)
	}
	assertSSHHosts(t, recs2, recs)
}

func TestWriteConfig(t *testing.T) {
	dir := t.TempDir()
	configout := filepath.Join(dir, "config")
	recs := []SSHHost{
		{Priority: 1, Host: "myhost", HostName: "192.168.1.1", User: "alice", IdentityFile: "id_rsa", Port: 22},
	}
	if err := writeConfig(dir, configout, recs); err != nil {
		t.Fatalf("writeConfig: %v", err)
	}
	data, err := os.ReadFile(configout)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	out := string(data)
	for _, want := range []string{
		"Host *",
		"IdentitiesOnly yes",
		"Host myhost",
		"User alice",
		"Port 22",
		filepath.Join(dir, ".ssh", "id_rsa"),
	} {
		if !strings.Contains(out, want) {
			t.Errorf("writeConfig: output missing %q", want)
		}
	}
}
