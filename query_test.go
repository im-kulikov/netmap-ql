package query

import (
	"math/rand"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
	gp "github.com/vito/go-parse"
)

var invalidQueries = []string{
	`SELECT 3`, // no option
	`SELECT 1 c 2`,
	`SELECT 1 c 2 FILTER a EQ b`,
	`SELEC 1 c`,                // wrong keyword
	`SELECT 1 c FILTER a EE b`, // wrong operation
	`SELECT 1 c FILTER a EE b ; SELECT 3`,
	`RF SELECT 1 c`,
}

func createParser(f gp.Parser) func(s string) (gp.Output, bool) {
	return func(s string) (gp.Output, bool) {
		sv := &gp.StringVessel{}
		sv.SetInput(s)
		return f(sv)
	}
}

func TestParseOperation(t *testing.T) {
	var (
		out     gp.Output
		ok      bool
		parseOp = createParser(parseOperation)
	)
	g := NewGomegaWithT(t)

	for s, op := range strToOp {
		out, ok = parseOp(s)
		g.Expect(ok).To(BeTrue())
		g.Expect(out).To(Equal(op))
	}

	for _, s := range []string{"EE", ">>", "==="} {
		out, ok = parseOp(s)
		g.Expect(ok).To(BeFalse())
	}
}

func TestParseNumber(t *testing.T) {
	var (
		out      gp.Output
		ok       bool
		n        int32
		parseNum = createParser(parseNumber)
	)
	g := NewGomegaWithT(t)

	for i := 0; i < 10; i++ {
		n = rand.Int31()
		out, ok = parseNum(strconv.FormatInt(int64(n), 10))
		g.Expect(ok).To(BeTrue())
		g.Expect(out).To(BeNumerically("==", n))
	}

	for _, s := range []string{"a", ""} {
		out, ok = parseNum(s)
		g.Expect(ok).To(BeFalse())
	}
}

func rule(s []SFGroup, rf uint32) *PlacementRule {
	return &PlacementRule{
		SFGroups: s,
		ReplFactor: rf,
	}
}

func TestParseQuery(t *testing.T) {
	var (
		out interface{}
		err error
		exp *PlacementRule
	)

	g := NewGomegaWithT(t)

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 1}},
	}}, defaultReplFactor)
	out, err = ParseQuery(`SELECT 1 Country`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(out).To(Equal(exp))

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 1}},
	}}, 10)
	out, err = ParseQuery(`RF 10 SELECT 1 Country`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(out).To(Equal(exp))

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 1}},
		Filters:   []Filter{{Key: "Country", F: FilterNE("Russia")}},
	}}, defaultReplFactor)
	out, err = ParseQuery(`SELECT 1 Country FILTER Country NE Russia`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(out).To(Equal(exp))

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 11}},
		Filters:   []Filter{{Key: "Trust", F: FilterGT(10)}},
	}}, defaultReplFactor)
	out, err = ParseQuery(`SELECT 11 Country FILTER Trust > 10`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(out).To(Equal(exp))

	exp = rule([]SFGroup{{
		Selectors: []Select{
			{Key: "Country", Count: 1},
			{Key: "City", Count: 2},
		},
		Filters: []Filter{{Key: "Location", F: FilterNE("Europe")}},
	}}, 4)
	out, err = ParseQuery(`RF 4 SELECT 1 Country 2 City FILTER Location NE Europe`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(out).To(Equal(exp))

	exp = rule([]SFGroup{
		{
			Selectors: []Select{
				{Key: "Country", Count: 1},
				{Key: "City", Count: 2},
			},
			Filters: []Filter{{Key: "Location", F: FilterNE("Europe")}},
		},
		{
			Selectors: []Select{{Key: "Country", Count: 1}},
			Filters:   []Filter{{Key: "Country", F: FilterEQ("Germany")}},
		},
	}, defaultReplFactor)
	out, err = ParseQuery(`
SELECT 1 Country 2 City
FILTER Location NE Europe
;
SELECT 1 Country
FILTER Country EQ Germany`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(out).To(Equal(exp))

	for _, q := range invalidQueries {
		_, err = ParseQuery(q)
		g.Expect(err).To(HaveOccurred(), "parsing query `%s`", q)
	}
}
