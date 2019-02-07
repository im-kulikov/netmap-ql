package query

import (
	"github.com/nspcc-dev/netmap"
)

type (
	// Filter struct
	Filter = netmap.Filter
	// Select struct
	Select = netmap.Select
	// SFGroup struct
	SFGroup = netmap.SFGroup
	// Operation struct
	Operation = netmap.Operation
)

const (
	// OperationEQ A == B
	OperationEQ = netmap.Operation_EQ
	// OperationNE A != B
	OperationNE = netmap.Operation_NE
	// OperationLT A < B
	OperationLT = netmap.Operation_LT
	// OperationLE A <= B
	OperationLE = netmap.Operation_LE
	// OperationGT A > B
	OperationGT = netmap.Operation_GT
	// OperationGE A >= B
	OperationGE = netmap.Operation_GE
)

var (
	// FilterEQ returns filter, which checks if value is equal to v.
	FilterEQ = netmap.FilterEQ
	// FilterNE returns filter, which checks if value is not equal to v.
	FilterNE = netmap.FilterNE
	// FilterGT returns filter, which checks if value is greater than v.
	FilterGT = netmap.FilterGT
	// NewFilter constructs SimpleFilter.
	NewFilter = netmap.NewFilter
)
