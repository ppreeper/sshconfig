package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	// "github.com/jmoiron/sqlx"
	// _ "github.com/mattn/go-sqlite3"
)

// SSHHost struct
type CSVRecord struct {
	Priority     uint   `db:"priority"`
	Host         string `db:"host"`
	HostName     string `db:"host_name"`
	User         string `db:"user"`
	IdentityFile string `db:"identity_file"`
	Port         uint   `db:"port"`
}

type SSHRecord struct {
	Host         string
	User         string
	IdentityFile string
	Port         int
}

var SSHHost map[string][]SSHRecord

func checkErr(err error) {
	if err != nil {
		panic(fmt.Errorf("Error: %w", err))
	}
}

func main() {
	// file system
	// homeDir, err := os.UserHomeDir()
	configDir, err := os.UserHomeDir()
	checkErr(err)
	// configout := fmt.Sprintf("%s/.ssh/config", homeDir)
	configfile := fmt.Sprintf("%s/.ssh/sshconfig.csv", configDir)

	// open database connection
	// sdb, err := OpenDatabase("sqlite3", ":memory:")
	// checkErr(err)
	// defer sdb.Close()

	// TODO: Check to make sure the file exists
	// cleanup csv and reload database
	// sdb.cleanCSV(configfile)
	// write ssh configfile
	// sdb.writeConfig(configout)

	// db.loadDB(configfile)
	loadCSV(configfile)
}

// func (db *Database) initDB() {
// 	// fmt.Println("initDB")
// 	db.Exec("DROP TABLE IF EXISTS sshconfig;")
// 	db.Exec(`CREATE TABLE sshconfig (
// 		priority int,
// 		host text,
// 		host_name text,
// 		user text,
// 		identity_file text,
// 		port int
// 		);`)
// 	return
// }

// func (db *Database) cleanCSV(configfile string) {
// 	db.initDB()
// 	db.loadDB(configfile)
//
// 	q := `SELECT distinct priority,host,host_name,user,identity_file,port from sshconfig order by host_name,priority`
// 	hosts := []SSHHost{}
// 	err := db.Select(&hosts, q)
// 	checkErr(err)
//
// 	f, err := os.Create(configfile)
// 	checkErr(err)
// 	defer f.Close()
// 	f.WriteString(fmt.Sprintf("\"priority\";\"host\";\"host_name\";\"user\";\"identity_file\";\"port\"\n"))
// 	for _, v := range hosts {
// 		// fmt.Println(v)
// 		f.WriteString(fmt.Sprintf("\"%d\";\"%s\";\"%s\";\"%s\";\"%s\";\"%d\"\n", v.Priority, v.Host, v.HostName, v.User, v.IdentityFile, v.Port))
// 	}
//
// 	db.initDB()
// 	db.loadDB(configfile)
// 	return
// }

// record
// Host nvgo
//
//	HostName 10.10.10.9
//	User sysuser
//	IdentityFile /home/sysuser/.ssh/id_rsa
//	Port 22
//
// Host nvpy
//
//	HostName 10.10.10.10
//	User sysuser
//	IdentityFile /home/sysuser/.ssh/id_rsa
//	Port 22
//
// Host nvpy
//
//	HostName 10.10.10.10
//	User root
//	IdentityFile /home/sysuser/.ssh/id_rsa
//	Port 22

func loadCSV(configfile string) {
	// Load db
	csvFile, err := os.Open(configfile)
	checkErr(err)
	r := csv.NewReader(bufio.NewReader(csvFile))
	r.Comma = ';'
	records, err := r.ReadAll()
	checkErr(err)
	var hostRecs []CSVRecord
	for k, record := range records {
		var hostRec CSVRecord
		if k != 0 {
			priority, _ := strconv.ParseUint(record[0], 10, 64)
			port, _ := strconv.ParseUint(record[5], 10, 16)

			hostRec.Priority = uint(priority)
			hostRec.Host = record[1]
			hostRec.HostName = record[2]
			hostRec.User = record[3]
			hostRec.IdentityFile = record[4]
			hostRec.Port = uint(port)
			hostRecs = append(hostRecs, hostRec)
		}
	}
	fmt.Println(hostRecs)
	//    order by
	// q := `SELECT distinct priority,host,host_name,user,identity_file,port
	//    from sshconfig order by host_name,priority`

}

// func (db *Database) loadDB(configfile string) {
// 	// Load db
// 	csvFile, err := os.Open(configfile)
// 	checkErr(err)
// 	r := csv.NewReader(bufio.NewReader(csvFile))
// 	r.Comma = ';'
// 	records, err := r.ReadAll()
// 	checkErr(err)
// 	var hostRec SSHHost
// 	for k, record := range records {
// 		if k != 0 {
//
// 			priority, _ := strconv.ParseUint(record[0], 10, 64)
// 			port, _ := strconv.ParseUint(record[5], 10, 16)
//
// 			hostRec.Priority = uint(priority)
// 			hostRec.Host = record[1]
// 			hostRec.HostName = record[2]
// 			hostRec.User = record[3]
// 			hostRec.IdentityFile = record[4]
// 			hostRec.Port = uint(port)
//
// 			// Insert Record
// 			_, err = db.NamedExec(`INSERT INTO sshconfig VALUES (:priority,:host,:host_name,:user,:identity_file,:port)`, hostRec)
// 			checkErr(err)
// 		}
// 	}
// 	return
// }

// func (db *Database) writeConfig(configout string) {
// 	f, err := os.Create(configout)
// 	checkErr(err)
// 	defer f.Close()
//
// 	q := `SELECT distinct priority,host,host_name,user,identity_file,port from sshconfig order by host_name,priority`
// 	hosts := []SSHHost{}
// 	err = db.Select(&hosts, q)
// 	checkErr(err)
//
// 	f.WriteString(configDefaults())
//
// 	homeDir, err := os.UserHomeDir()
// 	for _, h := range hosts {
// 		f.WriteString(fmt.Sprintf("\nHost %s", h.Host))
// 		f.WriteString(fmt.Sprintf("\n\tHostName %s", h.HostName))
// 		if h.User != "" {
// 			f.WriteString(fmt.Sprintf("\n\tUser %s", h.User))
// 		}
// 		if h.IdentityFile != "" {
// 			f.WriteString(fmt.Sprintf("\n\tIdentityFile %s/.ssh/%s", homeDir, h.IdentityFile))
// 		}
// 		f.WriteString(fmt.Sprintf("\n\tPort %d", h.Port))
// 	}
// 	return
// }

func configDefaults() (defaults string) {
	defaults = fmt.Sprintf("# defaults")
	defaults += fmt.Sprintf("\nHost *")
	defaults += fmt.Sprintf("\n\tForwardAgent no")
	defaults += fmt.Sprintf("\n\tForwardX11 no")
	defaults += fmt.Sprintf("\n\tForwardX11Trusted yes")
	defaults += fmt.Sprintf("\n\tProtocol 2")
	defaults += fmt.Sprintf("\n\tServerAliveInterval 60")
	defaults += fmt.Sprintf("\n\tServerAliveCountMax 30")
	defaults += fmt.Sprintf("\n\tIdentitiesOnly yes")
	defaults += fmt.Sprintf("\n\tCiphers chacha20-poly1305@openssh.com,aes128-ctr,aes192-ctr,aes256-ctr")
	defaults += fmt.Sprintf("\n\tCompression no")
	return
}
