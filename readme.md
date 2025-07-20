# dirty

dirtyはGo言語向けのエフェクトシステムもどきです。vetツールとして用います。

## 表記

関数宣言では、それが起こすエフェクトを表明できます。

```go
// dirty: { select[user] | select[organization] | insert[member] }
func f() {}
```

上記のように `// dirty:` から始まり、`{ }` で囲まれた中に `|` で区切られたエフェクトラベルを記述します。
dirtyではエフェクトラベルの集合として解釈されます。つまり、重複や順序は無視されます。

空のエフェクト宣言も可能です：
```go
// dirty: { }
func emptyEffects() {}
```

## 検査

dirtyはモジュール内の関数宣言を走査しエフェクトの表明が一貫していることを検査します。

以下のように、関数okの本体でfを呼び出す場合、okはfが起こすエフェクトを起こすと解釈します。
そのため、okのエフェクトはfのエフェクトのスーパーセットである必要があります。

```go
// dirty: { select[user] | select[organization] | insert[member] | insert[user] }
func ok() {
	...
	f()
	...
}
```

したがって、以下のようにfのエフェクトを包含しないエフェクトしか表明しない場合は、エフェクトの検査が失敗します。

```go
// dirty: { select[user] }
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

// dirty: { select[user] | select[organization] | insert[member] | insert[user] }
func ok() {
	...
	implicit()
	...
}

// dirty: { select[user] }
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

### 実例

詳細な例は[example/](example/)ディレクトリを参照してください：
- [example/simple.go](example/simple.go) - 基本的な使い方
- [example/complex_chain.go](example/complex_chain.go) - 複雑な呼び出しチェーン
- [example/jsoneffects/](example/jsoneffects/) - JSONエフェクト宣言の例

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

## JSONによるエフェクト宣言

dirtyはJSONファイルから関数のエフェクトを宣言できます。これにより、外部ツールで生成された関数や、ソースコードを変更できない関数に対してもエフェクトを宣言できます。

### JSONファイルの形式

```json
{
  "version": "1.0",
  "effects": {
    "GetUser": "{ select[users] }",
    "CreateUser": "{ insert[users] }",
    "ProcessPayment": "{ select[users] | update[balance] | insert[transactions] | network[payment_api] }",
    "ValidateInput": "{ }",
    "ExternalAPICall": "{ network[external_api] | select[cache] | insert[logs] }"
  }
}
```

**形式の説明：**
- `version`: フォーマットのバージョン（現在は"1.0"のみサポート）
- `effects`: 関数名をキー、エフェクト式を値とするマッピング
- エフェクト式はソースコード内の宣言と同じ文法を使用

**JSON Schema**: `schema/dirty-effects.schema.json`でスキーマが定義されています。VS CodeなどのエディタでIntelliSenseとバリデーションが利用できます。

### JSONファイルの配置

dirtyは以下の順序でJSONファイルを検索します：

1. **環境変数 `DIRTY_EFFECTS_JSON`** で指定されたパス（最優先）
2. **解析対象パッケージディレクトリ**の `dirty-effects.json`

```bash
# 方法1: 環境変数で明示的に指定
DIRTY_EFFECTS_JSON=/path/to/effects.json dirty ./...

# 方法2: パッケージディレクトリに配置（推奨）
myproject/
└── internal/
    └── db/
        ├── queries.go          # 解析対象のコード
        └── dirty-effects.json  # エフェクト宣言
```

### 優先順位

エフェクトの解決順序：
1. **ソースコード内の `// dirty:` コメント**（最優先）
2. JSONファイルでの宣言

ソースコードに`// dirty:`宣言がある場合、JSONの宣言は無視されます。

### 使用例

#### 1. JSONファイルの作成

```bash
cat > dirty-effects.json << EOF
{
  "version": "1.0",
  "effects": {
    "GetUser": "{ select[users] }",
    "CreateUser": "{ insert[users] }",
    "UpdateUserStatus": "{ update[users] | insert[audit_logs] }",
    "SendEmail": "{ network[smtp] | io[filesystem] }"
  }
}
EOF
```

#### 2. コードでの使用

```go
// dirty: { select[users] | insert[logs] }
func ProcessUser(id int64) error {
    user, err := GetUser(id)  // ✓ select[users] は宣言済み
    if err != nil {
        return err
    }
    return logAccess(user.ID)  // ✓ insert[logs] は宣言済み
}

// dirty: { insert[logs] }
func BrokenFunction(id int64) error {
    user, err := GetUser(id)  // ✗ エラー: select[users] が未宣言
    return err
}

// JSONで宣言された関数も、ソースコードで再宣言可能
// dirty: { select[users] | select[cache] }
func GetUser(id int64) (User, error) {
    // この宣言がJSONより優先される
    return User{}, nil
}
```

### sqlc-useとの統合例

[sqlc-use](https://github.com/naoyafurudono/sqlc-use)の出力を変換してdirtyで使用できます：

```bash
# sqlc-useの出力を変換するスクリプト例
sqlc-use analyze | jq '{
  version: "1.0",
  effects: (
    to_entries | map({
      key: .key,
      value: ("{ " + (
        .value | map("\(.operation)[\(.table)]") | join(" | ")
      ) + " }")
    }) | from_entries
  )
}' > dirty-effects.json
```

### CI/CDでの使用例

```yaml
# GitHub Actions
- name: Create effect declarations
  run: |
    cat > dirty-effects.json << EOF
    {
      "version": "1.0",
      "effects": {
        "DatabaseQuery": "{ select[users] | select[posts] }",
        "ExternalAPI": "{ network[api] | insert[logs] }"
      }
    }
    EOF

- name: Run dirty with JSON effects
  run: DIRTY_EFFECTS_JSON=dirty-effects.json dirty ./...
```

### 制限事項

- **ソースコード優先**: ソースコード内の`// dirty:`宣言は常にJSONより優先されます
- **エラー無視**: JSONファイルの読み込みエラーは警告なしに無視されます
- **大文字小文字の区別**: 関数名は大文字小文字を区別します

## 制限

実装をするのが面倒なので、今は色々な実装上のサボりをします。結果的に予期せぬ振る舞いがたくさん生じます。

- モジュール外のエフェクト表明は参照しません
- エフェクトの走査を真面目にやりません
  - 高階関数とかをサポートしません。本来なら型システムがやるようなことをするべきです。
  - 現時点では「関数宣言の中に出現した関数呼び出しのcalleeのエフェクトの和集合」をその関数宣言のエフェクトとします。
  - 無名関数の本体に囲まれていようが関係ないですし、引数として渡された関数を呼び出した場合はそのエフェクトを無視することになります。
