package query

import (
	"unicode"

	"github.com/nspcc-dev/netmap"
	"github.com/pkg/errors"
	"github.com/vito/go-parse"
)

var (
	ws      = parsec.Many(parsec.Satisfy(unicode.IsSpace))
	ws1     = parsec.Many1(parsec.Satisfy(unicode.IsSpace))
	strToOp = map[string]Operation{
		"EQ": OperationEQ,
		"=":  OperationEQ,
		"NE": OperationNE,
		"!=": OperationNE,
		"LT": OperationLT,
		"<":  OperationLT,
		"LE": OperationLE,
		"<=": OperationLE,
		"GT": OperationGT,
		">":  OperationGT,
		"GE": OperationGE,
		">=": OperationGE,
	}
)

const defaultReplFactor = 2

// ParseQuery converts string to netmap.PlacementRule struct
func ParseQuery(s string) (*netmap.PlacementRule, error) {
	sv := &parsec.StringVessel{}
	sv.SetInput(s)
	if out, ok := parseRule(sv); ok {
		return out.(*netmap.PlacementRule), nil
	}
	return nil, errors.New("cant parse query")
}

func parseRule(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := forceCollect(trySucceed(replFactor), sfUnion)(in)
	if !ok {
		return nil, ok
	}

	result := out.([]interface{})

	rf := uint32(defaultReplFactor)
	if result[0] != nil {
		rf = result[0].(uint32)
	}

	return &netmap.PlacementRule{
		ReplFactor: rf,
		SFGroups:   result[1].([]SFGroup),
	}, true
}

func replFactor(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := forceCollect(ws, parsec.String("RF"), ws1, parseNumber)(in)
	if !ok {
		return nil, ok
	}
	return out.([]interface{})[3].(uint32), true
}

func filterGroup(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := forceCollect(
		parsec.String("FILTER"),
		parsec.Many1(forceCollect(ws1, parseString, ws1, parseOperation, ws1, parseString)),
	)(in)
	if !ok {
		return nil, false
	}

	var fs []Filter
	if lst := out.([]interface{}); len(lst) == 2 {
		for _, c := range lst[1].([]interface{}) {
			rule := c.([]interface{})
			fs = append(fs, Filter{
				Key: rule[1].(string),
				F:   NewFilter(rule[3].(Operation), rule[5].(string)),
			})
		}
	}
	return fs, true
}

func sfGroup(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := forceCollect(
		ws, parsec.String("SELECT"),
		parsec.Many1(forceCollect(ws1, parseNumber, ws1, parseString)),
		trySucceed(forceCollect(ws1, filterGroup)),
	)(in)
	if !ok {
		return nil, ok
	}

	lst := out.([]interface{})

	var ss = make([]Select, 0, len(lst[2].([]interface{})))
	for _, c := range lst[2].([]interface{}) {
		ss = append(ss, Select{
			Key:   c.([]interface{})[3].(string),
			Count: c.([]interface{})[1].(uint32),
		})
	}

	var fs []Filter
	if lst[3] != nil {
		fs = lst[3].([]interface{})[1].([]Filter)
	}

	return SFGroup{Filters: fs, Selectors: ss}, ok
}

func sfUnion(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := forceCollect(
		sfGroup,
		parsec.Many(forceCollect(
			ws, parsec.String(";"), ws,
			sfGroup,
		)),
	)(in)
	if !ok {
		return nil, false
	}

	// return error if there are any extra symbols
	if _, ok := in.Next(); ok {
		return nil, false
	}

	lst := out.([]interface{})
	s := []SFGroup{lst[0].(SFGroup)}
	for _, c := range lst[1].([]interface{}) {
		s = append(s, c.([]interface{})[3].(SFGroup))
	}
	return s, true
}

func parseOperation(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := parsec.Many1(parsec.Satisfy(func(r rune) bool { return !unicode.IsSpace(r) }))(in)
	if !ok {
		return nil, false
	}

	var s = make([]rune, 0, len(out.([]interface{})))
	for _, c := range out.([]interface{}) {
		s = append(s, c.(rune))
	}

	if op, ok := strToOp[string(s)]; ok {
		return op, true
	}
	return nil, false
}

func parseString(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := parsec.Many1(parsec.Satisfy(
		func(r rune) bool { return unicode.IsDigit(r) || unicode.IsLetter(r) }))(in)
	if !ok {
		return nil, false
	}

	var s = make([]rune, 0, len(out.([]interface{})))
	for _, c := range out.([]interface{}) {
		s = append(s, c.(rune))
	}
	return string(s), true
}

// Token that satisfies a condition.
func parseNumber(in parsec.Vessel) (parsec.Output, bool) {
	out, ok := parsec.Many1(parsec.Satisfy(unicode.IsNumber))(in)
	if !ok {
		return nil, false
	}

	r := uint32(0)
	for _, c := range out.([]interface{}) {
		r = r*10 + uint32(c.(rune)-'0')
	}
	return r, true
}

func forceCollect(parsers ...parsec.Parser) parsec.Parser {
	return func(in parsec.Vessel) (parsec.Output, bool) {
		matches := make([]interface{}, 0, len(parsers))
		st, pos := in.GetState(), in.GetPosition()
		for _, parser := range parsers {
			match, ok := parser(in)
			if !ok {
				in.SetState(st)
				in.SetPosition(pos)
				return nil, false
			}

			matches = append(matches, match)
		}
		return matches, true
	}
}

func trySucceed(p parsec.Parser) parsec.Parser {
	return func(in parsec.Vessel) (parsec.Output, bool) {
		out, _ := parsec.Try(p)(in)
		return out, true
	}
}
