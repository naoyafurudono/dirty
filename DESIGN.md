# Dirty - 設計と実装

## 概要

dirtyは、Go言語向けのエフェクトシステムを実現する静的解析ツールです。

## アーキテクチャ

```
AST解析 → エフェクト収集 → 呼び出しグラフ構築 → エフェクト伝播 → 整合性チェック
```

## 現在の実装 ✅

- `//dirty:` コメントからエフェクトを抽出
- 呼び出しグラフの構築と暗黙的エフェクト計算
- ワークリストアルゴリズムによる効率的な伝播
- メソッド呼び出しのサポート
- 循環参照の正しい処理

### 実装済みの暗黙的エフェクト計算

```go
// 宣言なし関数も解析対象
func intermediate() {
    GetUser(1)  // select[user]を暗黙的に持つ
}

//dirty: select[member]
func caller() {
    intermediate()  // エラー: select[user]が不足
}
```

### 実装方針

1. **データ構造**
   ```go
   type FunctionInfo struct {
       DeclaredEffects []string  // //dirty: の内容
       ComputedEffects []string  // 計算されたエフェクト
       HasDeclaration  bool      // 宣言の有無
   }
   ```

2. **エフェクト伝播アルゴリズム**
   - ワークリスト方式で効率的に伝播
   - 循環参照は不動点まで反復

## sqlc-use統合 ✅

dirtyは[sqlc-use](https://github.com/naoyafurudono/sqlc-use)と統合され、SQLCで生成された関数のデータベースエフェクトを自動的に検出できます。

### 実装方法

1. **JSON形式**: sqlc-useの出力形式をそのまま利用
   ```json
   {
     "GetUser": [
       {"operation": "select", "table": "users"}
     ]
   }
   ```

2. **エフェクト変換**: `operation[table]` 形式に変換
   - `select[users]`, `insert[logs]`, `update[posts]`, `delete[sessions]`

3. **自動検出**: 
   - 環境変数 `DIRTY_SQLC_JSON` で指定
   - パッケージディレクトリの `query-table-operations.json` を自動検出

### 統合の流れ

```
sqlc-use analyze → JSON出力 → dirty読み込み → エフェクト自動適用
```

## 今後の改善

- パフォーマンスの最適化
- クロスパッケージ解析
- IDE統合（gopls拡張）

## 制限事項

- モジュール外の関数は未対応
- 高階関数は未対応
- 動的呼び出しは追跡不可