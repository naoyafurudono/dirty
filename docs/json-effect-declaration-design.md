# JSONベースのエフェクト宣言インターフェース設計

## 概要

任意の関数に対してエフェクトをJSONで宣言できる汎用的なインターフェースを実装します。sqlc-use統合を置き換え、より柔軟なエフェクト宣言を可能にします。

## JSON形式の仕様

```json
{
  "version": "1.0",
  "effects": {
    "GetUser": "{ select[users] }",
    "CreateUser": "{ insert[users] }",
    "ComplexOperation": "{ select[users] | update[users] | insert[logs] }",
    "SendEmail": "{ network[smtp] | io[filesystem] }",
    "ValidateInput": "{ }",
    "ReadConfig": "{ io[filesystem] | env[read] }"
  }
}
```

**特徴:**
- 関数名をキー、エフェクト式を値とするシンプルな構造
- エフェクト式は新しい文法（`{ ... }`）をそのまま使用
- 空集合 `{ }` も表現可能
- バージョン情報で将来の拡張に対応

## 実装設計

### 1. データ構造

```go
// EffectDeclarations represents JSON-based effect declarations
type EffectDeclarations struct {
    Version string            `json:"version"`
    Effects map[string]string `json:"effects"`
}

// LoadEffectDeclarations loads effect declarations from JSON
func LoadEffectDeclarations(path string) (*EffectDeclarations, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var decls EffectDeclarations
    if err := json.Unmarshal(data, &decls); err != nil {
        return nil, err
    }

    // Validate version
    if decls.Version != "1.0" {
        return nil, fmt.Errorf("unsupported version: %s", decls.Version)
    }

    return &decls, nil
}

// ParsedEffects holds parsed effect expressions
type ParsedEffects map[string]EffectExpr

// ParseAll parses all effect declarations
func (d *EffectDeclarations) ParseAll() (ParsedEffects, error) {
    result := make(ParsedEffects)
    for funcName, effectStr := range d.Effects {
        expr, err := ParseEffectDecl("// dirty: " + effectStr)
        if err != nil {
            return nil, fmt.Errorf("error parsing effects for %s: %w", funcName, err)
        }
        result[funcName] = expr
    }
    return result, nil
}
```

### 2. アナライザーへの統合

```go
type EffectAnalysis struct {
    // ... existing fields ...

    // JSON-based effect declarations
    JSONEffects ParsedEffects
}

// During analysis initialization
func run(pass *analysis.Pass) (interface{}, error) {
    // Load JSON effects
    jsonPath := os.Getenv("DIRTY_EFFECTS_JSON")
    if jsonPath == "" {
        // Try to find in package directory
        if len(pass.Files) > 0 {
            pkgDir := filepath.Dir(pass.Fset.Position(pass.Files[0].Pos()).Filename)
            jsonPath = filepath.Join(pkgDir, "dirty-effects.json")
        }
    }

    var jsonEffects ParsedEffects
    if jsonPath != "" && fileExists(jsonPath) {
        decls, err := LoadEffectDeclarations(jsonPath)
        if err == nil {
            jsonEffects, _ = decls.ParseAll()
        }
    }

    // ... rest of analysis ...
}
```

### 3. 優先順位

エフェクトの解決順序：
1. ソースコード内の `// dirty:` コメント（最優先）
2. JSONファイルでの宣言

### 4. ファイルの配置

```
project/
├── pkg/
│   ├── db/
│   │   ├── dirty-effects.json # パッケージ固有
│   │   └── queries.go
│   └── api/
│       └── handlers.go
```

検索順序：
1. 環境変数 `DIRTY_EFFECTS_JSON`で指定されたパス
2. 解析対象パッケージディレクトリの `dirty-effects.json`

## 使用例

### 基本的な使用

```bash
# エフェクト宣言の作成
cat > dirty-effects.json << EOF
{
  "version": "1.0",
  "effects": {
    "ProcessPayment": "{ select[users] | update[balance] | insert[transactions] | network[payment_api] }",
    "SendNotification": "{ select[users] | network[email] | insert[notifications] }",
    "ValidateInput": "{ }",
    "ExternalAPICall": "{ network[external_api] | select[cache] | insert[logs] }"
  }
}
EOF

# 実行
dirty ./...

# または環境変数で指定
DIRTY_EFFECTS_JSON=./my-effects.json dirty ./...
```

### コード内宣言との組み合わせ

```go
// コード内の宣言が優先される
// dirty: { select[users] | custom[validation] }
func ValidateUser(id int64) error { ... }

// JSONで宣言された関数
func ProcessPayment(amount float64) error {
    // JSONから: { select[users] | update[balance] | insert[transactions] | network[payment_api] }
}
```

## 実装ステップ

1. **既存のSQLC統合を削除**
   - `sqlc_analyzer.go`の削除
   - 関連するテストとドキュメントの削除

2. **新しいJSON形式の実装**
   - `effect_declarations.go`の作成
   - パーサーとローダーの実装

3. **アナライザーの更新**
   - JSONエフェクトの読み込み
   - 優先順位の実装

4. **テストとドキュメント**
   - 新形式のテストケース
   - READMEの更新

## まとめ

この設計により：
1. **シンプル**: 関数名とエフェクト式の単純なマッピング
2. **一貫性**: 新しい文法をそのまま使用
3. **汎用性**: あらゆる種類のエフェクトを表現可能
4. **明確な優先順位**: コード内宣言が常に優先
