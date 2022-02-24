package sqlstore

import (
	"fmt"
	"strconv"
	"strings"
)

// A handy little helper function for building queries. Appends 'value'
// to the 'args' slice and returns a string '$N' referring to the index
// of the value in args. For example:
//
// 	var args []interface{}
//  query = "select * from foo where id=" + nextBindVar(&args, 100)
//  db.Query(query, args...)
func nextBindVar(args *[]interface{}, value interface{}) string {
	*args = append(*args, value)
	return "$" + strconv.Itoa(len(*args))
}

// orderAndPaginateQuery is a helper function to simplify adding ordering and pagination statements to the end of a query
// with the appropriate binding variables amd returns the query string and list of arguments to pass to the query execution handler
func orderAndPaginateQuery(query string, orderColumns []string, pagination Pagination, args ...interface{}) (string, []interface{}) {
	ordering := "ASC"

	if pagination.Descending {
		ordering = "DESC"
	}

	sbOrderBy := strings.Builder{}

	if len(orderColumns) > 0 {
		sbOrderBy.WriteString("ORDER BY")

		for _, column := range orderColumns {
			sbOrderBy.WriteString(fmt.Sprintf(" %s %s", column, ordering))
		}
	}

	var paging string

	if pagination.Offset != 0 {
		paging = fmt.Sprintf("%sOFFSET %s ", paging, nextBindVar(&args, pagination.Offset))
	}

	if pagination.Limit != 0 {
		paging = fmt.Sprintf("%sLIMIT %s ", paging, nextBindVar(&args, pagination.Limit))
	}

	query = fmt.Sprintf("%s %s %s", query, sbOrderBy.String(), paging)

	return query, args
}
