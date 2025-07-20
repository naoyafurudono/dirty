# エフェクト宣言の拡張可能な文法設計

## 概要

現在の単純なラベル列挙から、より表現力のある集合ベースの文法への移行を提案します。

## 設計目標

1. **拡張性**: 将来的な機能追加が容易
2. **簡潔性**: 冗長でない、読みやすい文法
3. **一貫性**: 新しい文法に統一
4. **段階的実装**: 小さなステップで実装可能

## 提案する文法

### Phase 1: 基本的な集合記法

```go
// 新しい集合記法（これが標準）
//dirty: { select[users] | insert[logs] }

// 単一要素の場合も集合記法
//dirty: { select[users] }

// 空集合も可能
//dirty: { }
```

### Phase 2: 名前付きエフェクト（将来）

```go
// エフェクトセットの定義
//dirty-define: userOps = { select[users] | update[users] | delete[users] }
//dirty-define: auditOps = { insert[audit_logs] }

// 参照
//dirty: userOps | auditOps
//dirty: userOps | { insert[sessions] }
```

### Phase 3: 高度な集合演算（将来）

```go
// 差集合（除外）
//dirty: userOps \ { delete[users] }  // deleteを除くuserOps

// 交差
//dirty: userOps & dbOps  // 共通部分のみ
```

## AST（抽象構文木）の設計

ASTは、パースした文法を木構造で表現したものです。例えば `{ select[users] | insert[logs] | update[logs] }` は以下のような構造になります：

```
LiteralSet
    |
    +-- Elements[0]: EffectLabel("select[users]")
    |
    +-- Elements[1]: EffectLabel("insert[logs]")
    |
    +-- Elements[2]: EffectLabel("update[logs]")
```

### 提案するAST型定義

```go
// エフェクト式を表す基本インターフェース
type EffectExpr interface {
    // 式を評価して、エフェクトラベルの集合を返す
    Eval(resolver EffectResolver) (StringSet, error)
    // デバッグ用の文字列表現
    String() string
}

// 単一のエフェクトラベル（葉ノード）
type EffectLabel struct {
    Operation string  // "select", "insert", "update", "delete"
    Target    string  // "users", "logs", etc.
}

func (e *EffectLabel) Eval(resolver EffectResolver) (StringSet, error) {
    return NewStringSet(e.String()), nil
}

func (e *EffectLabel) String() string {
    return fmt.Sprintf("%s[%s]", e.Operation, e.Target)
}

// リテラル集合（要素を明示的に列挙）
// { a | b | c } の全体を表す
type LiteralSet struct {
    Elements []EffectExpr  // 集合の要素のスライス
}

func (s *LiteralSet) Eval(resolver EffectResolver) (StringSet, error) {
    result := NewStringSet()
    for _, elem := range s.Elements {
        set, err := elem.Eval(resolver)
        if err != nil {
            return nil, err
        }
        result = result.Union(set)
    }
    return result, nil
}

func (s *LiteralSet) String() string {
    parts := make([]string, len(s.Elements))
    for i, elem := range s.Elements {
        parts[i] = elem.String()
    }
    return fmt.Sprintf("{ %s }", strings.Join(parts, " | "))
}

// エフェクト参照（Phase 2以降）
type EffectRef struct {
    Name string  // e.g., "userOps"
}

// リゾルバーインターフェース（名前付きエフェクトの解決）
type EffectResolver interface {
    Resolve(name string) (StringSet, error)
}
```

### パース結果の例

入力: `//dirty: { select[users] | insert[logs] | update[users] }`

パース結果のAST:
```go
&LiteralSet{
    Elements: []EffectExpr{
        &EffectLabel{
            Operation: "select",
            Target: "users",
        },
        &EffectLabel{
            Operation: "insert",
            Target: "logs",
        },
        &EffectLabel{
            Operation: "update",
            Target: "users",
        },
    },
}
```

評価結果: `StringSet{"select[users]", "insert[logs]", "update[users]"}`

### ASTの構造の説明

1. **LiteralSet**: リテラル集合（要素を明示的に列挙）
   - `{ ... }` で囲まれた部分全体を表す
   - `Elements`スライスに、`|`で区切られた各要素を保持

2. **EffectLabel**: 個々のエフェクトラベル
   - `select[users]`のような単一の要素

3. **将来の拡張性**:
   - Phase 2で`EffectRef`（名前付きエフェクトの参照）を追加可能
   - Phase 3で他の集合演算（差集合、交差など）を追加可能

### FunctionInfoの更新

```go
type FunctionInfo struct {
    // 宣言されたエフェクト式（パース結果）
    DeclaredExpr    EffectExpr
    // 評価済みのエフェクト
    DeclaredEffects StringSet
    ComputedEffects StringSet
}
```

## パーサーの設計

### 文法（EBNF風）

```
effect_decl  = set_expr
set_expr     = "{" [ union_expr ] "}"
union_expr   = primary { "|" primary }
primary      = effect_label | effect_ref | "(" union_expr ")"
effect_label = IDENT "[" IDENT "]"
effect_ref   = IDENT  // Phase 2以降
```

### パーサー実装方針

