package logic

import (
	"context"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/pkg/errors"
	"eggServer/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/tonkeeper/tonapi-go"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"log"
	"time"
)

type TickerResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		InstType  string `json:"instType"`
		InstId    string `json:"instId"`
		Last      string `json:"last"`
		LastSz    string `json:"lastSz"`
		AskPx     string `json:"askPx"`
		AskSz     string `json:"askSz"`
		BidPx     string `json:"bidPx"`
		BidSz     string `json:"bidSz"`
		Open24H   string `json:"open24h"`
		High24H   string `json:"high24h"`
		Low24H    string `json:"low24h"`
		VolCcy24H string `json:"volCcy24h"`
		Vol24H    string `json:"vol24h"`
		Ts        string `json:"ts"`
		SodUtc0   string `json:"sodUtc0"`
		SodUtc8   string `json:"sodUtc8"`
	} `json:"data"`
}

type tonapiLogic struct {
	tonapi *tonapi.Client
	master *jetton.Client
	client *liteclient.ConnectionPool
}

var TonapiLogic = new(tonapiLogic)

func (s *tonapiLogic) Init() {
	tonapi, err := tonapi.New()
	if err != nil {
		log.Fatal(err)
	}
	s.tonapi = tonapi

	s.initJettonMaster()
}

func (s *tonapiLogic) initJettonMaster() {
	s.client = liteclient.NewConnectionPool()

	// connect to testnet lite server
	err := s.client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		panic(err)
	}

	// initialize ton api lite connection wrapper
	api := ton.NewAPIClient(s.client).WithRetry()

	tokenContract := address.MustParseAddr("EQCxE6mUtQJKFnGfaROTKOt1lZbDiiX1kCixRv7Nw2Id_sDs") // usdt合约的代币地址
	s.master = jetton.NewJettonMasterClient(api, tokenContract)
}

func (s *tonapiLogic) GetJettonWalletAddress(ctx context.Context, addr string) string {
	logger := contextx.FromLogger(ctx)

	tokenWallet, err := s.master.GetJettonWallet(s.client.StickyContext(ctx), address.MustParseRawAddr(addr))
	if err != nil {
		logger.Errorf("TonapiLogic.GetJettonWallet error:%s", err.Error())
		return ""
	}

	resp, err := s.tonapi.GetAccount(ctx, tonapi.GetAccountParams{AccountID: tokenWallet.Address().String()})
	if err != nil {
		logger.Errorf("TonapiLogic.GetJettonWalletAddress error:%s", err.Error())
		return ""
	}

	return resp.Address
}

func (s *tonapiLogic) TonApi() *tonapi.Client {
	return s.tonapi
}

// Ticker 行情数据
func (s *tonapiLogic) Ticker(ctx context.Context, instId string) (float64, error) {
	logger := contextx.FromLogger(ctx)

	rb := contextx.FromRB(ctx)

	// 分布式锁
	m := rb.NewMutex(instId)
	if err := m.Lock(ctx); err != nil {
		logger.Errorf("TonapiLogic.Ticker error:%s", err.Error())
		return 0, errors.NewResponseError(constant.ServerBusy, err)
	}

	defer func() {
		if _, err := m.Unlock(context.Background()); err != nil {
			logger.WithError(err).Error("error on mutex unlock")
		}
	}()

	result, err := rb.Client().Get(ctx, instId).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return 0, err
	}

	if result == "" {
		body, err := utils.Request("GET", fmt.Sprintf("https://www.okx.com/api/v5/market/ticker?instId=%s", instId), nil)
		if err != nil {
			logger.Errorf("TonapiLogic.Ticker error:%s", err.Error())
			return 0, errors.NewResponseError(constant.UnknownError, err)
		}

		var ticker TickerResponse

		if err := json.Unmarshal(body, &ticker); err != nil {
			logger.Errorf("TonapiLogic.Ticker error:%s", err.Error())
			return 0, errors.NewResponseError(constant.UnknownError, err)
		}

		if ticker.Code == "0" {
			for _, data := range ticker.Data {
				price := cast.ToFloat64(data.Last)
				if err := rb.Client().Set(ctx, instId, price, time.Minute).Err(); err != nil {
					logger.Errorf("TonapiLogic.Ticker error:%s", err.Error())
					return 0, err
				}
				return price, nil
			}
		} else {
			logger.Errorf("TonapiLogic.Ticker error:%s", fmt.Errorf("error from API: %s", ticker.Code))
		}
		return 0, errors.NewResponseError(constant.UnknownError, fmt.Errorf("error from API: %s", ticker.Code))
	}

	return cast.ToFloat64(result), nil
}
