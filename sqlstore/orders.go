package sqlstore

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/entities"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/pkg/errors"
)

type Orders struct {
	*ConnectionSource
	batcher                    MapBatcher[entities.OrderKey, entities.Order]
	liveOrders                 map[entities.OrderID]entities.Order
	liveOrdersCompletedInBlock map[entities.OrderID]entities.Order
	newLiveOrdersInBlock       map[entities.OrderID]entities.Order
	updatedLiveOrdersInBlock   map[entities.OrderID]entities.Order

	liveOrdersBatcher ListBatcher
}

func NewOrders(connectionSource *ConnectionSource) *Orders {
	a := &Orders{
		ConnectionSource: connectionSource,
		batcher: NewMapBatcher[entities.OrderKey, entities.Order](
			"order_history",
			entities.OrderColumns),
		liveOrdersCompletedInBlock: map[entities.OrderID]entities.Order{},
		newLiveOrdersInBlock:       map[entities.OrderID]entities.Order{},
		updatedLiveOrdersInBlock:   map[entities.OrderID]entities.Order{},

		liveOrdersBatcher: NewListBatcher("live_orders", entities.OrderColumns),
	}

	return a
}

func (os *Orders) Flush(ctx context.Context) error {

	var idsToDelete [][]byte

	for _, newLiveOrder := range os.newLiveOrdersInBlock {
		os.liveOrders[newLiveOrder.ID] = newLiveOrder
		os.liveOrdersBatcher.Add(newLiveOrder)
	}

	for _, updatedOrder := range os.updatedLiveOrdersInBlock {
		os.liveOrders[updatedOrder.ID] = updatedOrder
		bytes, _ := updatedOrder.ID.Bytes()
		idsToDelete = append(idsToDelete, bytes)

		os.liveOrdersBatcher.Add(updatedOrder)
	}

	for _, completedOrder := range os.liveOrdersCompletedInBlock {
		delete(os.liveOrders, completedOrder.ID)
		bytes, _ := completedOrder.ID.Bytes()
		idsToDelete = append(idsToDelete, bytes)
	}

	results, err := os.ConnectionSource.Connection.Exec(ctx, "delete from live_orders where id = ANY ($1)", idsToDelete)
	if err != nil {
		return errors.Wrap(err, "failed to delete from live orders")
	}

	if results.RowsAffected() != int64(len(idsToDelete)) {
		return errors.Errorf("expected to delete %d orders, deleted %d", len(idsToDelete), results.RowsAffected())
	}

	os.liveOrdersBatcher.Flush(ctx, os.Connection)

	os.liveOrdersCompletedInBlock = map[entities.OrderID]entities.Order{}
	os.newLiveOrdersInBlock = map[entities.OrderID]entities.Order{}
	os.updatedLiveOrdersInBlock = map[entities.OrderID]entities.Order{}

	return os.batcher.Flush(ctx, os.Connection)
}

// Add inserts an order update row into the database if an row for this (block time, order id, version)
// does not already exist; otherwise update the existing row with information supplied.
// Currently we only store the last update to an order per block, so the order history is not
// complete if multiple updates happen in one block.
func (os *Orders) Add(ctx context.Context, o entities.Order) error {

	if os.liveOrders == nil {
		os.liveOrders = map[entities.OrderID]entities.Order{}

		liveOrders, err := os.GetLiveOrders(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to get live orders")
		}

		for _, order := range liveOrders {
			os.liveOrders[order.ID] = order
		}
	}

	if o.IsLive() {
		if _, alreadyLive := os.liveOrders[o.ID]; alreadyLive {
			os.updatedLiveOrdersInBlock[o.ID] = o
		} else {
			os.newLiveOrdersInBlock[o.ID] = o
		}
		delete(os.liveOrdersCompletedInBlock, o.ID)
	} else {
		if _, alreadyLive := os.liveOrders[o.ID]; alreadyLive {
			os.liveOrdersCompletedInBlock[o.ID] = o
		}
		delete(os.updatedLiveOrdersInBlock, o.ID)
		delete(os.newLiveOrdersInBlock, o.ID)
	}

	os.batcher.Add(o)

	return nil
}

