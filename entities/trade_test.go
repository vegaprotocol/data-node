package entities_test

import (
	"encoding/hex"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/entities"
	"code.vegaprotocol.io/protos/vega"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProtoFromTrade(t *testing.T) {
	vegaTime := time.Now()
	idString := "bc2001bddac588f8aaae0d9bec3d6881a447b888447e5d0a9de92d149ba4e877"
	id, _ := hex.DecodeString(idString)
	marketIdString := "8cc0e020c0bc2f9eba77749d81ecec8283283b85941722c2cb88318aaf8b8cd8"
	marketId, _ := hex.DecodeString(marketIdString)
	priceString := "1000035452"
	price, _ := decimal.NewFromString(priceString)
	size := uint64(5)
	buyerIdString := "2e4f34a38204a2a155be678e670903ed8df96e813700729deacd3daf7e55039e"
	buyer, _ := hex.DecodeString(buyerIdString)
	sellerIdString := "8b6be1a03cc4d529f682887a78b66e6879d17f81e2b37356ca0acbc5d5886eb8"
	seller, _ := hex.DecodeString(sellerIdString)

	buyOrderIdString := "cf951606211775c43449807fe15f908704a85c514d65d549d67bbd6b5eef66bb"
	buyOrderId, _ := hex.DecodeString(buyOrderIdString)

	sellOrderIdString := "6a94947f724cdb7851bee793aca6888f68abbf8d49dfd0f778424a7ce42e7b7d"
	sellOrderId, _ := hex.DecodeString(sellOrderIdString)

	trade := entities.Trade{
		VegaTime:                vegaTime,
		ID:                      id,
		MarketID:                marketId,
		Price:                   price,
		Size:                    size,
		Buyer:                   buyer,
		Seller:                  seller,
		Aggressor:               entities.SideBuy,
		BuyOrder:                buyOrderId,
		SellOrder:               sellOrderId,
		Type:                    entities.TradeTypeNetworkCloseOutGood,
		BuyerMakerFee:           decimal.NewFromInt(2),
		BuyerInfrastructureFee:  decimal.NewFromInt(3),
		BuyerLiquidityFee:       decimal.NewFromInt(4),
		SellerMakerFee:          decimal.NewFromInt(1),
		SellerInfrastructureFee: decimal.NewFromInt(10),
		SellerLiquidityFee:      decimal.NewFromInt(100),
		BuyerAuctionBatch:       3,
		SellerAuctionBatch:      4,
	}

	p := trade.ToProto()

	assert.Equal(t, vegaTime.UnixNano(), p.Timestamp)
	assert.Equal(t, idString, p.Id)
	assert.Equal(t, marketIdString, p.MarketId)
	assert.Equal(t, priceString, p.Price)
	assert.Equal(t, size, p.Size)
	assert.Equal(t, buyerIdString, p.Buyer)
	assert.Equal(t, sellerIdString, p.Seller)
	assert.Equal(t, vega.Side_SIDE_BUY, p.Aggressor)
	assert.Equal(t, buyOrderIdString, p.BuyOrder)
	assert.Equal(t, sellOrderIdString, p.SellOrder)
	assert.Equal(t, vega.Trade_TYPE_NETWORK_CLOSE_OUT_GOOD, p.Type)
	assert.Equal(t, "2", p.BuyerFee.MakerFee)
	assert.Equal(t, "3", p.BuyerFee.InfrastructureFee)
	assert.Equal(t, "4", p.BuyerFee.LiquidityFee)
	assert.Equal(t, "1", p.SellerFee.MakerFee)
	assert.Equal(t, "10", p.SellerFee.InfrastructureFee)
	assert.Equal(t, "100", p.SellerFee.LiquidityFee)
	assert.Equal(t, uint64(3), p.BuyerAuctionBatch)
	assert.Equal(t, uint64(4), p.SellerAuctionBatch)
}

func TestTradeFromProto(t *testing.T) {
	tradeEventProto := vega.Trade{
		Id:        "521127F24B1FA40311BA2FB3F6977310346346604B275DB7B767B04240A5A5C3",
		MarketId:  "8cc0e020c0bc2f9eba77749d81ecec8283283b85941722c2cb88318aaf8b8cd8",
		Price:     "1000097674",
		Size:      1,
		Buyer:     "b4376d805a888548baabfae74ef6f4fa4680dc9718bab355fa7191715de4fafe",
		Seller:    "539e8c7c8c15044a6b37a8bf4d7d988588b2f63ed48666b342bc530c8312e002",
		Aggressor: vega.Side_SIDE_SELL,
		BuyOrder:  "0976E6CFE1513C46D5EC8877EFB51E6F12EB24709131D08EF358310FA4409158",
		SellOrder: "459B8150105322406C1CEABF596E0E13ED113A98C1E290E2144D7A6236EDC6C2",
		Timestamp: 1644573750767832307,
		Type:      vega.Trade_TYPE_DEFAULT,
		BuyerFee: &vega.Fee{
			MakerFee:          "4000142",
			InfrastructureFee: "10000036",
			LiquidityFee:      "10000355",
		},
		SellerFee:          nil,
		BuyerAuctionBatch:  3,
		SellerAuctionBatch: 0,
	}

	testVegaTime := time.Now()
	trade, err := entities.TradeFromProto(&tradeEventProto, testVegaTime, 5)
	if err != nil {
		t.Fatalf("failed to convert proto to trade:%s", err)
	}

	assert.Equal(t, testVegaTime.Add(5*time.Microsecond), trade.SyntheticTime)
	assert.Equal(t, testVegaTime, trade.VegaTime)
	assert.Equal(t, uint64(5), trade.SeqNum)

	idBytes, _ := hex.DecodeString(tradeEventProto.Id)
	assert.Equal(t, idBytes, trade.ID)
	marketIdBytes, _ := hex.DecodeString(tradeEventProto.MarketId)
	assert.Equal(t, marketIdBytes, trade.MarketID)
	price, _ := decimal.NewFromString(tradeEventProto.Price)
	assert.Equal(t, price, trade.Price)
	size := tradeEventProto.Size
	assert.Equal(t, size, trade.Size)
	buyerBytes, _ := hex.DecodeString(tradeEventProto.Buyer)
	assert.Equal(t, buyerBytes, trade.Buyer)
	sellerBytes, _ := hex.DecodeString(tradeEventProto.Seller)
	assert.Equal(t, sellerBytes, trade.Seller)
	assert.Equal(t, entities.SideSell, trade.Aggressor)
	buyOrderBytes, _ := hex.DecodeString(tradeEventProto.BuyOrder)
	assert.Equal(t, buyOrderBytes, trade.BuyOrder)
	sellOrderBytes, _ := hex.DecodeString(tradeEventProto.SellOrder)
	assert.Equal(t, sellOrderBytes, trade.SellOrder)
	assert.Equal(t, entities.TradeTypeDefault, trade.Type)

	buyerMakerFee, _ := decimal.NewFromString(tradeEventProto.BuyerFee.MakerFee)
	buyerLiquidityFee, _ := decimal.NewFromString(tradeEventProto.BuyerFee.LiquidityFee)
	buyerInfraFee, _ := decimal.NewFromString(tradeEventProto.BuyerFee.InfrastructureFee)
	assert.Equal(t, buyerMakerFee, trade.BuyerMakerFee)
	assert.Equal(t, buyerLiquidityFee, trade.BuyerLiquidityFee)
	assert.Equal(t, buyerInfraFee, trade.BuyerInfrastructureFee)
	assert.Equal(t, decimal.Zero, trade.SellerMakerFee)
	assert.Equal(t, decimal.Zero, trade.SellerLiquidityFee)
	assert.Equal(t, decimal.Zero, trade.SellerInfrastructureFee)
	assert.Equal(t, tradeEventProto.BuyerAuctionBatch, trade.BuyerAuctionBatch)
	assert.Equal(t, tradeEventProto.SellerAuctionBatch, trade.SellerAuctionBatch)
}
