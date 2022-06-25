package web

const (
	QueryType     = "$type"
	QueryFilter   = "$filter"
	QueryOrderBy  = "$orderBy"
	QueryPage     = "$page"
	QueryPageSize = "$pageSize"
	HeaderAttrs   = "$attrs"
)

// Reader function
type Reader func(ctx *Context, v Data) error

// Writer function
type Writer func(ctx *Context, v Data) error
