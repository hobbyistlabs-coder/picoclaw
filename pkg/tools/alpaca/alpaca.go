package alpaca

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"jane/pkg/logger"
	"jane/pkg/tools"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/patrickmn/go-cache"
	"github.com/shopspring/decimal"
)

// AlpacaAction represents the main action types supported
type AlpacaAction string

const (
	ActionEquity          AlpacaAction = "equity"
	ActionPrice           AlpacaAction = "price"
	ActionQuote           AlpacaAction = "quote"
	ActionSnapshot        AlpacaAction = "snapshot"
	ActionSMA             AlpacaAction = "sma"
	ActionHistorical      AlpacaAction = "historical"
	ActionBars            AlpacaAction = "bars"
	ActionPortfolio       AlpacaAction = "portfolio"
	ActionOrderCreate     AlpacaAction = "order_create"
	ActionOrderCancel     AlpacaAction = "order_cancel"
	ActionPositionGet     AlpacaAction = "position_get"
	ActionPositionsList   AlpacaAction = "positions_list"
	ActionAccount         AlpacaAction = "account"
	ActionClock           AlpacaAction = "clock"
	ActionCalendar        AlpacaAction = "calendar"
	ActionWatchlists      AlpacaAction = "watchlists"
	ActionWatchlistGet    AlpacaAction = "watchlist_get"
	ActionOptionContracts AlpacaAction = "option_contracts"
	ActionOptionSnapshot  AlpacaAction = "option_snapshot"
	ActionOrderBook       AlpacaAction = "order_book"
	ActionCryptoPrice     AlpacaAction = "crypto_price"
	ActionCryptoQuote     AlpacaAction = "crypto_quote"
	ActionCryptoSnapshot  AlpacaAction = "crypto_snapshot"
	ActionCryptoBars      AlpacaAction = "crypto_bars"
)

type AlpacaTool struct {
	client     *alpaca.Client
	marketData *marketdata.Client
	keyID      string
	secretKey  string
	baseURL    string
	configured bool
	cache      *cache.Cache
	cacheTTL   time.Duration
	mu         sync.Mutex
	lastCall   map[string]time.Time
	maxCalls   int
	interval   time.Duration
}

// NewAlpacaTool creates a new Alpaca tool instance
func NewAlpacaTool(keyID, secretKey, baseURL string) *AlpacaTool {
	keyID, secretKey, baseURL = resolveConfig(keyID, secretKey, baseURL)
	configured := keyID != "" && secretKey != ""

	// The trading client uses the configured baseURL (e.g. paper or live trading endpoint).
	client := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    keyID,
		APISecret: secretKey,
		BaseURL:   baseURL,
	})

	// The market data client MUST NOT receive the trading base URL.
	// Passing a trading URL (e.g. https://paper-api.alpaca.markets) here causes all
	// market data calls (quotes, snapshots, bars) to return HTTP 404 because that host
	// does not serve the /v2/stocks data routes. Leave BaseURL empty so the SDK
	// defaults to https://data.alpaca.markets.
	dataURL := resolveDataURL()
	if baseURL != "" && dataURL == "" && strings.Contains(baseURL, "data.alpaca.markets") {
		// Caller explicitly passed a data URL via the shared baseURL field — honour it.
		dataURL = baseURL
	}

	marketData := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    keyID,
		APISecret: secretKey,
		BaseURL:   dataURL, // empty → SDK default (https://data.alpaca.markets)
	})

	cacheTTL := 30 * time.Second
	maxCalls := 100
	interval := 60 * time.Second
	cacheImpl := cache.New(cacheTTL, 10*time.Minute)

	maskedKey := ""
	if len(keyID) > 4 {
		maskedKey = keyID[:4] + "****"
	}
	logger.InfoCF("alpaca", "AlpacaTool initialised", map[string]any{
		"configured":  configured,
		"key_prefix":  maskedKey,
		"trading_url": baseURL,
		"data_url":    dataURL,
	})

	return &AlpacaTool{
		client:     client,
		marketData: marketData,
		keyID:      keyID,
		secretKey:  secretKey,
		baseURL:    baseURL,
		configured: configured,
		cache:      cacheImpl,
		cacheTTL:   cacheTTL,
		maxCalls:   maxCalls,
		interval:   interval,
		lastCall:   make(map[string]time.Time),
	}
}

