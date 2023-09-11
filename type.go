package web

const (
	QueryType     = "_type"
	QueryFilter   = "_filter"
	QueryOrderBy  = "_orderBy"
	QueryPage     = "_page"
	QueryPageSize = "_pageSize"
	HeaderAttrs   = "_attrs"
)

// Reader function
type Reader func(ctx *Context, v Data) error

// Writer function
type Writer func(ctx *Context, v Data) error
