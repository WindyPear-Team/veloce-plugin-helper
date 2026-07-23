package pluginhelper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"unsafe"
)

const initialHostResponseBytes = 4096

type HostError struct {
	Code    string
	Message string
}

func (e *HostError) Error() string {
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type Client struct {
	UserID uint
}

type Balance struct {
	Amount string
}

type WalletSettlement struct {
	IdempotencyKey string
	Debit          string
	Credit         string
	ReferenceType  string
	ReferenceID    string
	Description    string
	Metadata       map[string]any
}

type WalletSettlementResult struct {
	Replay        bool
	TransactionID uint
	Debit         string
	Credit        string
	BalanceBefore string
	BalanceAfter  string
	Metadata      map[string]any
}

func NewClient(userID uint) *Client { return &Client{UserID: userID} }

func (c *Client) Balance(ctx context.Context) (Balance, error) {
	var response struct {
		Balance string `json:"balance"`
	}
	if err := c.call(ctx, "wallet.balance", nil, &response); err != nil {
		return Balance{}, err
	}
	return Balance{Amount: response.Balance}, nil
}

func (c *Client) Settle(ctx context.Context, input WalletSettlement) (WalletSettlementResult, error) {
	request := map[string]any{
		"idempotency_key": input.IdempotencyKey,
		"debit":           input.Debit,
		"credit":          input.Credit,
		"reference_type":  input.ReferenceType,
		"reference_id":    input.ReferenceID,
		"description":     input.Description,
		"metadata":        input.Metadata,
	}
	var response struct {
		Replay        bool           `json:"replay"`
		TransactionID uint           `json:"transaction_id"`
		Debit         string         `json:"debit"`
		Credit        string         `json:"credit"`
		BalanceBefore string         `json:"balance_before"`
		BalanceAfter  string         `json:"balance_after"`
		Metadata      map[string]any `json:"metadata"`
	}
	if err := c.call(ctx, "wallet.settle", request, &response); err != nil {
		return WalletSettlementResult{}, err
	}
	return WalletSettlementResult{Replay: response.Replay, TransactionID: response.TransactionID, Debit: response.Debit, Credit: response.Credit, BalanceBefore: response.BalanceBefore, BalanceAfter: response.BalanceAfter, Metadata: response.Metadata}, nil
}

func (c *Client) Settlement(ctx context.Context, idempotencyKey string) (WalletSettlementResult, bool, error) {
	var response struct {
		Found         bool           `json:"found"`
		Replay        bool           `json:"replay"`
		TransactionID uint           `json:"transaction_id"`
		Debit         string         `json:"debit"`
		Credit        string         `json:"credit"`
		BalanceBefore string         `json:"balance_before"`
		BalanceAfter  string         `json:"balance_after"`
		Metadata      map[string]any `json:"metadata"`
	}
	if err := c.call(ctx, "wallet.transaction", map[string]any{"idempotency_key": idempotencyKey}, &response); err != nil {
		return WalletSettlementResult{}, false, err
	}
	if !response.Found {
		return WalletSettlementResult{}, false, nil
	}
	return WalletSettlementResult{Replay: response.Replay, TransactionID: response.TransactionID, Debit: response.Debit, Credit: response.Credit, BalanceBefore: response.BalanceBefore, BalanceAfter: response.BalanceAfter, Metadata: response.Metadata}, true, nil
}

func (c *Client) KVGet(ctx context.Context, key string, destination any) (bool, error) {
	var response struct {
		Found bool            `json:"found"`
		Value json.RawMessage `json:"value"`
	}
	if err := c.call(ctx, "plugin.kv.get", map[string]any{"key": key}, &response); err != nil {
		return false, err
	}
	if !response.Found {
		return false, nil
	}
	if destination == nil {
		return true, nil
	}
	if err := json.Unmarshal(response.Value, destination); err != nil {
		return false, fmt.Errorf("decode plugin KV %q: %w", key, err)
	}
	return true, nil
}

func (c *Client) KVPut(ctx context.Context, key string, value any) error {
	return c.call(ctx, "plugin.kv.put", map[string]any{"key": key, "value": value}, nil)
}

func (c *Client) KVDelete(ctx context.Context, key string) error {
	return c.call(ctx, "plugin.kv.delete", map[string]any{"key": key}, nil)
}

func (c *Client) Log(ctx context.Context, level, message string, metadata any) error {
	return c.call(ctx, "plugin.log", map[string]any{"level": level, "message": message, "metadata": metadata}, nil)
}

func (c *Client) call(ctx context.Context, operation string, request map[string]any, destination any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	op := []byte(operation)
	rawRequest, err := json.Marshal(requestOrEmpty(request))
	if err != nil {
		return err
	}
	response := make([]byte, initialHostResponseBytes)
	for attempt := 0; attempt < 2; attempt++ {
		packed := hostCall(pointer(op), uint32(len(op)), pointer(rawRequest), uint32(len(rawRequest)), pointer(response), uint32(len(response)))
		status := uint32(packed >> 32)
		length := uint32(packed)
		if status == 1 {
			if length == 0 || length > 1<<20 {
				return errors.New("plugin host returned an invalid response size")
			}
			response = make([]byte, length)
			continue
		}
		if length > uint32(len(response)) {
			return errors.New("plugin host returned an invalid response length")
		}
		var envelope struct {
			OK      bool   `json:"ok"`
			Code    string `json:"code"`
			Message string `json:"error"`
		}
		if err := json.Unmarshal(response[:length], &envelope); err != nil {
			return fmt.Errorf("decode plugin host response: %w", err)
		}
		if status != 0 || !envelope.OK {
			return &HostError{Code: envelope.Code, Message: envelope.Message}
		}
		if destination != nil {
			if err := json.Unmarshal(response[:length], destination); err != nil {
				return fmt.Errorf("decode plugin host response: %w", err)
			}
		}
		return nil
	}
	return errors.New("plugin host response remained too large after retry")
}

func requestOrEmpty(request map[string]any) map[string]any {
	if request == nil {
		return map[string]any{}
	}
	return request
}

func pointer(value []byte) uint32 {
	if len(value) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&value[0])))
}