func (t *AlpacaTool) Name() string {
	return "alpaca_finance"
}

func (t *AlpacaTool) Description() string {
	return "Provides comprehensive financial market data, account management, and trading operations via Alpaca API. Supports real-time quotes, historical data, portfolio analysis, order execution, options contracts, and L2 order book analysis."
}

func (t *AlpacaTool) RequiresApproval() bool {
	return false
}

func (t *AlpacaTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "The action to perform.",
				"enum": []string{
					string(ActionEquity), string(ActionPrice), string(ActionQuote), string(ActionSnapshot),
					string(ActionSMA), string(ActionHistorical), string(ActionBars),
					string(ActionPortfolio), string(ActionOrderCreate), string(ActionOrderCancel),
					string(ActionPositionGet), string(ActionPositionsList), string(ActionAccount), string(ActionClock),
					string(ActionCalendar), string(ActionWatchlists), string(ActionWatchlistGet),
					string(ActionOptionContracts), string(ActionOptionSnapshot), string(ActionOrderBook),
					string(ActionCryptoPrice), string(ActionCryptoQuote), string(ActionCryptoSnapshot), string(ActionCryptoBars),
				},
			},
			"symbol":            map[string]any{"type": "string", "description": "Stock, crypto, or option symbol (e.g., AAPL, BTC/USD, AAPL240621C00150000)."},
			"underlying_symbol": map[string]any{"type": "string", "description": "Underlying stock symbol for options lookup."},
			"expiration_date":   map[string]any{"type": "string", "description": "Filter for options (YYYY-MM-DD)."},
			"watchlist_id":      map[string]any{"type": "string", "description": "Watchlist ID or name."},
			"qty":               map[string]any{"type": "number", "description": "Quantity of shares/contracts."},
			"order_type":        map[string]any{"type": "string", "enum": []string{"market", "limit", "stop", "stop_limit"}, "default": "market"},
			"side":              map[string]any{"type": "string", "enum": []string{"buy", "sell"}},
			"limit_price":       map[string]any{"type": "number"},
			"stop_price":        map[string]any{"type": "number"},
			"time_in_force":     map[string]any{"type": "string", "enum": []string{"day", "gtc", "opg", "ioc", "fok"}, "default": "day"},
			"start_date":        map[string]any{"type": "string", "description": "Start date (YYYY-MM-DD or RFC3339)."},
			"end_date":          map[string]any{"type": "string", "description": "End date (YYYY-MM-DD or RFC3339)."},
			"period":            map[string]any{"type": "string", "default": "1M"},
			"interval":          map[string]any{"type": "string", "default": "1D"},
			"timeframe":         map[string]any{"type": "string", "description": "Timeframe for bars: '1Min', '5Min', '15Min', '1Hour', '1Day', '1Week', '1Month'. Defaults to '1Day'."},
			"limit":             map[string]any{"type": "number", "description": "Number of bars to return. Defaults to 10."},
			"order_status":      map[string]any{"type": "string", "description": "Filter orders by status: 'open', 'closed', 'all'. Defaults to 'open'."},
		},
		"required": []string{"action"},
	}
}

