
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"encoding/json"
	"io/ioutil"
)

var initSleepers = flag.String("sleepy", "", "comma-sep list of initial sleepers")

func main() {
	flag.Parse()
	names := flag.Args()

	allFuncs := CallMap{}
	for _, name := range names {
		list := make([]*Sleeper, 0)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			log.Printf("err on file %v: %v", name, err)
		}

		if err := json.Unmarshal(data, &list); err != nil {
			log.Printf("err on file %v: %v", name, err)
		}

		for _, s := range list {
			allFuncs[s.Calls] = append(allFuncs[s.Calls], s.Caller)
		}
	}

	for _, s := range allFuncs.Sleepy() {
		fmt.Println(s)
	}
}

// map[callee]callers
type CallMap map[string][]string

func (m CallMap) markSleepy(name string, sleepy map[string]bool) {
	if sleepy[name] {
		return // prevent infinite recursion
	}

	sleepy[name] = true
	for _, caller := range m[name] {
		m.markSleepy(caller, sleepy)
	}
}

func (m CallMap) Sleepy() []string {
	sleepy := map[string]bool{}
	initNames := strings.Split(*initSleepers, ",")
	for _, initial := range initNames {
		m.markSleepy(initial, sleepy)
	}

	sleepyList := make([]string, 0, len(sleepy))
	for name, _ := range sleepy {
		sleepyList = append(sleepyList, name)
	}
	return sleepyList
}

func (m CallMap) String() string {
	str := ""
	for callee, callers := range m {
		str += fmt.Sprintf("%v called by %v\n", callee, callers)
	}
	return str
}


type Sleeper struct {
	Caller string `json:"caller"`
	Calls string `json:"calls"`
}

