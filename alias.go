package query

import (
	"github.com/nspcc-dev/netmap/netgraph"
)

type (
	// Filter struct
	Filter = netgraph.Filter
	// Select struct
	Select = netgraph.Select
	// SFGroup struct
	SFGroup = netgraph.SFGroup
	// Operation struct
	Operation = netgraph.Operation
)

const (
	// OperationEQ A == B
	OperationEQ = netgraph.Operation_EQ
	// OperationNE A != B
	OperationNE = netgraph.Operation_NE
	// OperationLT A < B
	OperationLT = netgraph.Operation_LT
	// OperationLE A <= B
	OperationLE = netgraph.Operation_LE
	// OperationGT A > B
	OperationGT = netgraph.Operation_GT
	// OperationGE A >= B
	OperationGE = netgraph.Operation_GE
)

var (
	// FilterEQ returns filter, which checks if value is equal to v.
	FilterEQ = netgraph.FilterEQ
	// FilterNE returns filter, which checks if value is not equal to v.
	FilterNE = netgraph.FilterNE
	// FilterGT returns filter, which checks if value is greater than v.
	FilterGT = netgraph.FilterGT
	// NewFilter constructs SimpleFilter.
	NewFilter = netgraph.NewFilter
)
