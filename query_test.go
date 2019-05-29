package query

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/nspcc-dev/netmap"
	"github.com/stretchr/testify/require"
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

	for s, op := range strToOp {
		out, ok = parseOp(s)
		require.True(t, ok)
		require.Equal(t, op, out)
	}

	for _, s := range []string{"EE", ">>", "==="} {
		out, ok = parseOp(s)
		require.Nil(t, out)
		require.False(t, ok)
	}
}

func TestParseNumber(t *testing.T) {
	var (
		out      gp.Output
		ok       bool
		n        int32
		parseNum = createParser(parseNumber)
	)
	for i := 0; i < 10; i++ {
		n = rand.Int31()
		out, ok = parseNum(strconv.FormatInt(int64(n), 10))
		require.True(t, ok)
		require.Equal(t, uint32(n), out)
	}

	for _, s := range []string{"a", ""} {
		out, ok = parseNum(s)
		require.Nil(t, out)
		require.False(t, ok)
	}
}

func rule(s []SFGroup, rf uint32) *netmap.PlacementRule {
	return &netmap.PlacementRule{
		SFGroups:   s,
		ReplFactor: rf,
	}
}

func TestParseQuery(t *testing.T) {
	var (
		out interface{}
		err error
		exp *netmap.PlacementRule
	)

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 1}},
	}}, defaultReplFactor)
	out, err = ParseQuery(`SELECT 1 Country`)
	require.NoError(t, err)
	require.Equal(t, exp, out)

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 1}},
	}}, 10)
	out, err = ParseQuery(`RF 10 SELECT 1 Country`)
	require.NoError(t, err)
	require.Equal(t, exp, out)

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 1}},
		Filters:   []Filter{{Key: "Country", F: FilterNE("Russia")}},
	}}, defaultReplFactor)
	out, err = ParseQuery(`SELECT 1 Country FILTER Country NE Russia`)
	require.NoError(t, err)
	require.Equal(t, exp, out)

	exp = rule([]SFGroup{{
		Selectors: []Select{{Key: "Country", Count: 11}},
		Filters:   []Filter{{Key: "Trust", F: FilterGT(10)}},
	}}, defaultReplFactor)
	out, err = ParseQuery(`SELECT 11 Country FILTER Trust > 10`)
	require.NoError(t, err)
	require.Equal(t, exp, out)

	exp = rule([]SFGroup{{
		Selectors: []Select{
			{Key: "Country", Count: 1},
			{Key: "City", Count: 2},
		},
		Filters: []Filter{{Key: "Location", F: FilterNE("Europe")}},
	}}, 4)
	out, err = ParseQuery(`RF 4 SELECT 1 Country 2 City FILTER Location NE Europe`)
	require.NoError(t, err)
	require.Equal(t, exp, out)

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
	require.NoError(t, err)
	require.Equal(t, exp, out)

	for _, q := range invalidQueries {
		_, err = ParseQuery(q)
		require.Errorf(t, err, "parsing query `%s`", q)
	}
}
