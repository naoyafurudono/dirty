# SQLC Integration Example

この例では、dirtyがsqlc-useの出力を使用してSQLCで生成された関数のエフェクトを自動的に検出する方法を示します。

## ファイル構成

- `query-table-operations.json` - sqlc-useが生成したクエリとテーブル操作のマッピング
- `db.go` - SQLCで生成されたデータベースアクセスコード（モック）
- `service.go` - SQLCの関数を使用するビジネスロジック

## 実行方法

### 1. 現在のディレクトリで実行（JSONが自動検出される）

```bash
cd example/sqlc
dirty .
```

### 2. プロジェクトルートから環境変数を使用

```bash
DIRTY_SQLC_JSON=example/sqlc/query-table-operations.json dirty ./example/sqlc
```

## エフェクトの例

### SQLCから自動検出されるエフェクト

`query-table-operations.json`から以下のエフェクトが自動的に検出されます：

- `GetUser` → `select[users]`
- `CreateUser` → `insert[users]`
- `GetUserWithOrganization` → `select[users], select[organizations]`
- `CreateUserWithAuditLog` → `insert[users], insert[audit_logs]`
- `ArchiveUserAndSessions` → `update[users], delete[sessions]`

### サービス層での使用例

```go
//dirty: select[users], insert[audit_logs]
func GetUserWithAuditLog(ctx context.Context, q *Queries, userID int64) (*User, error) {
    // q.GetUser() は自動的に select[users] エフェクトが検出される
    user, err := q.GetUser(ctx, userID)
    // ...
}
```

## エラーの例

`BrokenGetUser`関数は`select[users]`エフェクトを宣言していないため、dirtyがエラーを報告します：

```go
//dirty: insert[audit_logs]  // ❌ select[users]が不足
func BrokenGetUser(ctx context.Context, q *Queries, userID int64) (*User, error) {
    user, err := q.GetUser(ctx, userID)  // エラー: select[users]が未宣言
    // ...
}
```

## 詳細モード

エラーの詳細を確認するには：

```bash
DIRTY_VERBOSE=1 dirty .
```
