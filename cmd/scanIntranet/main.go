package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tknie/flynn"
	"github.com/tknie/flynn/common"
	"github.com/tknie/log"
)

const timeFormat = "2006-01-02 15:04:05"
const tableName = "ScanIntranet"
const description = "Descrption"

var insertFieldList = []string{"IP", "Hostname", "ScanTime", "State"}

type hostEntry struct {
	IP       string
	Hostname string
	ScanTime time.Time
	State    string
}

func DatabaseLocation() (*common.Reference, string, error) {
	url := os.Getenv("POSTGRES_URL")

	ref, passwd, err := common.NewReference(url)
	if err != nil {
		fmt.Println("URL error:", err)
		return nil, "", err
	}
	if passwd == "" {
		passwd = os.Getenv("POSTGRES_PASSWORD")
	}
	return ref, passwd, nil
}

func DatabaseHandler() (common.RegDbID, error) {
	ref, passwd, err := DatabaseLocation()
	if err != nil {
		return 0, err
	}
	log.Log.Debugf("Connect to %s:%d", ref.Host, ref.Port)
	id, err := flynn.Handler(ref, passwd)
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return 0, err
	}
	return id, nil
}
func main() {
	create := false
	flag.BoolVar(&create, "C", false, "Create database")
	flag.Usage = func() {
		fmt.Print(description)
		fmt.Println("Default flags:")
		flag.PrintDefaults()
	}
	flag.Parse()

	now := time.Now()
	tool := "ip"
	args := []string{"neigh"}
	cmd := exec.Command(tool, args...)
	stdout, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	id, err := DatabaseHandler()
	if err != nil {
		fmt.Println("Database connection err:", err)
		return
	}

	if create {
		entry := &hostEntry{}
		err = id.CreateTable(tableName, entry)
		if err != nil {
			fmt.Printf("Database creating failed: %v\n", err)
			return
		}
	}

	reader := bytes.NewReader(stdout)
	bufReader := bufio.NewReader(reader)
	line, complete, err := bufReader.ReadLine()
	for err == nil {
		vals := strings.Split(string(line), " ")
		state := "FAILED"
		switch {
		case len(vals) > 5:
			state = vals[5]
		case len(vals) > 2:
			state = vals[3]
		case len(vals) == 0:
			fmt.Println("Error parsing:", string(line))
			continue
		}
		if len(vals) > 0 {
			laddr, lerr := net.LookupAddr(vals[0])
			if lerr != nil {
				fmt.Println("Cannot resolv", vals[0])
			}
			log.Log.Debugf(vals[0], laddr, state, complete, now.Format(timeFormat))
			line, complete, err = bufReader.ReadLine()
			hostname := "<unresolved>"
			if len(laddr) > 0 {
				hostname = laddr[0]
			}
			record := &hostEntry{vals[0], hostname, now, state}

			insertPic := &common.Entries{Fields: insertFieldList,
				DataStruct: record,
				Values:     [][]any{{record}}}
			_, err := id.Insert(tableName, insertPic)
			if err != nil {
				fmt.Println("Record insert error:", err)
				return
			}
		}
	}
	if err != nil && err != io.EOF {
		fmt.Println(err.Error())
		return
	}

}
