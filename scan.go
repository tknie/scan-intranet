/*
* Copyright © 2024 private, Darmstadt, Germany and/or its licensors
*
* SPDX-License-Identifier: Apache-2.0
*
*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at
*
*       http://www.apache.org/licenses/LICENSE-2.0
*
*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
*
 */
package scanintranet

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/tknie/flynn"
	"github.com/tknie/flynn/common"
	"github.com/tknie/log"
)

const timeFormat = "2006-01-02 15:04:05"
const tableName = "ScanIntranet"

var insertFieldList = []string{"IP", "MACADDR", "Hostname",
	"ScanTime", "State", "Ping", "Generateon"}

var stateMap = make(map[string]bool)

var wg sync.WaitGroup

type hostEntry struct {
	IP         string
	MacAddr    string
	Hostname   string
	ScanTime   time.Time
	State      string
	Ping       bool
	GenerateOn string
}

var myHostname = "<unresolved>"

func init() {
	myHostname, _ = os.Hostname()
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

func callIPCommand() ([]byte, error) {
	tool := "ip"
	args := []string{"neigh"}
	cmd := exec.Command(tool, args...)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return stdout, nil
}

func ScanIntranet(create, usePingCmd bool) error {

	now := time.Now()

	wg.Add(1)
	go fping("192.168.178.0/24")

	// Call IP command
	stdout, err := callIPCommand()
	if err != nil {
		return err
	}

	// Get database handler id
	id, err := DatabaseHandler()
	if err != nil {
		fmt.Println("Database connection err:", err)
		return err
	}

	if create {
		entry := &hostEntry{}
		err = id.CreateTable(tableName, entry)
		if err != nil {
			fmt.Printf("Database creating failed: %v\n", err)
			return err
		}
	}

	wg.Wait()
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
			if usePingCmd {
				lerr = ping(vals[0])
				if lerr != nil {
					fmt.Println("Cannot ping", vals[0])
				}
			}
			pingState := false
			if ok, ps := stateMap[vals[0]]; ok {
				pingState = ps
			}
			// fmt.Println(vals[0], stateMap[vals[0]], pingState)
			log.Log.Debugf(vals[0], laddr, state, complete, now.Format(timeFormat))
			line, complete, err = bufReader.ReadLine()
			hostname := "<unresolved>"
			if len(laddr) > 0 {
				hostname = laddr[0]
				record := &hostEntry{vals[0], vals[4], hostname, now,
					state, pingState, myHostname}
				// fmt.Println(record.IP, record.Ping)

				insertPic := &common.Entries{Fields: insertFieldList,
					DataStruct: record,
					Values:     [][]any{{record}}}
				_, err := id.Insert(tableName, insertPic)
				if err != nil {
					fmt.Println("Record insert error:", err)
					return err
				}
			}
		}
	}
	if err != nil && err != io.EOF {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func ping(host string) error {
	out, _ := exec.Command("ping", host, "-c 2", "-i 3", "-w 10").Output()
	if strings.Contains(string(out), "Destination Host Unreachable") {
		fmt.Println("TANGO DOWN")
	} else {
		fmt.Println("IT'S ALIVEEE")
	}

	return nil
}

func fping(network string) error {
	fmt.Println("Call fping ....", network)
	cmd := exec.Command("fping", "-g", network, "-r", "1")
	stdout, _ := cmd.Output()
	// fmt.Println("STDOUT:" + string(stdout))
	defer wg.Done()

	reader := bytes.NewReader(stdout)
	bufReader := bufio.NewReader(reader)
	line, complete, err := bufReader.ReadLine()
	for err == nil {
		vals := strings.Split(string(line), " ")
		log.Log.Debugf("%v -> %v", complete, vals)
		switch vals[2] {
		case "alive":
			stateMap[vals[0]] = true
		case "unreachable":
			stateMap[vals[0]] = false
		default:
			fmt.Println("Unknown output", vals)
		}
		line, complete, err = bufReader.ReadLine()
	}
	fmt.Println("Ended fping", stateMap["192.168.178.131"])
	return nil
}
