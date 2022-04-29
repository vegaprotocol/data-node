package entities

import (
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	"github.com/pkg/errors"
)

type Pagination struct {
	Skip       uint64
	Limit      uint64
	Descending bool
}

func PaginationFromProto(pp *v2.Pagination) Pagination {
	return Pagination{
		Skip:       pp.Skip,
		Limit:      pp.Limit,
		Descending: pp.Descending,
	}
}

type Cursor struct {
	Forward  *offset
	Backward *offset
}

func CursorFromProto(cp *v2.Cursor) (Cursor, error) {
	if cp == nil {
		return Cursor{}, errors.New("cursor is not provided")
	}

	cursor := Cursor{
		Forward: &offset{
			Limit:  cp.First,
			Cursor: cp.After,
		},
		Backward: &offset{
			Limit:  cp.Last,
			Cursor: cp.Before,
		},
	}

	err := validateCursor(cursor)
	if err != nil {
		return Cursor{}, err
	}

	return cursor, nil
}

type offset struct {
	Limit  *int32
	Cursor *string
}

func (o offset) IsSet() bool {
	return o.Limit != nil
}

func validateCursor(cursor Cursor) error {
	if cursor.Forward.IsSet() && cursor.Backward.IsSet() {
		return errors.New("cannot provide both forward and backward cursors")
	}

	var cursorOffset offset

	if cursor.Forward.IsSet() {
		cursorOffset = *cursor.Forward
	} else if cursor.Backward.IsSet() {
		cursorOffset = *cursor.Backward
	} else {
		// no cursor is provided is okay
		return nil
	}

	limit := *cursorOffset.Limit
	if limit <= 0 {
		return errors.New("cursor limit must be greater than 0")
	}

	return nil
}

type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
	EndCursor       string
}

func (p PageInfo) ToProto() *v2.PageInfo {
	return &v2.PageInfo{
		HasNextPage:     p.HasNextPage,
		HasPreviousPage: p.HasPreviousPage,
		StartCursor:     p.StartCursor,
		EndCursor:       p.EndCursor,
	}
}