func (t *AlpacaTool) Execute(ctx context.Context, args map[string]any) *tools.ToolResult {
	if !t.configured {
		logger.WarnCCtx(ctx, "alpaca", "Execute called but tool is not configured")
		return tools.ErrorResult("alpaca is not configured; set tools.alpaca in config, export ALPACA_API_KEY/ALPACA_SECRET_KEY, or add .env.alpaca")
	}

	action, _ := args["action"].(string)
	symbol, _ := args["symbol"].(string)

	logger.InfoCFCtx(ctx, "alpaca", "Execute", map[string]any{
		"action": action,
		"symbol": symbol,
	})

	if !t.checkRateLimit(action) {
		logger.WarnCFCtx(ctx, "alpaca", "rate limit exceeded", map[string]any{"action": action})
		return tools.ErrorResult("Rate limit exceeded")
	}

	switch AlpacaAction(action) {
	// Account & trading
	case ActionAccount:
		return t.getAccount()
	case ActionEquity:
		return t.getEquity()
	case ActionPortfolio:
		return t.getPortfolio()
	case ActionPositionsList:
		return t.listPositions()
	case ActionPositionGet:
		return t.getPosition(args)
	case ActionOrderCreate:
		return t.createOrder(args)
	case ActionOrderCancel:
		return t.cancelOrder(args)
	case ActionClock:
		return t.getClock()
	case ActionCalendar:
		return t.getCalendar(args)
	case ActionWatchlists:
		return t.listWatchlists()
	case ActionWatchlistGet:
		return t.getWatchlist(args)

	// Stock market data
	case ActionPrice:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getPrice(symbol)
	case ActionQuote:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getQuote(symbol)
	case ActionSnapshot:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getSnapshot(symbol)
	case ActionHistorical, ActionBars:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getBars(symbol, args)
	case ActionSMA:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getSMA(symbol, args)

	// Options
	case ActionOptionContracts:
		return t.getOptionContracts(args)
	case ActionOptionSnapshot:
		return t.getOptionSnapshot(args)
	case ActionOrderBook:
		return t.getOrderBook(args)

	// Crypto market data
	case ActionCryptoPrice:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getCryptoPrice(symbol)
	case ActionCryptoQuote:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getCryptoQuote(symbol)
	case ActionCryptoSnapshot:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getCryptoSnapshot(symbol)
	case ActionCryptoBars:
		if symbol == "" {
			return tools.ErrorResult("missing 'symbol' parameter")
		}
		return t.getCryptoBars(symbol, args)

	default:
		return tools.ErrorResult(fmt.Sprintf("unknown action: %s", action))
	}
}

// --- Rate Limiting ---

func (t *AlpacaTool) checkRateLimit(action string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	key := fmt.Sprintf("%s-%s", action, now.Format("2006-01-02"))
	if _, exists := t.lastCall[key]; !exists {
		t.lastCall[key] = now
		return true
	}
	t.lastCall[key] = now
	return true
}

// --- Account & Trading ---

func (t *AlpacaTool) getAccount() *tools.ToolResult {
	acct, err := t.client.GetAccount()
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	jsonData, _ := json.MarshalIndent(acct, "", "  ")
	return tools.UserResult(string(jsonData))
}

func (t *AlpacaTool) getEquity() *tools.ToolResult {
	acct, err := t.client.GetAccount()
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	return tools.UserResult(fmt.Sprintf("Equity: $%s | Buying Power: $%s", acct.Equity, acct.BuyingPower))
}

func (t *AlpacaTool) getPortfolio() *tools.ToolResult {
	acct, err := t.client.GetAccount()
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	return tools.UserResult(fmt.Sprintf("Portfolio Value: $%s", acct.PortfolioValue))
}

func (t *AlpacaTool) listPositions() *tools.ToolResult {
	positions, err := t.client.GetPositions()
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get positions: %v", err))
	}
	if len(positions) == 0 {
		return tools.UserResult("No open positions.")
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Open Positions (%d):\n", len(positions)))
	for _, p := range positions {
		sb.WriteString(fmt.Sprintf(
			"\n%s: %s shares @ $%s avg | Market Value: $%s | P&L: $%s (%s%%)",
			p.Symbol, p.Qty.String(), p.AvgEntryPrice.String(),
			p.MarketValue.String(), p.UnrealizedPL.String(), p.UnrealizedPLPC.String(),
		))
	}
	return tools.UserResult(sb.String())
}

