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

## 今後の改善

- sqlc-use統合（JSONからエフェクト自動生成）
- パフォーマンスの最適化
- クロスパッケージ解析
- IDE統合（gopls拡張）

## 制限事項

- モジュール外の関数は未対応
- 高階関数は未対応
- 動的呼び出しは追跡不可