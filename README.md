# veloce-plugin-helper

`github.com/WindyPear-Team/veloce-plugin-helper` 是 Veloce WASM 插件 SDK。它把 stdin/stdout JSON ABI 和宿主调用封装起来，插件开发者通常只需要保留四个很薄的 WASM 导出函数。

## 最小结构

```go
package main

import plugin "github.com/WindyPear-Team/veloce-plugin-helper"

var app = plugin.New(plugin.Manifest{
    ID: "my-plugin",
    Name: "My Plugin",
    Version: "0.1.0",
})

func init() {
    app.Action("hello", func(ctx *plugin.ActionContext) (any, error) {
        return map[string]any{"message": "hello " + ctx.Values.String("name", "world")}, nil
    })
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
```

使用 TinyGo 构建：

```bash
tinygo build -target=wasi -o my-plugin.wasm .
```

## 钱包

manifest 中声明：

```go
Permissions: []string{
    plugin.PermissionWalletBalanceRead,
    plugin.PermissionWalletBalanceDebit,
    plugin.PermissionWalletBalanceCredit,
}
```

Action 中可以直接调用：

```go
balance, err := ctx.Balance()
settlement, err := ctx.Settle(plugin.WalletSettlement{
    Debit: "1",
    Credit: "5",
    ReferenceType: "lottery_draw",
    ReferenceID: ctx.RequestID,
    Metadata: map[string]any{"prize": "5"},
})
```

`ActionContext` 还直接提供 `KVGet`、`KVPut`、`KVDelete` 和 `Log`，无需手工创建宿主客户端或传递用户 ID。

`Settle` 在 community 后端中以一个数据库事务完成余额更新和流水写入。相同用户、插件来源和幂等键重复调用会返回原结算；参数变化会返回 `idempotency_conflict`。

对于包含随机结果的 Action，可在抽奖前调用 `ctx.PreviousSettlement()`。它会按当前 `RequestID` 读取原流水及 metadata，使超时重试恢复第一次已经提交的奖品，而不是再次随机。

金额使用字符串，最多支持 6 位小数。`Debit` 和 `Credit` 都必须是非负数；余额不足会返回 `insufficient_balance`。

## KV 和日志

声明 `plugin.kv.read` 或 `plugin.kv.write` 后可使用 `KVGet`、`KVPut`、`KVDelete`。KV 自动按用户和插件 ID 隔离。日志可通过 `ctx.Log` 写入 community 的插件日志库。

## 前端声明

`Page`、`Card`、`Form`、`Input`、`Button` 等 helper 会生成当前 Web 前端支持的声明式页面结构。更复杂的节点仍可以直接使用 `map[string]any`。

## 注意事项

- 插件只能调用 manifest 声明并且 community 认可的宿主能力，不能直接访问数据库或网络。
- 随机抽奖应使用 `SecureIndex`，并把抽奖结果放入结算 metadata；重试时使用 `ctx.RequestID` 作为幂等键。
- 当前 SDK 依赖 TinyGo 的 `//export` 和 `//go:wasmimport` 支持。