func (t *AlpacaTool) getPosition(args map[string]any) *tools.ToolResult {
	symbol, _ := args["symbol"].(string)
	pos, err := t.client.GetPosition(symbol)
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	return tools.UserResult(fmt.Sprintf("%s: %s shares @ $%s", pos.Symbol, pos.Qty, pos.AvgEntryPrice))
}

func (t *AlpacaTool) createOrder(args map[string]any) *tools.ToolResult {
	symbol, _ := args["symbol"].(string)
	qtyVal, _ := args["qty"].(float64)
	side, _ := args["side"].(string)
	qty := decimal.NewFromFloat(qtyVal)
	order, err := t.client.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      strings.ToUpper(symbol),
		Qty:         &qty,
		Side:        alpaca.Side(side),
		Type:        alpaca.OrderType(args["order_type"].(string)),
		TimeInForce: alpaca.TimeInForce(args["time_in_force"].(string)),
	})
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	return tools.UserResult(fmt.Sprintf("Order %s for %s submitted", order.ID, symbol))
}

func (t *AlpacaTool) cancelOrder(args map[string]any) *tools.ToolResult {
	id, _ := args["order_id"].(string)
	if err := t.client.CancelOrder(id); err != nil {
		return tools.ErrorResult(err.Error())
	}
	return tools.UserResult("Order cancelled")
}

func (t *AlpacaTool) getClock() *tools.ToolResult {
	clock, err := t.client.GetClock()
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get market clock: %v", err))
	}
	status := "CLOSED"
	if clock.IsOpen {
		status = "OPEN"
	}
	return tools.UserResult(fmt.Sprintf(
		"Market is %s\nNext Open: %s\nNext Close: %s",
		status,
		clock.NextOpen.Format(time.RFC3339),
		clock.NextClose.Format(time.RFC3339),
	))
}

func (t *AlpacaTool) getCalendar(args map[string]any) *tools.ToolResult {
	cal, err := t.client.GetCalendar(alpaca.GetCalendarRequest{})
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	return tools.UserResult(fmt.Sprintf("Next trading day: %s", cal[0].Date))
}

func (t *AlpacaTool) listWatchlists() *tools.ToolResult {
	wls, err := t.client.GetWatchlists()
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	var data []map[string]any
	for _, wl := range wls {
		data = append(data, map[string]any{
			"id":         wl.ID,
			"name":       wl.Name,
			"created_at": wl.CreatedAt,
			"updated_at": wl.UpdatedAt,
		})
	}
	res, _ := json.MarshalIndent(data, "", "  ")
	return tools.UserResult(string(res))
}

func (t *AlpacaTool) getWatchlist(args map[string]any) *tools.ToolResult {
	id, _ := args["watchlist_id"].(string)
	wl, err := t.client.GetWatchlist(id)
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	res, _ := json.MarshalIndent(wl, "", "  ")
	return tools.UserResult(string(res))
}

// --- Stock Market Data ---

func (t *AlpacaTool) getPrice(symbol string) *tools.ToolResult {
	trade, err := t.marketData.GetLatestTrade(strings.ToUpper(symbol), marketdata.GetLatestTradeRequest{})
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get latest trade for %s: %v", symbol, err))
	}
	return tools.UserResult(fmt.Sprintf("Latest price for %s: $%.2f (size: %d, at %s)",
		symbol, trade.Price, trade.Size, trade.Timestamp.Format(time.RFC3339)))
}

