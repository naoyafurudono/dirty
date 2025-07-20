# dirty

dirtyはGo言語向けのエフェクトシステムもどきです。vetツールとして用います。

## 表記

関数宣言では、それが起こすエフェクトを表明できます。

```go
//dirty: select[user] select[organization] insert[member]
func f() {}
```

上記のように `//dirty: ` から始まるスペース区切りのエフェクトラベルの列が、その関数が起こすエフェクトです。
dirtyではエフェクトラベルの集合として解釈されます。つまり、重複や順序は無視されます。

## 検査

dirtyはモジュール内の関数宣言を走査しエフェクトの表明が一貫していることを検査します。

以下のように、関数okの本体でfを呼び出す場合、okはfが起こすエフェクトを起こすと解釈します。
そのため、okのエフェクトはfのエフェクトのスーパーセットである必要があります。

```go
//dirty: select[user] select[organization] insert[member] insert[user]
func ok() {
	...
	f()
	...
}
```

したがって、以下のようにfのエフェクトを包含しないエフェクトしか表明しない場合は、エフェクトの検査が失敗します。

```go
//dirty: select[user]
func ng() {
	...
	f()
	...
}
```

表明のない関数は、検査の対象にはなりませんが、暗黙的にその関数が起こすエフェクトを計算されます。

```go
func implicit() {
	...
	f()
	...
}

//dirty: select[user] select[organization] insert[member] insert[user]
func ok() {
	...
	implicit()
	...
}

//dirty: select[user]
func ng() {
	...
	implicit()
	...
}
```

この例ではimplicitにエフェクトの表明はありません。そのためimplicitの表明に対する検証は行われません。ただしimplicitはfを呼び出すので、fのエフェクトを生じると扱われます。
ok, ngではimplicitはfを呼び出すので、結果的にそれらはfのエフェクトを生じると扱われ、それぞれの表明に対する検証に反映されます。

## インストール

```bash
go install github.com/naoyafurudono/dirty/cmd/dirty@latest
```

## 使い方

### 基本的な使い方

```bash
# カレントパッケージをチェック
dirty .

# 特定のパッケージをチェック
dirty ./pkg/...

# すべてのパッケージをチェック
dirty ./...
```

### go vetツールとして使用

```bash
# vet-dirtyをインストール
go install github.com/naoyafurudono/dirty/cmd/vet-dirty@latest

# go vetのカスタムツールとして実行
go vet -vettool=$(go env GOPATH)/bin/vet-dirty ./...
```

### Makefileでの使用例

```makefile
.PHONY: lint
lint:
	go vet ./...
	dirty ./...

# または
.PHONY: vet
vet:
	go vet -vettool=$$(go env GOPATH)/bin/vet-dirty ./...
```

### CI/CDでの使用例

```yaml
# GitHub Actions
- name: Install dirty
  run: go install github.com/naoyafurudono/dirty/cmd/dirty@latest

- name: Run dirty analyzer
  run: dirty ./...
```

### エラー出力の例

通常モード:
```bash
$ dirty ./...
example/simple.go:29:12: function calls GetUserByID which has effects [select[user]] not declared in this function
```

詳細モード（環境変数 `DIRTY_VERBOSE=1` を設定）:
```bash
$ DIRTY_VERBOSE=1 dirty ./...
example/simple.go:43:12: function calls HelperFunction which has effects [select[user]] not declared in this function

  Called function 'HelperFunction' requires:
    - select[user]

  Function 'UseHelper' declares:
    - insert[log]

  Missing effects:
    - select[user]

  Effect propagation path:
    HelperFunction
       effects: [select[user]]
      └─ GetUserByID (from HelperFunction)
         effects: [select[user]]

  To fix, add the missing effects to the function declaration:
    //dirty: insert[log], select[user]
```

## sqlc-use との統合

dirtyは[sqlc-use](https://github.com/naoyafurudono/sqlc-use)の出力を読み込んで、SQLCで生成された関数のエフェクトを自動的に検出できます。

### 使い方

1. sqlc-useでクエリのテーブル操作を解析:
```bash
sqlc-use analyze > query-table-operations.json
```

2. dirtyでSQLCエフェクトを使用:
```bash
# 環境変数で指定
DIRTY_SQLC_JSON=query-table-operations.json dirty ./...

# またはパッケージディレクトリに配置（自動検出）
cp query-table-operations.json ./pkg/db/
dirty ./pkg/db/...
```

### 例

sqlc-useが生成するJSON:
```json
{
  "GetUser": [
    {"operation": "select", "table": "users"}
  ],
  "CreateUserWithAudit": [
    {"operation": "insert", "table": "users"},
    {"operation": "insert", "table": "audit_logs"}
  ]
}
```

dirtyでの検証:
```go
//dirty: select[users], insert[logs]
func ProcessUser(ctx context.Context, q *Queries, id int64) error {
    user, err := q.GetUser(ctx, id) // 自動的に select[users] を検出
    if err != nil {
        return err
    }
    return logAccess(user.ID) // insert[logs]
}
```

## 制限

実装をするのが面倒なので、今は色々な実装上のサボりをします。結果的に予期せぬ振る舞いがたくさん生じます。

- モジュール外のエフェクト表明は参照しません
- エフェクトの走査を真面目にやりません
  - 高階関数とかをサポートしません。本来なら型システムがやるようなことをするべきです。
  - 現時点では「関数宣言の中に出現した関数呼び出しのcalleeのエフェクトの和集合」をその関数宣言のエフェクトとします。
  - 無名関数の本体に囲まれていようが関係ないですし、引数として渡された関数を呼び出した場合はそのエフェクトを無視することになります。
