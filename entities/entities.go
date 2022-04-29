package entities

type PagedEntity interface {
	Market | Party | Trade
	Cursor() string
}

func PageEntities[T PagedEntity](items []T, cursor Cursor) ([]T, PageInfo) {
	var pagedItems []T
	var limit int
	var pageInfo PageInfo

	if len(items) == 0 {
		return pagedItems, pageInfo
	}

	if cursor.Forward != nil && cursor.Forward.Limit != nil {
		limit = int(*cursor.Forward.Limit)
		switch len(items) {
		case limit + 2:
			pagedItems = items[1 : limit+1]
			pageInfo.HasNextPage = true
			pageInfo.HasPreviousPage = true
		case limit + 1:
			if cursor.Forward.Cursor == nil {
				pagedItems = items[0:limit]
				pageInfo.HasNextPage = true
				pageInfo.HasPreviousPage = false
			} else {
				pagedItems = items[1:]
				pageInfo.HasNextPage = false
				pageInfo.HasPreviousPage = true
			}
		default:
			pagedItems = items
			pageInfo.HasNextPage = false
			pageInfo.HasPreviousPage = false
		}
	} else if cursor.Backward != nil && cursor.Backward.Limit != nil {
		limit = int(*cursor.Backward.Limit)
		switch len(items) {
		case limit + 2:
			pagedItems = reverseSlice(items[1 : limit+1])
			pageInfo.HasNextPage = true
			pageInfo.HasPreviousPage = true
		case limit + 1:
			if cursor.Backward.Cursor == nil {
				pagedItems = reverseSlice(items[0:limit])
				pageInfo.HasNextPage = false
				pageInfo.HasPreviousPage = true
			} else {
				pagedItems = reverseSlice(items[1:])
				pageInfo.HasNextPage = true
				pageInfo.HasPreviousPage = false
			}
		default:
			pagedItems = reverseSlice(items)
			pageInfo.HasNextPage = false
			pageInfo.HasPreviousPage = false
		}
	} else {
		pagedItems = items
		pageInfo.HasNextPage = false
		pageInfo.HasPreviousPage = false
	}

	pageInfo.StartCursor = pagedItems[0].Cursor()
	pageInfo.EndCursor = pagedItems[len(pagedItems)-1].Cursor()

	return pagedItems, pageInfo
}

func reverseSlice[T any](input []T) (reversed []T) {
	reversed = make([]T, len(input))
	copy(reversed, input)
	for i, j := 0, len(input)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = input[j], input[i]
	}
	return
}
