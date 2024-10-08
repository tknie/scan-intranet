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
	"time"

	"github.com/tknie/flynn"
	"github.com/tknie/flynn/common"
	"github.com/tknie/log"
)

const timeFormat = "2006-01-02 15:04:05"
const tableName = "ScanIntranet"

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

func ScanIntranet(create bool) error {

	now := time.Now()

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
				return err
			}
		}
	}
	if err != nil && err != io.EOF {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
