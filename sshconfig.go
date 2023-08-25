package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
)

// SSHHost struct
type SSHHost struct {
	Priority     uint
	Host         string
	HostName     string
	User         string
	IdentityFile string
	Port         uint
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Errorf("error: %w", err))
	}
}

func main() {
	// file system
	homeDir, err := os.UserHomeDir()
	checkErr(err)
	configout := fmt.Sprintf("%s/.ssh/config", homeDir)
	csvfile := fmt.Sprintf("%s/.ssh/sshconfig.csv", homeDir)

	// TODO: Check to make sure the file exists
	recs := loadCSV(csvfile)
	writeCSV(csvfile, recs)
	writeConfig(configout, recs)
}

func sortRecs(crecs []SSHHost) (recs []SSHHost) {
	sort.Slice(crecs, func(i, j int) bool {
		if crecs[i].HostName != crecs[j].HostName {
			return crecs[i].HostName < crecs[j].HostName
		}
		if crecs[i].Host != crecs[j].Host {
			return crecs[i].Host < crecs[j].Host
		}
		return crecs[i].Priority < crecs[j].Priority
	})
	recs = append(recs, crecs...)
	return
}

func uniqueRecs(crecs []SSHHost) (recs []SSHHost) {
	for k := range crecs {
		if k > 0 {
			if !reflect.DeepEqual(crecs[k-1], crecs[k]) {
				recs = append(recs, crecs[k])
			}
		} else {
			recs = append(recs, crecs[k])
		}
	}
	return
}

func loadCSV(csvfile string) (recs []SSHHost) {
	csvFile, err := os.Open(csvfile)
	checkErr(err)
	r := csv.NewReader(bufio.NewReader(csvFile))
	r.Comma = ';'
	records, err := r.ReadAll()
	checkErr(err)
	for i, record := range records {
		if i != 0 {
			priority, _ := strconv.ParseUint(record[0], 10, 64)
			port, _ := strconv.ParseUint(record[5], 10, 16)
			recs = append(recs, SSHHost{
				Priority:     uint(priority),
				Host:         record[1],
				HostName:     record[2],
				User:         record[3],
				IdentityFile: record[4],
				Port:         uint(port),
			})
		}
	}
	recs = sortRecs(recs)
	recs = uniqueRecs(recs)
	return
}

func writeCSV(csvfile string, recs []SSHHost) {
	csvFile, err := os.Create(csvfile)
	checkErr(err)
	defer csvFile.Close()
	w := csv.NewWriter(bufio.NewWriter(csvFile))
	w.Comma = ';'
	defer w.Flush()
	w.Write([]string{"priority", "host", "hostname", "user", "identityfile", "port"})
	for _, r := range recs {
		record := []string{}
		record = append(record, fmt.Sprintf("%d", r.Priority), r.Host, r.HostName, r.User, r.IdentityFile, fmt.Sprintf("%d", r.Port))
		w.Write(record)
	}
}

func writeConfig(configout string, recs []SSHHost) {
	defaults := "# defaults"
	defaults += "\nHost *"
	defaults += "\n\tForwardAgent no"
	defaults += "\n\tForwardX11 no"
	defaults += "\n\tForwardX11Trusted yes"
	defaults += "\n\tProtocol 2"
	defaults += "\n\tServerAliveInterval 60"
	defaults += "\n\tServerAliveCountMax 30"
	defaults += "\n\tIdentitiesOnly yes"
	defaults += "\n\tCiphers chacha20-poly1305@openssh.com,aes128-ctr,aes192-ctr,aes256-ctr"
	defaults += "\n\tCompression no"

	f, err := os.Create(configout)
	checkErr(err)
	defer f.Close()
	f.WriteString(defaults)
	for _, h := range recs {
		f.WriteString(fmt.Sprintf("\nHost %s\n\tHostname %s\n\tUser %s\n\tIdentityFile %s\n\tPort %d", h.Host, h.HostName, h.User, h.IdentityFile, h.Port))
	}
}
