# Dirty Effect Analyzer - Design Document

## 概要

dirtyは、Go言語向けのエフェクトシステムを実現する静的解析ツールです。関数が引き起こす副作用（エフェクト）をコメントベースのアノテーションで表明し、それらの整合性を検証します。

## アーキテクチャ

### 全体構成

```
┌─────────────────┐
│   AST Parser    │ → Go ASTから関数とコメントを抽出
└────────┬────────┘
         │
┌────────▼────────┐
│ Effect Collector│ → エフェクト宣言を収集
└────────┬────────┘
         │
┌────────▼────────┐
│ Call Graph      │ → 関数呼び出し関係を構築
│   Builder       │
└────────┬────────┘
         │
┌────────▼────────┐
│ Effect          │ → 暗黙的エフェクトを計算
│ Propagator      │
└────────┬────────┘
         │
┌────────▼────────┐
│ Effect Checker  │ → エフェクトの整合性を検証
└────────┬────────┘
         │
┌────────▼────────┐
│ Error Reporter  │ → 診断メッセージを生成
└─────────────────┘
```

## 主要コンポーネント

### 1. Effect Collector
- **役割**: 関数の`//dirty:`コメントからエフェクト宣言を収集
- **入力**: AST内の関数宣言ノード
- **出力**: 関数名からエフェクトリストへのマッピング

### 2. Call Graph Builder
- **役割**: 関数間の呼び出し関係を解析
- **考慮事項**:
  - 直接的な関数呼び出し
  - メソッド呼び出し
  - （将来）関数型変数を通じた呼び出し

### 3. Effect Propagator
- **役割**: 暗黙的エフェクトの計算
- **アルゴリズム**: 不動点反復法
  1. 各関数の初期エフェクトを設定（宣言されたもののみ）
  2. 各関数について、呼び出し先のエフェクトを収集
  3. エフェクトセットが変化しなくなるまで繰り返す

### 4. Effect Checker
- **役割**: エフェクト宣言の整合性を検証
- **検証ルール**:
  - 宣言された関数: 呼び出し先のエフェクトが宣言に含まれているか
  - 宣言されていない関数: 検証をスキップ（暗黙的エフェクトは計算するが検証はしない）

## データ構造

```go
// FunctionInfo は関数の情報を保持
type FunctionInfo struct {
    Name            string
    DeclaredEffects []string  // //dirty: で宣言されたエフェクト
    ComputedEffects []string  // 計算された実際のエフェクト
    HasDeclaration  bool      // //dirty: コメントの有無
    Position        token.Pos
}

// CallGraph は関数呼び出し関係を表現
type CallGraph struct {
    // Calls[A] = [B, C] は関数Aが関数B,Cを呼び出すことを示す
    Calls   map[string][]CallSite
    // CalledBy[B] = [A] は関数Bが関数Aから呼ばれることを示す
    CalledBy map[string][]string
}

type CallSite struct {
    Callee   string
    Position token.Pos
}
```

## アルゴリズムの詳細

### エフェクト伝播アルゴリズム

```
function PropagateEffects(functions, callGraph):
    changed = true
    while changed:
        changed = false
        for each function f in functions:
            oldEffects = f.ComputedEffects
            newEffects = f.DeclaredEffects
            
            for each callee in callGraph.Calls[f.Name]:
                newEffects = union(newEffects, functions[callee].ComputedEffects)
            
            if newEffects != oldEffects:
                f.ComputedEffects = newEffects
                changed = true
    
    return functions
```

### 循環参照の処理

相互再帰や自己再帰的な関数呼び出しは、不動点反復によって自然に処理されます。

## 実装フェーズ

### フェーズ1: 基本実装（現在）
- [x] エフェクト宣言のパース
- [x] 直接的な関数呼び出しのチェック
- [x] 基本的なエラー報告

### フェーズ2: 完全な仕様実装
- [ ] 暗黙的エフェクトの計算
- [ ] 呼び出しグラフの構築
- [ ] エフェクト伝播アルゴリズム
- [ ] メソッド呼び出しのサポート

### フェーズ3: 拡張機能
- [ ] クロスパッケージ解析
- [ ] エフェクトの詳細な分類（read/write等）
- [ ] カスタムエフェクトルール
- [ ] IDE統合のための追加API

## 制限事項

### 現在の制限
1. **モジュール境界**: 他モジュールの関数のエフェクトは考慮しない
2. **高階関数**: 関数を引数として受け取る場合のエフェクトは追跡しない
3. **動的呼び出し**: リフレクションやインターフェース経由の呼び出しは追跡しない

### 設計上の決定
- **シンプルさ優先**: 完全性よりも実用性を重視
- **漸進的な採用**: 既存コードに段階的に適用可能
- **拡張可能性**: 将来的により高度な解析を追加可能

## パフォーマンス考慮事項

1. **インクリメンタル解析**: 変更された関数のみ再解析
2. **キャッシング**: 計算済みエフェクトのキャッシュ
3. **並列処理**: 独立した関数群の並列解析

## エラー報告の改善

### 現在
```
function calls GetUser which has effects [select[user]] not declared in this function
```

### 提案
```
function calls GetUser which has effects [select[user]] not declared in this function
  GetUser declares: select[user]
  This function declares: select[member]
  Missing effects: select[user]
  
  Effect propagation path:
    UpdateUserProfile -> GetUserWithAudit -> GetUser
                                              ^^^^^^^^
                                              introduces: select[user]
```

## まとめ

この設計により、dirtyは段階的に実装可能でありながら、将来の拡張にも対応できる柔軟なアーキテクチャを持つことができます。最も重要なのは、実用的なツールとして既存のGoプロジェクトに容易に導入できることです。