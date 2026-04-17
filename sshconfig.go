// Command sshconfig regenerates ~/.ssh/config from a semicolon-delimited CSV.
package main

import (
	"bufio"
	"cmp"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
)

const sshDefaults = "# defaults" +
	"\nHost *" +
	"\n\tForwardAgent no" +
	"\n\tForwardX11 no" +
	"\n\tForwardX11Trusted yes" +
	"\n\tServerAliveInterval 60" +
	"\n\tServerAliveCountMax 30" +
	"\n\tIdentitiesOnly yes" +
	"\n\tCiphers chacha20-poly1305@openssh.com,aes128-ctr,aes192-ctr,aes256-ctr" +
	"\n\tCompression no"

// SSHHost represents a single SSH host entry from the CSV config.
type SSHHost struct {
	Priority     uint
	Host         string
	HostName     string
	User         string
	IdentityFile string
	Port         uint
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "sshconfig: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	configout := filepath.Join(homeDir, ".ssh", "config")
	csvfile := filepath.Join(homeDir, ".ssh", "sshconfig.csv")

	recs, err := loadCSV(csvfile)
	if err != nil {
		return fmt.Errorf("load csv: %w", err)
	}
	if err := writeCSV(csvfile, recs); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}
	if err := writeConfig(homeDir, configout, recs); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func sortRecs(crecs []SSHHost) []SSHHost {
	slices.SortFunc(crecs, func(a, b SSHHost) int {
		if n := cmp.Compare(a.HostName, b.HostName); n != 0 {
			return n
		}
		if n := cmp.Compare(a.Host, b.Host); n != 0 {
			return n
		}
		return cmp.Compare(a.Priority, b.Priority)
	})
	return crecs
}

func loadCSV(csvfile string) (_ []SSHHost, err error) {
	csvFile, err := os.Open(csvfile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := csvFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	r := csv.NewReader(csvFile)
	r.Comma = ';'
	r.FieldsPerRecord = 6
	if _, err = r.Read(); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read records: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}
	recs := make([]SSHHost, 0, len(records))
	for i, record := range records {
		priority, err := strconv.ParseUint(record[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid priority %q: %w", i+1, record[0], err)
		}
		port, err := strconv.ParseUint(record[5], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid port %q: %w", i+1, record[5], err)
		}
		recs = append(recs, SSHHost{
			Priority:     uint(priority),
			Host:         record[1],
			HostName:     record[2],
			User:         record[3],
			IdentityFile: record[4],
			Port:         uint(port),
		})
	}
	recs = sortRecs(recs)
	recs = slices.Compact(recs)
	return recs, nil
}

func writeCSV(csvfile string, recs []SSHHost) error {
	tmp, err := os.CreateTemp(filepath.Dir(csvfile), ".sshconfig-*.csv")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	var success bool
	defer func() {
		if !success {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	w := csv.NewWriter(bufio.NewWriter(tmp))
	w.Comma = ';'
	if err := w.Write([]string{"priority", "host", "hostname", "user", "identityfile", "port"}); err != nil {
		return err
	}
	for _, r := range recs {
		record := []string{
			strconv.FormatUint(uint64(r.Priority), 10),
			r.Host, r.HostName, r.User, r.IdentityFile,
			strconv.FormatUint(uint64(r.Port), 10),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	success = true
	return os.Rename(tmpName, csvfile)
}

func writeConfig(homeDir, configout string, recs []SSHHost) error {
	tmp, err := os.CreateTemp(filepath.Dir(configout), ".sshconfig-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	var success bool
	defer func() {
		if !success {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	w := bufio.NewWriter(tmp)
	if _, err = w.WriteString(sshDefaults); err != nil {
		return err
	}
	for _, h := range recs {
		if _, err = fmt.Fprintf(w, "\nHost %s\n\tHostname %s\n\tUser %s\n\tIdentityFile %s\n\tPort %d\n",
			h.Host, h.HostName, h.User,
			filepath.Join(homeDir, ".ssh", h.IdentityFile),
			h.Port); err != nil {
			return err
		}
	}
	if err = w.Flush(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	success = true
	return os.Rename(tmpName, configout)
}
