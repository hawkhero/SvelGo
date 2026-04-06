# SvelGo CLI 設計計劃：`svelgo new` + `svelgo add`

## 背景

目前建立新 SvelGo 專案需要手動操作 8 個步驟，每新增一個 component 需要手動修改 6 個地方。目標是讓開發者安裝一次 CLI 後，所有 boilerplate 全部自動化。

**目標工作流程：**
```bash
go install github.com/hawkhero/svelgo/cmd/svelgo@latest

svelgo new myapp
cd myapp && make dev    # 馬上能跑，有 Hello component

svelgo add Counter      # 自動生成 stubs + 更新 registrations + 跑 protoc/pbjs
# 開發者只需填寫 4 個地方的 app logic
```

---

## 新目錄結構（加在 framework repo 內）

```
SvelGo/
└── cmd/
    └── svelgo/
        ├── main.go
        ├── commands/
        │   ├── new.go
        │   ├── add.go
        │   └── shared.go          ← renderTemplate, runCmd, appendIfAbsent 等
        ├── templates/
        │   ├── templates.go       ← //go:embed + Parse()
        │   ├── new/               ← svelgo new 的 template 檔案
        │   │   ├── go.mod.tmpl
        │   │   ├── Makefile.tmpl
        │   │   ├── main.go.tmpl
        │   │   ├── embed.go.tmpl
        │   │   ├── app.proto.tmpl
        │   │   ├── package.json.tmpl
        │   │   ├── vite.config.ts.tmpl
        │   │   ├── index.html.tmpl
        │   │   ├── main.ts.tmpl
        │   │   ├── proto.ts.tmpl
        │   │   ├── registry.ts.tmpl
        │   │   └── Hello.svelte.tmpl
        │   └── add/
        │       ├── component.go.tmpl
        │       └── component.svelte.tmpl
        └── internal/
            └── gomod.go           ← ReadModuleName("go.mod")
```

---

## `svelgo new <name>` 做的事

1. 建立目錄結構 + 所有 boilerplate 檔案
2. 寫入 `svelgo.json`（標記 app root，供 `svelgo add` 辨識）
3. 執行 `go mod tidy`
4. 執行 `cd frontend && npm install`
5. 執行 protoc + pbjs（生成 gen/app/app.pb.go + frontend/src/app_descriptor.json）
6. 印出 `Done! cd <name> && make dev`

**Flags：**
- `--local /path/to/SvelGo`：go.mod 加 replace 指令，npm 用 `file:` 路徑

---

## `svelgo add <ComponentName>` 做的事

在 app root（含 `svelgo.json`）執行：

1. 生成 `<name>.go`（package main，Go struct stub + HandleEvent TODO）
2. Append `message <Name>State { // TODO: add fields }` 到 `proto/app.proto`
3. 生成 `frontend/src/components/<Name>.svelte`（Svelte stub）
4. Append import + `registerComponent('<Name>', <Name>)` 到 `registry.ts`
5. Append `registerComponentDecoder('<Name>', appRoot.lookupType('app.<Name>State'))` 到 `proto.ts`
6. 執行 protoc → 重新生成 `gen/app/app.pb.go`
7. 執行 pbjs → 重新生成 `frontend/src/app_descriptor.json`
8. 印出 next steps（4 個手動 TODO）

**重複防止：** `appendIfAbsent()` 先確認檔案中是否已含 sentinel 字串，避免重複登錄

---

## 關鍵設計決策

| 決策 | 選擇 | 原因 |
|------|------|------|
| CLI 依賴 | 無第三方（只用標準庫） | 減少依賴；只有 2 個命令不需要 cobra |
| Template 系統 | `text/template` + `//go:embed` | 模板獨立管理、可讀性高 |
| TS 檔案修改 | String append（非 AST） | registry.ts/proto.ts 結構固定，append 夠用 |
| protoc 執行 | `exec.Command` 直呼叫 | 不依賴 make；跨平台可移植 |
| pbjs 位置 | 先找 `frontend/node_modules/.bin/pbjs` | 不需全域安裝 |
| Go struct 位置 | 根目錄 `<name>.go`（package main） | 與 example 一致；main.go 直接使用 |
| svelgo.json | 標記 app root + 記錄 npm path | `svelgo add` 用來判斷執行位置 |

---

## 開發者體驗對比

### 之前（手動 8 步驟）

1. `go mod init` + 設定 go.mod
2. 定義 proto message，跑 protoc + pbjs
3. 寫 embed.go
4. 寫 main.go（HTTP handler + component struct）
5. 建 frontend/，寫 package.json + vite.config.ts
6. 寫 main.ts / proto.ts / registry.ts
7. 建 static/.gitkeep
8. `npm install`

### 之後（CLI）

```bash
svelgo new myapp     # 全自動
cd myapp
make dev             # 直接跑
```

### 新增 Component：之前（6 個手動步驟）

1. proto 加 message → 跑 make proto
2. Go struct
3. page.Add() 進 handler
4. 建 Svelte component
5. 更新 registry.ts
6. 更新 proto.ts

### 新增 Component：之後

```bash
svelgo add Counter
# 自動完成步驟 1（proto append）、2（Go stub）、4（Svelte stub）、5、6
# 開發者只需填寫：
#   1. proto/app.proto 的 CounterState 欄位
#   2. counter.go 的 HandleEvent 邏輯
#   3. Counter.svelte 的 UI
#   4. main.go 的 page.Add(&Counter{...})
```

---

## 實作順序

```
Step 1: cmd/svelgo/internal/gomod.go     ← ReadModuleName()
Step 2: cmd/svelgo/templates/            ← 所有 .tmpl 檔案 + templates.go
Step 3: cmd/svelgo/commands/shared.go    ← renderTemplate, runCmd, appendIfAbsent
Step 4: cmd/svelgo/commands/new.go       ← RunNew()
Step 5: cmd/svelgo/commands/add.go       ← RunAdd()
Step 6: cmd/svelgo/main.go               ← CLI 入口點
```

CLI 不 import framework runtime（無循環依賴）。

---

## Verification 步驟

1. `go build ./cmd/svelgo/` 在 framework root 成功編譯
2. `svelgo new testapp --local .` 產生所有預期檔案
3. `cd testapp && make dev` → http://localhost:8080 顯示 Hello component
4. `svelgo add Counter` → 5 個改動 + protoc/pbjs 自動執行成功
5. main.go 加 `page.Add(&Counter{id:"c1"})` 後 `make dev` 顯示 Counter component
