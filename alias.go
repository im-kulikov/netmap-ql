package query

import (
	"github.com/nspcc-dev/netmap/netgraph"
)

type (
	Filter    = netgraph.Filter
	Select    = netgraph.Select
	SFGroup   = netgraph.SFGroup
	Operation = netgraph.Operation
)

const (
	Operation_EQ = netgraph.Operation_EQ
	Operation_NE = netgraph.Operation_NE
	Operation_LT = netgraph.Operation_LT
	Operation_LE = netgraph.Operation_LE
	Operation_GT = netgraph.Operation_GT
	Operation_GE = netgraph.Operation_GE
)

var (
	FilterEQ  = netgraph.FilterEQ
	FilterNE  = netgraph.FilterNE
	FilterGT  = netgraph.FilterGT
	NewFilter = netgraph.NewFilter
)