func (t *AlpacaTool) getQuote(symbol string) *tools.ToolResult {
	logger.DebugCF("alpaca", "GetLatestQuote", map[string]any{"symbol": symbol})
	quote, err := t.marketData.GetLatestQuote(strings.ToUpper(symbol), marketdata.GetLatestQuoteRequest{})
	if err != nil {
		logger.ErrorCF("alpaca", "GetLatestQuote failed", map[string]any{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return tools.ErrorResult(fmt.Sprintf("failed to get latest quote for %s: %v", symbol, err))
	}
	return tools.UserResult(fmt.Sprintf(
		"Quote for %s:\n  Bid: $%.2f x %d\n  Ask: $%.2f x %d\n  Spread: $%.2f\n  At: %s",
		symbol, quote.BidPrice, quote.BidSize, quote.AskPrice, quote.AskSize,
		quote.AskPrice-quote.BidPrice, quote.Timestamp.Format(time.RFC3339),
	))
}

func (t *AlpacaTool) getSnapshot(symbol string) *tools.ToolResult {
	logger.DebugCF("alpaca", "GetSnapshot", map[string]any{"symbol": symbol})
	snap, err := t.marketData.GetSnapshot(strings.ToUpper(symbol), marketdata.GetSnapshotRequest{})
	if err != nil {
		logger.ErrorCF("alpaca", "GetSnapshot failed", map[string]any{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return tools.ErrorResult(fmt.Sprintf("failed to get snapshot for %s: %v", symbol, err))
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Snapshot for %s:", symbol))
	if snap.LatestTrade != nil {
		sb.WriteString(fmt.Sprintf("\n  Latest Trade: $%.2f (size: %d)", snap.LatestTrade.Price, snap.LatestTrade.Size))
	}
	if snap.LatestQuote != nil {
		sb.WriteString(fmt.Sprintf("\n  Bid: $%.2f x %d | Ask: $%.2f x %d",
			snap.LatestQuote.BidPrice, snap.LatestQuote.BidSize,
			snap.LatestQuote.AskPrice, snap.LatestQuote.AskSize))
	}
	if snap.DailyBar != nil {
		sb.WriteString(fmt.Sprintf("\n  Daily Bar: O:%.2f H:%.2f L:%.2f C:%.2f V:%d VWAP:%.2f",
			snap.DailyBar.Open, snap.DailyBar.High, snap.DailyBar.Low, snap.DailyBar.Close,
			snap.DailyBar.Volume, snap.DailyBar.VWAP))
	}
	if snap.PrevDailyBar != nil {
		sb.WriteString(fmt.Sprintf("\n  Prev Close: $%.2f", snap.PrevDailyBar.Close))
		if snap.DailyBar != nil && snap.PrevDailyBar.Close > 0 {
			change := snap.DailyBar.Close - snap.PrevDailyBar.Close
			changePct := (change / snap.PrevDailyBar.Close) * 100
			sb.WriteString(fmt.Sprintf(" | Change: $%.2f (%.2f%%)", change, changePct))
		}
	}
	return tools.UserResult(sb.String())
}

func (t *AlpacaTool) getBars(symbol string, args map[string]any) *tools.ToolResult {
	req := t.buildBarsRequest(args)
	logger.DebugCF("alpaca", "GetBars", map[string]any{
		"symbol":     symbol,
		"timeframe":  req.TimeFrame,
		"start":      req.Start,
		"end":        req.End,
		"totalLimit": req.TotalLimit,
	})
	bars, err := t.marketData.GetBars(strings.ToUpper(symbol), req)
	if err != nil {
		logger.ErrorCF("alpaca", "GetBars failed", map[string]any{
			"symbol": symbol,
			"error":  err.Error(),
		})
		return tools.ErrorResult(fmt.Sprintf("failed to get bars for %s: %v", symbol, err))
	}
	if len(bars) == 0 {
		return tools.ErrorResult(fmt.Sprintf("no bar data found for %s", symbol))
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Bars for %s (%d results):\n", symbol, len(bars)))
	for _, bar := range bars {
		sb.WriteString(fmt.Sprintf(
			"\n%s | O:%.2f H:%.2f L:%.2f C:%.2f V:%d VWAP:%.2f",
			bar.Timestamp.Format("2006-01-02 15:04"),
			bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, bar.VWAP,
		))
	}
	return tools.UserResult(sb.String())
}

func (t *AlpacaTool) getSMA(symbol string, args map[string]any) *tools.ToolResult {
	limit := intArg(args, "limit", 10)
	req := marketdata.GetBarsRequest{
		TimeFrame:  marketdata.OneDay,
		Start:      time.Now().AddDate(0, 0, -(limit*2 + 5)),
		End:        time.Now(),
		TotalLimit: limit,
		Adjustment: marketdata.AdjustmentAll,
	}
	bars, err := t.marketData.GetBars(strings.ToUpper(symbol), req)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get bars for %s: %v", symbol, err))
	}
	if len(bars) == 0 {
		return tools.ErrorResult(fmt.Sprintf("no data found for %s", symbol))
	}
	var sum float64
	for _, bar := range bars {
		sum += bar.Close
	}
	sma := sum / float64(len(bars))
	return tools.UserResult(fmt.Sprintf("%d-Day SMA for %s: $%.2f (based on %d bars)", limit, symbol, sma, len(bars)))
}

// --- Options & Order Book ---

func (t *AlpacaTool) getOptionContracts(args map[string]any) *tools.ToolResult {
	underlying, _ := args["underlying_symbol"].(string)
	if underlying == "" {
		return tools.ErrorResult("underlying_symbol required")
	}

	url := fmt.Sprintf("%s/v2/options/contracts?underlying_symbols=%s", t.baseURL, strings.ToUpper(underlying))
	if exp, ok := args["expiration_date"].(string); ok && exp != "" {
		url += "&expiration_date=" + exp
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	req.Header.Add("APCA-API-KEY-ID", t.keyID)
	req.Header.Add("APCA-API-SECRET-KEY", t.secretKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return tools.ErrorResult(fmt.Sprintf("failed to get options contracts: API returned status %d", res.StatusCode))
	}

	var response struct {
		OptionContracts []struct {
			Symbol         string `json:"symbol"`
			StrikePrice    string `json:"strike_price"`
			ExpirationDate string `json:"expiration_date"`
			Type           string `json:"type"`
		} `json:"option_contracts"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return tools.ErrorResult("failed to parse options response: " + err.Error())
	}

	if len(response.OptionContracts) == 0 {
		return tools.UserResult(fmt.Sprintf("No option contracts found for %s", underlying))
	}

	var resStr []string
	for i, c := range response.OptionContracts {
		if i >= 50 {
			resStr = append(resStr, "... and more")
			break
		}
		resStr = append(resStr, fmt.Sprintf("- %s: Strike $%s, Expires %s, Type %s", c.Symbol, c.StrikePrice, c.ExpirationDate, c.Type))
	}
	return tools.UserResult(fmt.Sprintf("Option Contracts for %s:\n%s", underlying, strings.Join(resStr, "\n")))
}

func (t *AlpacaTool) getOptionSnapshot(args map[string]any) *tools.ToolResult {
	symbol, _ := args["symbol"].(string)
	snapshot, err := t.marketData.GetOptionSnapshot(strings.ToUpper(symbol), marketdata.GetOptionSnapshotRequest{})
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	return tools.UserResult(fmt.Sprintf("%s Snapshot: IV: %v | Delta: %v | Price: $%.2f",
		symbol, snapshot.ImpliedVolatility, snapshot.Greeks.Delta, snapshot.LatestTrade.Price))
}

func (t *AlpacaTool) getOrderBook(args map[string]any) *tools.ToolResult {
	symbol, _ := args["symbol"].(string)
	if symbol == "" {
		return tools.ErrorResult("symbol required")
	}

	url := fmt.Sprintf("https://data.alpaca.markets/v1beta3/crypto/us/latest/orderbooks?symbols=%s", strings.ToUpper(symbol))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	req.Header.Add("APCA-API-KEY-ID", t.keyID)
	req.Header.Add("APCA-API-SECRET-KEY", t.secretKey)
	req.Header.Add("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return tools.ErrorResult(err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return tools.ErrorResult(fmt.Sprintf("failed to get orderbook (note: REST L2 orderbooks are Crypto-only on Alpaca): status %d", res.StatusCode))
	}

	var response struct {
		Orderbooks map[string]struct {
			Bids []struct {
				Price float64 `json:"p"`
				Size  float64 `json:"s"`
			} `json:"b"`
			Asks []struct {
				Price float64 `json:"p"`
				Size  float64 `json:"s"`
			} `json:"a"`
		} `json:"orderbooks"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return tools.ErrorResult("failed to parse orderbook response: " + err.Error())
	}

	ob, exists := response.Orderbooks[strings.ToUpper(symbol)]
	if !exists || len(ob.Bids) == 0 || len(ob.Asks) == 0 {
		return tools.UserResult(fmt.Sprintf("Order book is empty or unavailable for %s", symbol))
	}

	summary := fmt.Sprintf("L2 Order Book for %s:\nBest Ask: $%.2f (Size: %v)\nBest Bid: $%.2f (Size: %v)",
		symbol, ob.Asks[0].Price, ob.Asks[0].Size, ob.Bids[0].Price, ob.Bids[0].Size)
	return tools.UserResult(summary)
}

// --- Crypto Market Data ---

func (t *AlpacaTool) getCryptoPrice(symbol string) *tools.ToolResult {
	trade, err := t.marketData.GetLatestCryptoTrade(symbol, marketdata.GetLatestCryptoTradeRequest{})
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get latest crypto trade for %s: %v", symbol, err))
	}
	return tools.UserResult(fmt.Sprintf("Latest price for %s: $%.2f (size: %.6f, at %s)",
		symbol, trade.Price, trade.Size, trade.Timestamp.Format(time.RFC3339)))
}

func (t *AlpacaTool) getCryptoQuote(symbol string) *tools.ToolResult {
	quote, err := t.marketData.GetLatestCryptoQuote(symbol, marketdata.GetLatestCryptoQuoteRequest{})
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get latest crypto quote for %s: %v", symbol, err))
	}
	return tools.UserResult(fmt.Sprintf(
		"Crypto Quote for %s:\n  Bid: $%.2f x %.6f\n  Ask: $%.2f x %.6f\n  Spread: $%.2f\n  At: %s",
		symbol, quote.BidPrice, quote.BidSize, quote.AskPrice, quote.AskSize,
		quote.AskPrice-quote.BidPrice, quote.Timestamp.Format(time.RFC3339),
	))
}

func (t *AlpacaTool) getCryptoSnapshot(symbol string) *tools.ToolResult {
	snap, err := t.marketData.GetCryptoSnapshot(symbol, marketdata.GetCryptoSnapshotRequest{})
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get crypto snapshot for %s: %v", symbol, err))
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Crypto Snapshot for %s:", symbol))
	if snap.LatestTrade != nil {
		sb.WriteString(fmt.Sprintf("\n  Latest Trade: $%.2f (size: %.6f)", snap.LatestTrade.Price, snap.LatestTrade.Size))
	}
	if snap.LatestQuote != nil {
		sb.WriteString(fmt.Sprintf("\n  Bid: $%.2f x %.6f | Ask: $%.2f x %.6f",
			snap.LatestQuote.BidPrice, snap.LatestQuote.BidSize,
			snap.LatestQuote.AskPrice, snap.LatestQuote.AskSize))
	}
	if snap.DailyBar != nil {
		sb.WriteString(fmt.Sprintf("\n  Daily Bar: O:%.2f H:%.2f L:%.2f C:%.2f V:%.2f VWAP:%.2f",
			snap.DailyBar.Open, snap.DailyBar.High, snap.DailyBar.Low, snap.DailyBar.Close,
			snap.DailyBar.Volume, snap.DailyBar.VWAP))
	}
	if snap.PrevDailyBar != nil {
		sb.WriteString(fmt.Sprintf("\n  Prev Close: $%.2f", snap.PrevDailyBar.Close))
		if snap.DailyBar != nil && snap.PrevDailyBar.Close > 0 {
			change := snap.DailyBar.Close - snap.PrevDailyBar.Close
			changePct := (change / snap.PrevDailyBar.Close) * 100
			sb.WriteString(fmt.Sprintf(" | Change: $%.2f (%.2f%%)", change, changePct))
		}
	}
	return tools.UserResult(sb.String())
}

func (t *AlpacaTool) getCryptoBars(symbol string, args map[string]any) *tools.ToolResult {
	tf := parseTimeFrame(stringArg(args, "timeframe", "1Day"))
	limit := intArg(args, "limit", 10)
	start, end := parseDateRange(args, 30)

	req := marketdata.GetCryptoBarsRequest{
		TimeFrame:  tf,
		Start:      start,
		End:        end,
		TotalLimit: limit,
	}
	bars, err := t.marketData.GetCryptoBars(symbol, req)
	if err != nil {
		return tools.ErrorResult(fmt.Sprintf("failed to get crypto bars for %s: %v", symbol, err))
	}
	if len(bars) == 0 {
		return tools.ErrorResult(fmt.Sprintf("no crypto bar data found for %s", symbol))
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Crypto Bars for %s (%d results):\n", symbol, len(bars)))
	for _, bar := range bars {
		sb.WriteString(fmt.Sprintf(
			"\n%s | O:%.2f H:%.2f L:%.2f C:%.2f V:%.2f VWAP:%.2f",
			bar.Timestamp.Format("2006-01-02 15:04"),
			bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, bar.VWAP,
		))
	}
	return tools.UserResult(sb.String())
}

// --- Helpers ---

func (t *AlpacaTool) buildBarsRequest(args map[string]any) marketdata.GetBarsRequest {
	tf := parseTimeFrame(stringArg(args, "timeframe", stringArg(args, "interval", "1Day")))
	limit := intArg(args, "limit", 10)
	start, end := parseDateRange(args, 30)

	return marketdata.GetBarsRequest{
		TimeFrame:  tf,
		Start:      start,
		End:        end,
		TotalLimit: limit,
		Adjustment: marketdata.AdjustmentAll,
	}
}

func parseTimeFrame(s string) marketdata.TimeFrame {
	switch strings.ToLower(s) {
	case "1min", "1m":
		return marketdata.OneMin
	case "5min":
		return marketdata.NewTimeFrame(5, marketdata.Min)
	case "15min":
		return marketdata.NewTimeFrame(15, marketdata.Min)
	case "1hour", "1h":
		return marketdata.OneHour
	case "1day", "1d", "":
		return marketdata.OneDay
	case "1week":
		return marketdata.OneWeek
	case "1month":
		return marketdata.OneMonth
	default:
		return marketdata.OneDay
	}
}

func parseDateRange(args map[string]any, defaultDaysBack int) (time.Time, time.Time) {
	end := time.Now()
	start := end.AddDate(0, 0, -defaultDaysBack)

	// Support both "start"/"end" and "start_date"/"end_date" param names
	startStr := stringArg(args, "start", stringArg(args, "start_date", ""))
	endStr := stringArg(args, "end", stringArg(args, "end_date", ""))

	if startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = parsed
		} else if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			start = parsed
		}
	}
	if endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = parsed
		} else if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			end = parsed
		}
	}
	return start, end
}

func stringArg(args map[string]any, key, defaultVal string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return defaultVal
}

func intArg(args map[string]any, key string, defaultVal int) int {
	if v, ok := args[key].(float64); ok {
		return int(v)
	}
	return defaultVal
}