```go
// トークンの種類
type TokenType int

const (
    TOKEN_EOF TokenType = iota
    TOKEN_LBRACE      // {
    TOKEN_RBRACE      // }
    TOKEN_LPAREN      // (
    TOKEN_RPAREN      // )
    TOKEN_PIPE        // |
    TOKEN_LBRACKET    // [
    TOKEN_RBRACKET    // ]
    TOKEN_IDENT       // 識別子
)

// レキサー
type Lexer struct {
    input string
    pos   int
}

// パーサー
type Parser struct {
    lexer *Lexer
    cur   Token
}

// メインのパース関数
func ParseEffectDecl(comment string) (EffectExpr, error) {
    // "//dirty:" プレフィックスを削除
    content := strings.TrimPrefix(strings.TrimSpace(comment), "//dirty:")
    content = strings.TrimSpace(content)

    parser := &Parser{
        lexer: &Lexer{input: content},
    }
    parser.next() // 最初のトークンを読む

    return parser.parseSetExpr()
}

// 集合式のパース: { ... }
func (p *Parser) parseSetExpr() (EffectExpr, error) {
    if p.cur.Type != TOKEN_LBRACE {
        return nil, fmt.Errorf("expected '{', got %s", p.cur.Value)
    }
    p.next() // skip {

    // 空集合の場合
    if p.cur.Type == TOKEN_RBRACE {
        p.next() // skip }
        return &LiteralSet{Elements: []EffectExpr{}}, nil
    }

    // 要素をパース
    elements := []EffectExpr{}
    for {
        elem, err := p.parsePrimary()
        if err != nil {
            return nil, err
        }
        elements = append(elements, elem)

        // 次が | なら続ける
        if p.cur.Type == TOKEN_PIPE {
            p.next() // skip |
            continue
        }

        // } なら終了
        if p.cur.Type == TOKEN_RBRACE {
            p.next() // skip }
            break
        }

        return nil, fmt.Errorf("expected '|' or '}', got %s", p.cur.Value)
    }

    return &LiteralSet{Elements: elements}, nil
}
```

## 実装計画

### Phase 1: 基本実装（このIssueの範囲）

1. **AST定義**
   - `EffectExpr`インターフェースと基本型
   - `EffectLabel`と`UnionExpr`の実装

2. **パーサー実装**
   - 新しい文法のパーサー
   - 後方互換性の維持

3. **評価器**
   - AST→StringSetへの評価

4. **既存コードの更新**
   - `ParseEffects`を新パーサーで置き換え
   - テストの更新

### 実装の順序

1. **AST定義とパーサー実装**
   - `EffectExpr`インターフェースと基本型
   - レキサーとパーサーの実装
   - エラーハンドリング

2. **既存コードの更新**
   - `ParseEffects`関数を新しいパーサーで置き換え
   - すべてのテストケースを新文法に更新
   - ドキュメントとサンプルコードの更新

3. **段階的な機能追加**
   - Phase 1: 基本的な集合記法のみ
   - Phase 2: 名前付きエフェクトの定義と参照（将来）
   - Phase 3: 差集合などの高度な演算（将来）

## テスト計画

### パーサーテスト

```go
tests := []struct {
    input    string
    expected []string
    wantErr  bool
}{
    // 基本的なケース
    {"//dirty: { select[users] }",
     []string{"select[users]"}, false},

    // 和集合
    {"//dirty: { select[users] | insert[logs] }",
     []string{"select[users]", "insert[logs]"}, false},

    // 3つ以上の和集合
    {"//dirty: { select[users] | insert[logs] | update[users] }",
     []string{"select[users]", "insert[logs]", "update[users]"}, false},

    // 空集合
    {"//dirty: { }",
     []string{}, false},

    // 括弧を使った優先順位（将来の拡張のため）
    {"//dirty: { (select[users] | select[posts]) | insert[logs] }",
     []string{"select[users]", "select[posts]", "insert[logs]"}, false},

    // エラーケース
    {"//dirty: select[users]",  // {} がない
     nil, true},
    {"//dirty: { select[users }",  // } がない
     nil, true},
    {"//dirty: { select users ]",  // [] の構文エラー
     nil, true},
}
```

### 既存テストの更新

すべての既存テストケースを新しい文法に更新：

```go
// 変更前
//dirty: select[user], insert[log]

// 変更後
//dirty: { select[user] | insert[log] }
```

## リスクと対策

1. **パフォーマンス**
   - リスク: AST評価のオーバーヘッド
   - 対策: 評価結果をキャッシュ

2. **複雑性**
   - リスク: 実装が複雑になる
   - 対策: シンプルなPhase 1から始める

3. **エラーメッセージ**
   - リスク: 新しい文法でエラーが分かりにくい
   - 対策: 丁寧なエラーメッセージとヒント

## まとめ

この設計により、より表現力のある文法に移行し、将来の拡張に備えた基盤を構築できます。

### 主な変更点

1. **統一された文法**: すべてのエフェクト宣言が `{ ... }` 形式に統一
2. **構造化された内部表現**: ASTによる柔軟な表現が可能
3. **拡張可能な設計**: 将来的な機能追加が容易

### Phase 1で実現されること

- 集合記法 `{ a | b | c }`
- 空集合 `{ }`
- より良いエラーメッセージ
- 将来の拡張への準備
