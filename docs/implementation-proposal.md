# 実装提案: 暗黙的エフェクト計算の実現

## 現状の課題

現在の実装では、`//dirty:`コメントを持つ関数のみがチェック対象となっており、コメントを持たない関数を経由したエフェクトの伝播が考慮されていません。

```go
// 現在の実装では、このケースが正しく処理されない
func intermediate() {  // エフェクト宣言なし
    GetUser(1)  // select[user]のエフェクトを持つ
}

//dirty: select[member]
func caller() {
    intermediate()  // ここでselect[user]のエフェクトが必要だが、検出されない
}
```

## 提案する実装アプローチ

### 1. Two-Pass Analysis

#### Pass 1: エフェクト収集と呼び出しグラフ構築
```go
type EffectAnalysis struct {
    // 関数名 -> 関数情報
    Functions map[string]*FunctionInfo
    // 呼び出しグラフ
    CallGraph *CallGraph
}

type FunctionInfo struct {
    Name            string
    Package         string
    DeclaredEffects StringSet // //dirty:で宣言されたエフェクト
    ComputedEffects StringSet // 計算された実際のエフェクト
    HasDeclaration  bool
    Decl            *ast.FuncDecl
    CallSites       []CallSite // この関数から呼び出される関数
}

type CallSite struct {
    Callee   string
    Position token.Pos
}
```

#### Pass 2: エフェクト伝播
```go
func (a *EffectAnalysis) PropagateEffects() {
    // ワークリストアルゴリズムを使用
    worklist := make([]string, 0, len(a.Functions))
    inWorklist := make(map[string]bool)
    
    // すべての関数をワークリストに追加
    for name := range a.Functions {
        worklist = append(worklist, name)
        inWorklist[name] = true
    }
    
    for len(worklist) > 0 {
        // ワークリストから関数を取り出す
        funcName := worklist[len(worklist)-1]
        worklist = worklist[:len(worklist)-1]
        inWorklist[funcName] = false
        
        fn := a.Functions[funcName]
        oldEffects := fn.ComputedEffects.Clone()
        
        // 呼び出し先のエフェクトを収集
        for _, call := range fn.CallSites {
            if callee, ok := a.Functions[call.Callee]; ok {
                fn.ComputedEffects.AddAll(callee.ComputedEffects)
            }
        }
        
        // エフェクトが変化した場合、この関数を呼び出す関数を再計算
        if !oldEffects.Equals(fn.ComputedEffects) {
            for _, caller := range a.CallGraph.CalledBy[funcName] {
                if !inWorklist[caller] {
                    worklist = append(worklist, caller)
                    inWorklist[caller] = true
                }
            }
        }
    }
}
```

### 2. StringSet実装（エフェクトの集合演算）

```go
type StringSet map[string]struct{}

func NewStringSet(items ...string) StringSet {
    s := make(StringSet)
    for _, item := range items {
        s[item] = struct{}{}
    }
    return s
}

func (s StringSet) Add(item string) {
    s[item] = struct{}{}
}

func (s StringSet) AddAll(other StringSet) {
    for item := range other {
        s[item] = struct{}{}
    }
}

func (s StringSet) Contains(item string) bool {
    _, ok := s[item]
    return ok
}

func (s StringSet) Clone() StringSet {
    clone := make(StringSet)
    for item := range s {
        clone[item] = struct{}{}
    }
    return clone
}

func (s StringSet) Equals(other StringSet) bool {
    if len(s) != len(other) {
        return false
    }
    for item := range s {
        if !other.Contains(item) {
            return false
        }
    }
    return true
}

func (s StringSet) ToSlice() []string {
    result := make([]string, 0, len(s))
    for item := range s {
        result = append(result, item)
    }
    sort.Strings(result)
    return result
}
```

### 3. 改良されたrun関数

