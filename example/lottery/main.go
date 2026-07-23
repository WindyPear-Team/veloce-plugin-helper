package main

import (
	"fmt"

	plugin "github.com/WindyPear-Team/veloce-plugin-helper"
)

var app = plugin.New(plugin.Manifest{
	ID:          "balance-lottery-example",
	Name:        "余额抽奖示例",
	Version:     "0.1.0",
	Description: "A small example of an idempotent balance-funded lottery plugin.",
	Author:      "Veloce",
	Permissions: []string{plugin.PermissionWalletBalanceRead, plugin.PermissionWalletBalanceDebit, plugin.PermissionWalletBalanceCredit},
	Frontend: plugin.Frontend{
		Sidebar: []plugin.SidebarItem{{Label: "余额抽奖", Path: "draw"}},
		Routes:  []plugin.Route{{Path: "draw", Title: "余额抽奖", Page: plugin.Card("余额抽奖", plugin.Text("每次抽奖扣除 1 余额，奖品由后端原子结算。"), plugin.Button("draw", "立即抽奖"))}},
	},
	Settings: plugin.SettingsSchema{
		Type:   "form",
		Fields: []plugin.SettingsField{{Type: "input", Name: "entry_fee", Label: "抽奖费用", Default: "1"}},
	},
})

func init() {
	app.Action("draw", draw)
}

type prize struct {
	Label  string
	Amount string
}

func draw(ctx *plugin.ActionContext) (any, error) {
	if previous, found, err := ctx.PreviousSettlement(); err != nil {
		return nil, err
	} else if found {
		return drawResponse(previous), nil
	}
	prizes := []prize{{Label: "谢谢参与", Amount: "0"}, {Label: "小奖", Amount: "1"}, {Label: "大奖", Amount: "5"}}
	index, err := plugin.SecureIndex(len(prizes))
	if err != nil {
		return nil, err
	}
	selected := prizes[index]
	fee := ctx.Settings.String("entry_fee", "1")
	settlement, err := ctx.Settle(plugin.WalletSettlement{
		Debit:         fee,
		Credit:        selected.Amount,
		ReferenceType: "lottery_draw",
		ReferenceID:   ctx.RequestID,
		Description:   "余额抽奖",
		Metadata:      map[string]any{"prize": selected.Label, "prize_amount": selected.Amount},
	})
	if err != nil {
		if previous, found, lookupErr := ctx.PreviousSettlement(); lookupErr == nil && found {
			return drawResponse(previous), nil
		}
		return nil, err
	}
	return drawResponse(settlement), nil
}

func drawResponse(settlement plugin.WalletSettlementResult) map[string]any {
	selected := prize{Label: "未知奖品", Amount: settlement.Credit}
	label := selected.Label
	amount := selected.Amount
	if settlement.Metadata != nil {
		if value, ok := settlement.Metadata["prize"].(string); ok {
			label = value
		}
		if value, ok := settlement.Metadata["prize_amount"].(string); ok {
			amount = value
		}
	}
	return map[string]any{
		"message": fmt.Sprintf("抽奖结果：%s（%s）", label, amount),
		"prize":   label,
		"amount":  amount,
		"balance": settlement.BalanceAfter,
		"replay":  settlement.Replay,
	}
}

func main() {}

//export plugin_manifest
func manifest() { app.ExportManifest() }

//export plugin_init
func initPlugin() { app.ExportInit() }

//export plugin_handle_action
func action() { app.ExportAction() }

//export plugin_handle_hook
func hook() { app.ExportHook() }
