# dirty

dirtyはGo言語向けのエフェクトシステムもどきです。vetツールとして用います。

## 表記

関数宣言では、それが起こすエフェクトを表明できます。

```go
// dirty: select[user] select[organization] insert[member]
func f() {}
```

上記のように `// dirty: ` から始まるスペース区切りのエフェクトラベルの列が、その関数が起こすエフェクトです。
dirtyではエフェクトラベルの集合として解釈されます。つまり、重複や順序は無視されます。

## 検査

dirtyはモジュール内の関数宣言を走査しエフェクトの表明が一貫していることを検査します。

以下のように、関数okの本体でfを呼び出す場合、okはfが起こすエフェクトを起こすと解釈します。
そのため、okのエフェクトはfのエフェクトのスーパーセットである必要があります。

```go
// dirty: select[user] select[organization] insert[member] insert[user]
func ok() {
	...
	f()
	...
}
```

したがって、以下のようにfのエフェクトを包含しないエフェクトしか表明しない場合は、エフェクトの検査が失敗します。

```go
// dirty: select[user]
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

// dirty: select[user] select[organization] insert[member] insert[user]
func ok() {
	...
	implicit()
	...
}

// dirty: select[user]
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
    // dirty: insert[log], select[user]
```

## sqlc-use との統合

dirtyは[sqlc-use](https://github.com/naoyafurudono/sqlc-use)の出力を読み込んで、SQLCで生成された関数のエフェクトを自動的に検出できます。

### JSONファイルの配置

dirtyは以下の順序でJSONファイルを検索します：

1. **環境変数 `DIRTY_SQLC_JSON`** で指定されたパス（最優先）
2. **カレントディレクトリ**の `query-table-operations.json`
3. **解析対象パッケージディレクトリ**の `query-table-operations.json`

```bash
# 方法1: 環境変数で明示的に指定
DIRTY_SQLC_JSON=/path/to/query-table-operations.json dirty ./...

# 方法2: パッケージディレクトリに配置（推奨）
myproject/
└── internal/
    └── db/
        ├── queries.go                    # SQLCで生成されたコード
        └── query-table-operations.json   # ここに配置
```

### JSONフォーマット仕様

sqlc-useが生成するJSONファイルは以下の形式である必要があります：

```json
{
  "関数名": [
    {
      "operation": "操作種別",
      "table": "テーブル名"
    }
  ]
}
```

**フィールド説明：**

- **関数名**: SQLCで生成された関数名（完全一致）
- **operation**: `select`, `insert`, `update`, `delete` のいずれか
- **table**: 操作対象のテーブル名

**具体例：**

```json
{
  "GetUser": [{ "operation": "select", "table": "users" }],
  "GetUserWithPosts": [
    { "operation": "select", "table": "users" },
    { "operation": "select", "table": "posts" }
  ],
  "CreateUserWithAudit": [
    { "operation": "insert", "table": "users" },
    { "operation": "insert", "table": "audit_logs" }
  ],
  "UpdateUserStatus": [{ "operation": "update", "table": "users" }],
  "DeleteSession": [{ "operation": "delete", "table": "sessions" }]
}
```

### エフェクトマッピングと動作

#### 1. エフェクトの変換ルール

JSONの操作はdirtyのエフェクトラベルに以下のように変換されます：

| operation | table | dirtyエフェクト |
| --------- | ----- | --------------- |
| select    | users | select[users]   |
| insert    | logs  | insert[logs]    |
| update    | posts | update[posts]   |
| delete    | auth  | delete[auth]    |

#### 2. 関数の認識

dirtyは以下の条件で関数を認識します：

- メソッド名が完全一致する（例：`q.GetUser()` → `GetUser`）
- 通常の関数呼び出しも同様（例：`GetUser()` → `GetUser`）

#### 3. エフェクトの適用

SQLCから読み込まれたエフェクトは、その関数に`// dirty:`宣言があるかのように扱われます：

```go
// JSONに "GetUser": [{"operation": "select", "table": "users"}] がある場合

// 以下の2つは同等に扱われる：
func (q *Queries) GetUser(ctx context.Context, id int64) (User, error)

// dirty: select[users]
func (q *Queries) GetUser(ctx context.Context, id int64) (User, error)
```

### 使用例

1. sqlc-useでクエリを解析:

```bash
sqlc-use analyze > query-table-operations.json
```

2. dirtyでの検証:

```go
// dirty: select[users], insert[logs]
func ProcessUser(ctx context.Context, q *Queries, id int64) error {
    user, err := q.GetUser(ctx, id)  // ✓ select[users] は宣言済み
    if err != nil {
        return err
    }
    return logAccess(user.ID)  // ✓ insert[logs] は宣言済み
}

// dirty: insert[logs]
func BrokenFunction(ctx context.Context, q *Queries, id int64) error {
    user, err := q.GetUser(ctx, id)  // ✗ エラー: select[users] が未宣言
    return err
}
```

### トラブルシューティング

#### JSONファイルが読み込まれない場合

1. ファイル名が正確に `query-table-operations.json` であることを確認
2. ファイルパスが正しいことを確認（相対パス/絶対パス）
3. JSONの形式が正しいことを確認（不正なJSONはスキップされます）

#### エフェクトが検出されない場合

1. 関数名がJSONのキーと完全一致していることを確認
2. SQLCで生成された関数の呼び出し方法を確認（メソッド呼び出しの場合、レシーバーは無視されます）

### 制限事項

- **ローカル定義優先**: パッケージ内に同名の関数が定義されている場合、その関数の`// dirty:`宣言が優先されます
- **エラー無視**: JSONファイルの読み込みエラーは警告なしに無視されます（dirtyの実行は継続）
- **大文字小文字の区別**: 関数名は大文字小文字を区別します

### CI/CDでの使用例

```yaml
# GitHub Actions
- name: Generate SQLC effects
  run: sqlc-use analyze > query-table-operations.json

- name: Run dirty with SQLC integration
  run: DIRTY_SQLC_JSON=query-table-operations.json dirty ./...

# または、生成したJSONをパッケージに配置
- name: Place SQLC effects
  run: |
    sqlc-use analyze > internal/db/query-table-operations.json
    dirty ./...
```

## 制限

実装をするのが面倒なので、今は色々な実装上のサボりをします。結果的に予期せぬ振る舞いがたくさん生じます。

- モジュール外のエフェクト表明は参照しません
- エフェクトの走査を真面目にやりません
  - 高階関数とかをサポートしません。本来なら型システムがやるようなことをするべきです。
  - 現時点では「関数宣言の中に出現した関数呼び出しのcalleeのエフェクトの和集合」をその関数宣言のエフェクトとします。
  - 無名関数の本体に囲まれていようが関係ないですし、引数として渡された関数を呼び出した場合はそのエフェクトを無視することになります。
