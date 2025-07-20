# sqlc-use統合

## 概要

[sqlc-use](https://github.com/naoyafurudono/sqlc-use)のJSON出力をdirtyのエフェクト表明として活用し、SQLクエリのデータベース操作を自動的に検証します。

## 基本フロー

```
1. sqlc generate         → Goコード生成
2. sqlc-use (plugin)     → query-table-operations.json
3. dirty --sqlc-json=... → エフェクト検証
```

## 変換ルール

```json
// sqlc-use出力
{
  "GetUser": [
    {"operation": "select", "table": "users"}
  ],
  "CreateUserWithLog": [
    {"operation": "insert", "table": "users"},
    {"operation": "insert", "table": "audit_logs"}
  ]
}
```

→ dirtyエフェクト: `select[users]`, `insert[users]`, `insert[audit_logs]`

## 実装設計

### 1. JSONパーサー (`analyzer/sqlc_analyzer.go`)

```go
type SQLCOperation struct {
    Operation string `json:"operation"`
    Table     string `json:"table"`
}

func LoadSQLCEffects(jsonPath string) (map[string][]SQLCOperation, error)
func ConvertToEffects(ops []SQLCOperation) []string
```

### 2. 関数マッチング

sqlc生成パターンに対応：
- `func (q *Queries) GetUser(...)`
- `func GetUser(ctx context.Context, db DBTX, ...)`

### 3. CLIサポート

```bash
# 基本使用
dirty --sqlc-json=query-table-operations.json ./...

# 環境変数
DIRTY_SQLC_JSON=query-table-operations.json dirty ./...

# 自動検出（カレントディレクトリ）
dirty ./...  # query-table-operations.jsonを自動検索
```

## 使用例

### Makefile
```makefile
check-effects:
	sqlc generate
	dirty --sqlc-json=query-table-operations.json ./...
```

### GitHub Actions
```yaml
- name: Check database effects
  run: |
    sqlc generate
    dirty --sqlc-json=query-table-operations.json ./...
```

### エラー出力（詳細モード）
```
function calls GetUser which has effects [select[users]] not declared

  Effect source: sqlc-use (query-table-operations.json)
  Database operation: SELECT on table 'users'
```

## 実装ステップ

1. **MVP**: JSONパーサーと基本的な変換
2. **関数マッチング**: sqlc生成関数の検出
3. **統合**: 既存アナライザーへの組み込み
4. **CLI**: フラグと自動検出の実装