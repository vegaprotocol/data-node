package entities_test

import (
	"encoding/base64"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	v2 "code.vegaprotocol.io/protos/data-node/api/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageEntities(t *testing.T) {
	t.Run("Number of results is 2 more then the page limit", func(t *testing.T) {
		t.Run("The results are returned in order and we have next and previous when we are moving forward", testPageEntitiesForwardHasNextAndPrevious)
		t.Run("The results are returned in order and we have next and previous when we are moving backward", testPageEntitiesBackwardHasNextAndPrevious)
	})

	t.Run("Number of results is 1 more than the page limit", func(t *testing.T) {
		t.Run("When moving forward, we have a previous page, but no next page", testPagedEntitiesForwardHasPreviousButNoNext)
		t.Run("When moving backward, we have a next page, but no previous page", testPagedEntitiesBackwardHasNextButNoPrevious)
	})

	t.Run("Number of results is equal to the page limit", func(t *testing.T) {
		t.Run("When moving forward, we have no previous or next page", testPagedEntitiesForwardNoNextOrPreviousEqualLimit)
		t.Run("When moving backward, we have no previous or next page", testPagedEntitiesBackwardNoNextOrPreviousEqualLimit)
	})

	t.Run("Number of results is less than the page limit", func(t *testing.T) {
		t.Run("When moving forward, we have no previous or next page", testPagedEntitiesForwardNoNextOrPreviousLessThanLimit)
		t.Run("When moving backward, we have no previous or next page", testPagedEntitiesBackwardNoNextOrPreviousLessThanLimit)
	})
}

func testPageEntitiesForwardHasNextAndPrevious(t *testing.T) {
	trades := getTradesForward(t, 0, 0) // 0, 0 return all entries
	first := int32(5)
	afterTs := time.Unix(0, 1000000000000)
	after := base64.StdEncoding.EncodeToString([]byte(afterTs.Format(time.RFC3339Nano)))
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  &first,
			After:  &after,
			Last:   nil,
			Before: nil,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000001000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000005000000).Format(time.RFC3339Nano)

	wantPaged := trades[1:6]
	wantInfo := entities.PageInfo{
		HasNextPage:     true,
		HasPreviousPage: true,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func testPageEntitiesBackwardHasNextAndPrevious(t *testing.T) {
	trades := getTradesBackward(t, 0, 0) // 0, 0 return all entries
	last := int32(5)
	beforeTs := time.Unix(0, 1000006000000)
	before := base64.StdEncoding.EncodeToString([]byte(beforeTs.Format(time.RFC3339Nano)))
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  nil,
			After:  nil,
			Last:   &last,
			Before: &before,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000001000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000005000000).Format(time.RFC3339Nano)

	wantPaged := getTradesForward(t, 1, 6)
	wantInfo := entities.PageInfo{
		HasNextPage:     true,
		HasPreviousPage: true,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func testPagedEntitiesForwardHasPreviousButNoNext(t *testing.T) {
	trades := getTradesForward(t, 1, 0) // 0, 0 return all entries
	first := int32(5)
	afterTs := time.Unix(0, 1000001000000)
	after := base64.StdEncoding.EncodeToString([]byte(afterTs.Format(time.RFC3339Nano)))
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  &first,
			After:  &after,
			Last:   nil,
			Before: nil,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000002000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000006000000).Format(time.RFC3339Nano)

	wantPaged := trades[1:6]
	wantInfo := entities.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: true,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func testPagedEntitiesBackwardHasNextButNoPrevious(t *testing.T) {
	trades := getTradesBackward(t, 1, 0) // 0, 0 return all entries
	last := int32(5)
	beforeTs := time.Unix(0, 1000005000000)
	before := base64.StdEncoding.EncodeToString([]byte(beforeTs.Format(time.RFC3339Nano)))
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  nil,
			After:  nil,
			Last:   &last,
			Before: &before,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000000000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000004000000).Format(time.RFC3339Nano)

	wantPaged := getTradesForward(t, 0, 5)
	wantInfo := entities.PageInfo{
		HasNextPage:     true,
		HasPreviousPage: false,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func testPagedEntitiesForwardNoNextOrPreviousEqualLimit(t *testing.T) {
	trades := getTradesForward(t, 0, 5) // 0, 0 return all entries
	first := int32(5)
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  &first,
			After:  nil,
			Last:   nil,
			Before: nil,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000000000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000004000000).Format(time.RFC3339Nano)

	wantPaged := trades
	wantInfo := entities.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func testPagedEntitiesBackwardNoNextOrPreviousEqualLimit(t *testing.T) {
	trades := getTradesBackward(t, 0, 5) // 0, 0 return all entries
	last := int32(5)
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  nil,
			After:  nil,
			Last:   &last,
			Before: nil,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000002000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000006000000).Format(time.RFC3339Nano)

	wantPaged := getTradesForward(t, 2, 0)
	wantInfo := entities.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func testPagedEntitiesForwardNoNextOrPreviousLessThanLimit(t *testing.T) {
	trades := getTradesForward(t, 0, 3) // 0, 0 return all entries
	first := int32(5)
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  &first,
			After:  nil,
			Last:   nil,
			Before: nil,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000000000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000002000000).Format(time.RFC3339Nano)

	wantPaged := trades
	wantInfo := entities.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func testPagedEntitiesBackwardNoNextOrPreviousLessThanLimit(t *testing.T) {
	trades := getTradesBackward(t, 0, 3) // 0, 0 return all entries
	last := int32(5)
	cursor, err := entities.PaginationFromProto(
		&v2.Pagination{
			First:  nil,
			After:  nil,
			Last:   &last,
			Before: nil,
		})
	require.NoError(t, err)
	gotPaged, gotInfo := entities.PageEntities(trades, cursor)

	startCursor := time.Unix(0, 1000004000000).Format(time.RFC3339Nano)
	endCursor := time.Unix(0, 1000006000000).Format(time.RFC3339Nano)

	wantPaged := getTradesForward(t, 4, 0)
	wantInfo := entities.PageInfo{
		HasNextPage:     false,
		HasPreviousPage: false,
		StartCursor:     base64.StdEncoding.EncodeToString([]byte(startCursor)),
		EndCursor:       base64.StdEncoding.EncodeToString([]byte(endCursor)),
	}
	assert.Equal(t, wantPaged, gotPaged)
	assert.Equal(t, wantInfo, gotInfo)
}

func getTradesForward(t *testing.T, start, end int) []entities.Trade {
	t.Helper()
	trades := []entities.Trade{
		{
			SyntheticTime: time.Unix(0, 1000000000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000001000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000002000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000003000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000004000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000005000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000006000000),
		},
	}

	if end == 0 {
		end = len(trades)
	}

	if end < start {
		end = start
	}

	return trades[start:end]
}

func getTradesBackward(t *testing.T, start, end int) []entities.Trade {
	t.Helper()
	trades := []entities.Trade{
		{
			SyntheticTime: time.Unix(0, 1000006000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000005000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000004000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000003000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000002000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000001000000),
		},
		{
			SyntheticTime: time.Unix(0, 1000000000000),
		},
	}

	if end == 0 {
		end = len(trades)
	}

	if end < start {
		end = start
	}

	return trades[start:end]
}