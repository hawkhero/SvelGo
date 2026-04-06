# r/golang UI Framework 需求研究

> 研究日期：2026-03-29
> 目的：評估 Go-first UI framework（如 SvelGo）的市場需求

---

## 目前社群使用的解決方案

主流 Go + 前端組合是「**GoTH stack**」：Go + Templ + HTMX + Tailwind（有時加 DaisyUI）。

| 工具 | 用途 | 狀態 |
|---|---|---|
| **templ** (a-h/templ) | JSX-like HTML 模板，compile-time type-safe | 10.2k stars，7k+ 專案使用，**主流** |
| **HTMX** | Hypermedia attributes（AJAX/WebSocket/SSE） | 普及；非 Go 專屬 |
| **Datastar** | HTMX 替代品，SSE + reactive signals | 成長中；Go 是其主要示例語言 |
| **DatastarUI** | shadcn/ui 風格的 Go/templ/Datastar component library | 125 stars，非常新 |
| **htmgo** | Go+HTMX 的 framework 層，含 component 概念 | 小型，早期 |
| **jfyne/live** | Phoenix LiveView 的 Go 實作 | 734 stars，低活躍度 |
| **golive (brendonmatos)** | 另一個 Go LiveView | 265 stars，**2022 年 archived** |
| **golive (canopyclimate)** | 較新的 Go LiveView | 61 stars，alpha，最後 commit 2024-01 |
| **Bud** | Rails 風格的 Go full-stack framework（SSR + Svelte/React） | 2022 Show HN 高關注，之後沉寂 |
| **Vecty / Vugu** | Go 編譯成 WebAssembly 寫 UI | 實驗性，多數已停止維護 |

---

## 社群痛點（Go + Frontend）

### 1. 開發流程複雜
- 沒有等同 `npm run dev` 的一鍵啟動。要同時跑 Vite + Air + templ watcher，依賴脆弱的 Makefile。
- `templ generate` 的 code generation 步驟增加摩擦。

### 2. 沒有成熟的 UI component library
- Go 沒有等同 shadcn/ui、Radix、Flowbite 的東西可以直接搭配 server-rendered Go 使用。
- 開發者普遍反映「所有 component 都要自己寫」。
- htmgo 的 HN 討論明確要求「pre-built components」。

### 3. Client-side 狀態管理有缺口
- HTMX 沒有內建 client-side state，通常要加 Alpine.js。
- HTMX + Alpine.js 的同步問題是 Datastar 切換的主因。
- 真實即時多人 UI，Go-native 沒有清楚的解決方案。

### 4. Go LiveView 替代品停滯
- 三個主要 Go LiveView 專案：archived、alpha（61 stars）、低活躍（734 stars）。
- 沒有任何一個達到 Phoenix LiveView 的生產就緒程度。

### 5. 還是需要一些 JS
- GoTH 六個月回顧文：「React 有時仍然必要」（例如用了 React Flow 畫圖）。
- 複雜 widget 沒有 Go-native 等價物。

### 6. 兩個 codebase 的維護成本
- Dagger 將 React 前端換成 Go WASM，主因就是「兩個獨立 codebase 的維護稅」。

---

## 是否有對 SvelGo 這類框架的需求？

**有，且是明確表達的需求。**

關鍵訊號：

- htmgo 的 HN 討論：社群明確要求 full framework parity、pre-built components、更快的 live reload。
- templ HN 討論：「client-side JS 會讓大型專案變得混亂」—— 開發者想要 reactive state 但不離開 Go。
- Bud 的 Show HN：「single Go binary 包含整個 web app」獲得強烈正面反應；Go concurrency 用於 UI 的概念受到認可。
- 多篇部落格把自己的 stack 描述為「duct tape」—— 表示在繞過缺口。
- Datastar 生態系的快速成長，顯示市場在朝這個方向走。

**SvelGo 特別有需求的差異化點：**
- 單一 binary 部署（社群頻繁稱讚這點）
- 生產環境不需要 Node process（明確被要求）
- Go-native component model + pre-built components（到處都在指出這個缺口）
- WebSocket-based 即時性不需要寫 JS
- Go struct → 互動 UI（兩個 codebase 的維護成本是真實痛點）

---

## 最接近 SvelGo 的競品

| 專案 | 相似處 | 主要差異 | 狀態 |
|---|---|---|---|
| **Datastar + templ + DatastarUI** | server-driven reactive UI、Go templating、component library | 用 SSE 而非 WebSocket；HTML attributes 而非 Go structs；三個獨立部分 | 活躍，成長中 |
| **jfyne/live** | WebSocket-based server-driven UI in Go | 用 html/template 不用 Svelte；無 pre-built components | 低活躍 |
| **canopyclimate/golive** | LiveView pattern + WebSocket diffs | 不同 wire format；無 pre-built components | Alpha，61 stars |
| **htmgo** | Go framework + component abstraction | component 用 Go function/HTML 定義，非 protobuf；無 WebSocket reactivity | 非常小，早期 |
| **Bud** | Go + embedded Svelte + single binary | 用 V8 runtime（47MB binary）；Svelte 是 rendering layer 非 embedded bundle；已沉寂 | 2022 後沉寂 |

SvelGo 的獨特組合——Go struct 定義 component、protobuf wire protocol、binary WebSocket frames、Svelte 預編譯並 embed、單一 Go binary——目前沒有完全對應的專案。

哲學上最接近的祖先是 **Bud**（Go + embedded Svelte + single binary），但 Bud 用 V8 runtime，從未像 SvelGo 這樣 embed 預編譯的 bundle。

---

## 風險評估

1. **Go 社群的反框架文化**：r/golang 歷來對 framework 有抵觸。定位為「library」可能比「framework」更受歡迎。
2. **需要 pre-built component library 才能驅動採用**：目前 SvelGo 沒有 component set，而缺乏 component 是社群最常提到的缺口。
3. **競爭對手是整個生態系組合**：Datastar + templ + DatastarUI 三件套在快速演進，是最直接的競爭對手。
4. **文件與 DX**：templ 成功的重要因素是優秀的文件與 VS Code 插件；SvelGo 需要同等投資。
