package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/nspcc-dev/netmap-ql"
	"github.com/nspcc-dev/netmap/netgraph"
	"github.com/pkg/errors"
	"gopkg.in/abiosoft/ishell.v2"
)

type state struct {
	b  *netgraph.Bucket
	rf uint32
	ss []netgraph.Select
	fs []netgraph.Filter
}

const stateKey = "state"

var (
	errWrongFormat = errors.New("wrong command format")
	defaultSource  = rand.NewSource(0)
)

var commands = []*ishell.Cmd{
	{
		Name: "get-selection",
		Help: "apply current selection rules",
		LongHelp: `Usage: get-selection
Example:
>>> load /examples/map2
>>> query SELECT 1 Country FILTER Country NE Russia
>>> get-selection
[13 14]`,
		Func: getSelection,
	},
	{
		Name:     "clear-selection",
		Help:     "clear selection rules",
		LongHelp: "Usage: clear-selection",
		Func:     clearSelection,
	},
	{
		Name:     "dump-selection",
		Help:     "dump selection result in *.dot format",
		LongHelp: "Usage: dump-selection <filename>",
		Func:     dumpNetmap,
	},
	{
		Name:     "clear",
		Help:     "clear netmap",
		LongHelp: "Usage: clear",
		Func:     clearNetmap,
	},
	{
		Name:     "load",
		Help:     "load netmap from file",
		LongHelp: "Usage: load <filename>",
		Func:     loadFromFile,
	},
	{
		Name:     "save",
		Help:     "save netmap to file",
		LongHelp: "Usage: save <filename>",
		Func:     saveToFile,
	},
	{
		Name: "add",
		Help: "add node to netmap",
		LongHelp: `Usage: add <number> /key1:value1/key2:value2 [option2 [...]]
Example:
>>> add 1 /Location:Europe/Country:Germany /Trust:10
>>> add 2 /Location:Europe/Country:Austria`,
		Func: addNode,
	},
	{
		Name: "query",
		Help: "use query placement rule",
		LongHelp: `Usage: query <STATEMENT>
Example:
>>> add 1 /Location:Europe/Country:Germany
>>> add 2 /Location:Europe/Country:Austria
>>> add 2 /Location:Asia/Country:Korea
>>> add 2 /Location:Asia/Country:Japan
>>> SELECT 1 Country FILTER Country NE Russia`,
		Func: addQuery,
	},
	{
		Name: "spew",
		Func: func(c *ishell.Context) {
			spew.Dump(getState(c))
		},
	},
}

func main() {
	var (
		st = &state{
			b:  new(netgraph.Bucket),
			ss: nil,
			fs: nil,
		}
		shell = ishell.New()
	)

	shell.Set(stateKey, st)
	for _, c := range commands {
		shell.AddCmd(c)
	}

	shell.Run()
}

func getState(c *ishell.Context) *state {
	return c.Get(stateKey).(*state)
}

func read(b *netgraph.Bucket, name string) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	return b.UnmarshalBinary(data)
}

func write(b *netgraph.Bucket, name string) error {
	data, err := b.MarshalBinary()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(name, data, os.ModePerm)
}

func getSelection(c *ishell.Context) {
	s := getState(c)
	if b := s.b.GetMaxSelection(s.ss, s.fs); b != nil {
		if b = b.GetSelection(s.ss, rand.New(defaultSource)); b != nil {
			c.Println(b.Nodelist())
			return
		}
	}
	c.Println(nil)
}

func clearSelection(c *ishell.Context) {
	s := getState(c)
	s.ss = nil
	s.fs = nil
}

func dumpNetmap(c *ishell.Context) {
	if len(c.Args) != 1 {
		c.Err(errWrongFormat)
		return
	}
	s := getState(c)
	if len(s.fs) == 0 && len(s.ss) == 0 {
		if err := s.b.Dump(c.Args[0]); err != nil {
			c.Err(err)
			return
		}
		if err := dotToPng(c.Args[0], c.Args[0]+".png"); err != nil {
			c.Err(err)
			return
		}
	}
	if b := s.b.GetMaxSelection(s.ss, s.fs); b != nil {
		if b = b.GetSelection(s.ss, rand.New(defaultSource)); b != nil {
			if err := s.b.DumpWithSelection(c.Args[0], *b); err != nil {
				c.Err(err)
				return
			}
			if err := dotToPng(c.Args[0], c.Args[0]+".png"); err != nil {
				c.Err(err)
				return
			}
		}
	}
}

func clearNetmap(c *ishell.Context) {
	s := getState(c)
	s.b = new(netgraph.Bucket)
	s.ss = nil
	s.fs = nil
	s.rf = 0
}

func loadFromFile(c *ishell.Context) {
	if len(c.Args) == 0 {
		c.Err(errWrongFormat)
		return
	}
	s := getState(c)
	if err := read(s.b, c.Args[0]); err != nil {
		c.Err(err)
	}
}

func saveToFile(c *ishell.Context) {
	if len(c.Args) == 0 {
		c.Err(errWrongFormat)
		return
	}
	s := getState(c)
	if err := write(s.b, c.Args[0]); err != nil {
		c.Err(err)
	}
}

func addNode(c *ishell.Context) {
	if len(c.Args) < 2 {
		c.Err(errWrongFormat)
		return
	}
	node, err := strconv.Atoi(c.Args[0])
	if err != nil || node < 0 {
		c.Err(err)
		return
	}
	s := getState(c)
	if err = s.b.AddNode(int32(node), c.Args[1:]...); err != nil {
		c.Err(err)
	}
}

func addQuery(c *ishell.Context) {
	raw := strings.Join(c.Args, " ")

	rule, err := query.ParseQuery(raw)
	if err != nil {
		c.Err(errors.Wrapf(err, "bad query: %s", raw))
		return
	}

	state := getState(c)
	for _, item := range rule.SFGroups {
		for _, ss := range item.Selectors {
			state.ss = append(state.ss, ss)
		}
		for _, fs := range item.Filters {
			state.fs = append(state.fs, fs)
		}
	}
	state.rf = rule.ReplFactor
}

func dotToPng(in, out string) error {
	cmd := exec.Command("dot", "-Tpng", in, "-o", out)
	return cmd.Run()
}
