
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"bytes"
	"encoding/json"
	"io/ioutil"
)

var initSleepers = flag.String("sleepy", "", "comma-sep list of initial sleepers")
var preprocess = flag.Bool("preprocess", false, "clean up the input files first")
var ocaml = flag.Bool("ocaml-out", false, "format output as ocaml list")

func main() {
	flag.Parse()
	names := flag.Args()


	if len(*initSleepers) == 0 {
		log.Fatal("No initial sleepers provided.")
	}

	allFuncs := CallMap{}
	for _, name := range names {
		list := make([]*Sleeper, 0)
		data, err := ioutil.ReadFile(name)

		if *preprocess {
			data = cleanup(data)
		}

		if err != nil {
			log.Fatalf("err: %v", err)
		} else if err := json.Unmarshal(data, &list); err != nil {
			log.Fatalf("err on file '%v': %v", name, err)
		}

		for _, s := range list {
			allFuncs[s.Calls] = append(allFuncs[s.Calls], s.Caller)
		}
	}

	initNames := strings.Split(*initSleepers, ",")
	sleepy := allFuncs.Sleepy(initNames)
	if *ocaml {
		var buf bytes.Buffer
		for _, s := range sleepy {
			fmt.Fprintf(&buf, "\"%v\";\n", s)
		}
		buf.Truncate(buf.Len() - 2)
		fmt.Printf("[%s]", buf.Bytes())
	} else {
		for _, s := range sleepy {
			fmt.Println(s)
		}
	}
}

func cleanup(data []byte) []byte {
	data = bytes.Trim(data, " \t\r\n,][")
	data = append([]byte("["), data...)
	return append(data, byte(']'))
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

func (m CallMap) Sleepy(initNames []string) []string {
	sleepy := map[string]bool{}
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

