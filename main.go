
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
var chain = flag.Bool("chains", false, "show the sleepy chain for every function instead of normal output")

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

	if *chain {
		for name, sleepChain := range sleepy {
			fmt.Printf("%v: %v\n", name, sleepChain)
		}
	} else if *ocaml {
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

func (m CallMap) Sleepy(initNames []string) map[string][]string {
	sleepy := map[string][]string{}
	for _, initial := range initNames {
		m.markSleepy(initial, "", sleepy)
	}
	return sleepy
}

func (m CallMap) markSleepy(caller, callee string, sleepy map[string][]string) {
	if len(sleepy[caller]) > 0 {
		return // prevent infinite recursion
	}

	if len(sleepy[callee]) > 0 {
		sleepy[caller] = append([]string{caller}, sleepy[callee]...)
	} else {
		sleepy[caller] = []string{caller}
	}

	for _, newCaller := range m[caller] {
		m.markSleepy(newCaller, caller, sleepy)
	}
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