```go
func run(pass *analysis.Pass) (interface{}, error) {
    analysis := &EffectAnalysis{
        Functions: make(map[string]*FunctionInfo),
        CallGraph: &CallGraph{
            Calls:    make(map[string][]CallSite),
            CalledBy: make(map[string][]string),
        },
    }
    
    // Pass 1: 関数とエフェクトを収集
    collectFunctionsAndEffects(pass, analysis)
    
    // Pass 2: 呼び出し関係を解析
    buildCallGraph(pass, analysis)
    
    // Pass 3: エフェクトを伝播
    analysis.PropagateEffects()
    
    // Pass 4: エフェクトの整合性をチェック
    checkEffectConsistency(pass, analysis)
    
    return nil, nil
}
```

## メソッド呼び出しの対応

```go
func resolveMethodCall(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
    switch fun := call.Fun.(type) {
    case *ast.SelectorExpr:
        // レシーバの型を解決
        if sel := pass.TypesInfo.Selections[fun]; sel != nil {
            recv := sel.Recv()
            method := sel.Obj().Name()
            
            // 型名を取得（ポインタ型の場合は基底型を使用）
            typeName := getTypeName(recv)
            if typeName != "" {
                return fmt.Sprintf("(%s).%s", typeName, method), true
            }
        }
    }
    return "", false
}
```

## 循環参照の最適化

```go
func (a *EffectAnalysis) findStronglyConnectedComponents() [][]string {
    // Tarjanのアルゴリズムを使用してSCCを検出
    // SCCごとにまとめて処理することで効率化
}
```

## エラー報告の改善

```go
type EffectError struct {
    Function        string
    Position        token.Pos
    DeclaredEffects []string
    RequiredEffects []string
    MissingEffects  []string
    CallChain       []CallChainEntry
}

type CallChainEntry struct {
    Function string
    Position token.Pos
    Effects  []string
}

func (e *EffectError) Format() string {
    var buf strings.Builder
    
    fmt.Fprintf(&buf, "function %s has undeclared effects: %s\n",
        e.Function, strings.Join(e.MissingEffects, ", "))
    
    if len(e.CallChain) > 0 {
        fmt.Fprintln(&buf, "\nEffect propagation:")
        for i, entry := range e.CallChain {
            indent := strings.Repeat("  ", i)
            fmt.Fprintf(&buf, "%s%s (effects: %s)\n",
                indent, entry.Function, strings.Join(entry.Effects, ", "))
        }
    }
    
    return buf.String()
}
```

## 段階的な実装計画

### Phase 1: 基本的な暗黙的エフェクト計算（1週間）
- [x] StringSetの実装
- [ ] 基本的なエフェクト伝播
- [ ] 単純な関数呼び出しのサポート

### Phase 2: 完全な呼び出しグラフ（1週間）
- [ ] メソッド呼び出しのサポート
- [ ] 呼び出しグラフの可視化（デバッグ用）
- [ ] 循環参照の処理

### Phase 3: エラー報告の改善（3日）
- [ ] 詳細なエラーメッセージ
- [ ] エフェクト伝播経路の表示
- [ ] 修正提案の生成

### Phase 4: 最適化（3日）
- [ ] SCCベースの最適化
- [ ] インクリメンタル解析
- [ ] 並列処理

## テスト戦略

```go
func TestImplicitEffectPropagation(t *testing.T) {
    tests := []struct {
        name string
        code string
        want map[string][]string // 関数名 -> 計算されたエフェクト
    }{
        {
            name: "simple propagation",
            code: `
                //dirty: select[user]
                func getUser() {}
                
                func intermediate() {
                    getUser()
                }
                
                //dirty: select[member]
                func top() {
                    intermediate()
                }
            `,
            want: map[string][]string{
                "getUser":      {"select[user]"},
                "intermediate": {"select[user]"},
                "top":          {"select[member]", "select[user]"},
            },
        },
    }
    // ...
}
```

## まとめ

この実装提案により、dirtyは仕様に完全に準拠した動作を実現できます。段階的な実装により、各フェーズで動作する状態を保ちながら、最終的に完全な機能を提供できます。