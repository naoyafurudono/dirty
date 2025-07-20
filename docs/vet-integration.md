# go vetへの統合方法

## 方法1: go vetのカスタムツールとして実行

最も簡単な方法は、`go vet`の`-vettool`フラグを使用することです。

```bash
# dirtyツールをビルド
go build -o dirty ./cmd/dirty

# go vetのカスタムツールとして実行
go vet -vettool=$(pwd)/dirty ./...
```

### シェルエイリアスの設定

```bash
# ~/.bashrc or ~/.zshrc に追加
alias govet-dirty='go vet -vettool=$(go env GOPATH)/bin/dirty'

# 使用方法
govet-dirty ./...
```

## 方法2: マルチチェッカーツールの作成

複数のアナライザーを組み合わせた独自のvetツールを作成：

```go
// cmd/myvet/main.go
package main

import (
    "github.com/naoyafurudono/dirty/analyzer"
    "golang.org/x/tools/go/analysis/unitchecker"

    // 標準のvetチェッカー
    "golang.org/x/tools/go/analysis/passes/atomic"
    "golang.org/x/tools/go/analysis/passes/bools"
    "golang.org/x/tools/go/analysis/passes/copylocks"
    // ... 他のチェッカー
)

func main() {
    unitchecker.Main(
        // 標準のチェッカー
        atomic.Analyzer,
        bools.Analyzer,
        copylocks.Analyzer,

        // dirtyアナライザーを追加
        analyzer.Analyzer,
    )
}
```

ビルドして使用：
```bash
go build -o myvet ./cmd/myvet
go vet -vettool=$(pwd)/myvet ./...
```

## 方法3: go/analysisのmultichecker

開発時に便利な方法：

```go
// cmd/multichecker/main.go
package main

import (
    "github.com/naoyafurudono/dirty/analyzer"
    "golang.org/x/tools/go/analysis/multichecker"
    "golang.org/x/tools/go/analysis/passes/printf"
    "golang.org/x/tools/go/analysis/passes/shadow"
)

func main() {
    multichecker.Main(
        analyzer.Analyzer,
        printf.Analyzer,
        shadow.Analyzer,
        // 必要な他のアナライザー
    )
}
```

## 方法4: golangci-lintへの統合

golangci-lintのプラグインとして追加（要プルリクエスト）：

1. golangci-lintをフォーク
2. `pkg/golinters/dirty.go`を作成
3. `.golangci.yml`で有効化

```yaml
linters:
  enable:
    - dirty
```

## 推奨される統合方法

### プロジェクト固有の場合

Makefileに追加：
```makefile
# Makefile
.PHONY: vet
vet:
	go vet ./...
	go vet -vettool=$$(go env GOPATH)/bin/dirty ./...

# または
lint:
	go vet ./...
	dirty ./...
```

### CI/CDでの使用

```yaml
# .github/workflows/test.yml
- name: Install dirty
  run: go install github.com/naoyafurudono/dirty/cmd/dirty@latest

- name: Run go vet
  run: go vet ./...

- name: Run dirty analyzer
  run: go vet -vettool=$(go env GOPATH)/bin/dirty ./...
```

## goコマンドへの統合（将来的な目標）

Go本体への統合を目指す場合：

1. golang/go へのプロポーザル提出
2. x/tools への実験的実装
3. 十分な実績後、cmd/vet への統合

現時点では、**方法1のカスタムツール**が最も実用的です。