// GetAll returns all updates to all orders (including changes to orders that don't increment the version number)
func (os *Orders) GetAll(ctx context.Context) ([]entities.Order, error) {
	orders := []entities.Order{}
	err := pgxscan.Select(ctx, os.Connection, &orders, `
		SELECT * from order_history;`)
	return orders, err
}

// GetByOrderId returns the last update of the order with the given ID
func (os *Orders) GetByOrderID(ctx context.Context, orderIdStr string, version *int32) (entities.Order, error) {
	var err error
	order := entities.Order{}
	orderId := entities.NewOrderID(orderIdStr)

	if version != nil && *version > 0 {
		err = pgxscan.Get(ctx, os.Connection, &order, `SELECT * FROM orders_current_versions WHERE id=$1 and version=$2`, orderId, version)
	} else {
		err = pgxscan.Get(ctx, os.Connection, &order, `SELECT * FROM orders_current WHERE id=$1`, orderId)
	}
	return order, err
}

// GetByMarket returns the last update of the all the orders in a particular market
func (os *Orders) GetByMarket(ctx context.Context, marketIdStr string, p entities.Pagination) ([]entities.Order, error) {
	marketId := entities.NewMarketID(marketIdStr)
	query := `SELECT * from orders_current WHERE market_id=$1`
	args := []interface{}{marketId}
	return os.queryOrders(ctx, query, args, &p)
}

// GetByParty returns the last update of the all the orders in a particular party
func (os *Orders) GetByParty(ctx context.Context, partyIdStr string, p entities.Pagination) ([]entities.Order, error) {
	partyId := entities.NewPartyID(partyIdStr)
	query := `SELECT * from orders_current WHERE party_id=$1`
	args := []interface{}{partyId}
	return os.queryOrders(ctx, query, args, &p)
}

// GetByReference returns the last update of orders with the specified user-suppled reference
func (os *Orders) GetByReference(ctx context.Context, reference string, p entities.Pagination) ([]entities.Order, error) {
	query := `SELECT * from orders_current WHERE reference=$1`
	args := []interface{}{reference}
	return os.queryOrders(ctx, query, args, &p)
}

// GetAllVersionsByOrderID the last update to all versions (e.g. manual changes that lead to
// incrementing the version field) of a given order id.
func (os *Orders) GetAllVersionsByOrderID(ctx context.Context, id string, p entities.Pagination) ([]entities.Order, error) {
	query := `SELECT * from orders_current_versions WHERE id=$1`
	args := []interface{}{entities.NewOrderID(id)}
	return os.queryOrders(ctx, query, args, &p)
}

// GetLiveOrders fetches all currently live orders so the market depth data can be rebuilt
// from the orders data in the database
func (os *Orders) GetLiveOrders(ctx context.Context) ([]entities.Order, error) {
	query := `select * from live_orders`
	return os.queryOrders(ctx, query, nil, nil)
}

// -------------------------------------------- Utility Methods

func (os *Orders) queryOrders(ctx context.Context, query string, args []interface{}, p *entities.Pagination) ([]entities.Order, error) {
	if p != nil {
		query, args = paginateOrderQuery(query, args, *p)
	}

	orders := []entities.Order{}
	err := pgxscan.Select(ctx, os.Connection, &orders, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying orders: %w", err)
	}
	return orders, nil
}

func paginateOrderQuery(query string, args []interface{}, p entities.Pagination) (string, []interface{}) {
	dir := "ASC"
	if p.Descending {
		dir = "DESC"
	}

	var limit interface{} = nil
	if p.Limit != 0 {
		limit = p.Limit
	}

	query = fmt.Sprintf(" %s ORDER BY vega_time %s, id %s LIMIT %s OFFSET %s",
		query, dir, dir, nextBindVar(&args, limit), nextBindVar(&args, p.Skip))

	return query, args
}
