# クロスパッケージ解析の再設計案

## 現状の問題点

1. **同一モジュール内のパッケージ間解析が機能しない**
   - 現在はJSONファイルに依存した外部パッケージ対応のみ
   - 同じGoモジュール内の他パッケージのソースコードを解析できない

2. **singlecheckerの制限**
   - 単一パッケージ解析のみサポート
   - マルチパッケージ解析には対応していない

## 修正方針：Factメカニズムの活用

`golang.org/x/tools/go/analysis`パッケージが提供するFactメカニズムを使用して、パッケージ間で効果情報を共有する。

### Factメカニズムとは

- analysisパッケージが提供するパッケージ間情報共有の仕組み
- 各パッケージの解析結果を「Fact」として保存し、依存パッケージから参照可能
- JSONファイルなしで自動的にクロスパッケージ解析が実現

### 実装アプローチ

#### 1. PackageFact型の定義

```go
// PackageEffectsFact はパッケージ内の全関数の効果情報を保持
type PackageEffectsFact struct {
    // 関数名 -> 効果のマップ
    FunctionEffects map[string]StringSet
}

// AFact marker method
func (*PackageEffectsFact) AFact() {}
```

#### 2. ObjectFact型の定義（オプション）

```go
// FunctionEffectsFact は個別の関数の効果情報を保持
type FunctionEffectsFact struct {
    Effects StringSet
}

func (*FunctionEffectsFact) AFact() {}
```

#### 3. 解析フローの変更

##### Phase 1: 効果情報の収集とエクスポート
```go
// 現在のパッケージの全関数の効果を収集
packageFact := &PackageEffectsFact{
    FunctionEffects: make(map[string]StringSet),
}

// 各関数の効果を収集
for funcName, info := range effectAnalysis.Functions {
    // 完全修飾名で保存（例: "github.com/user/pkg.FuncName"）
    qualifiedName := pass.Pkg.Path() + "." + funcName
    packageFact.FunctionEffects[qualifiedName] = info.ComputedEffects
    
    // オプション：個別の関数にもFactを付与
    if info.Decl != nil {
        fact := &FunctionEffectsFact{Effects: info.ComputedEffects}
        pass.ExportObjectFact(info.Decl.Name, fact)
    }
}

// パッケージFactをエクスポート
pass.ExportPackageFact(packageFact)
```

##### Phase 2: 依存パッケージの効果情報インポート
```go
// インポートされたパッケージから効果情報を取得
func getImportedEffects(pass *analysis.Pass, pkgPath string, funcName string) (StringSet, bool) {
    // まずPackageFactから検索
    for _, imp := range pass.Pkg.Imports() {
        if imp.Path() == pkgPath {
            var fact PackageEffectsFact
            if pass.ImportPackageFact(imp, &fact) {
                qualifiedName := pkgPath + "." + funcName
                if effects, ok := fact.FunctionEffects[qualifiedName]; ok {
                    return effects, true
                }
            }
        }
    }
    
    // 見つからない場合はJSONフォールバック（後方互換性）
    return getEffectsFromJSON(pkgPath, funcName)
}
```

##### Phase 3: クロスパッケージ呼び出しの解析
```go
// EnhanceWithCrossPackageSupport の改良版
func EnhanceWithCrossPackageSupport(ea *EffectAnalysis) {
    // ... 既存のインポート解析 ...
    
    // クロスパッケージ呼び出しの検出
    ast.Inspect(info.Decl, func(n ast.Node) bool {
        call, ok := n.(*ast.CallExpr)
        if !ok {
            return true
        }
        
        // セレクタ式の場合（pkg.Func形式）
        if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
            if ident, ok := sel.X.(*ast.Ident); ok {
                if pkgPath, isImport := imports[ident.Name]; isImport {
                    // Factから効果情報を取得
                    if effects, found := getImportedEffects(ea.Pass, pkgPath, sel.Sel.Name); found {
                        // 効果情報を使用
                        resolvedName := pkgPath + "." + sel.Sel.Name
                        ea.Functions[resolvedName] = &FunctionInfo{
                            Name:            resolvedName,
                            Package:         pkgPath,
                            DeclaredEffects: effects,
                            ComputedEffects: effects,
                            HasDeclaration:  true,
                        }
                        ea.CallGraph.AddCall(funcName, resolvedName, call.Pos())
                    }
                }
            }
        }
        return true
    })
}
```

### 利点

1. **自動的なクロスパッケージ解析**
   - JSONファイル不要で同一モジュール内の解析が可能
   - go vetやgolangci-lintとの統合が容易

2. **段階的な移行**
   - JSONフォールバックにより後方互換性を維持
   - 既存のユーザーへの影響を最小限に

3. **analysisパッケージの設計思想に合致**
   - 標準的な方法でパッケージ間情報を共有
   - 他のアナライザーとの相互運用性

### 実装ステップ

1. **Fact型の定義と基本実装**
   - PackageEffectsFact型の実装
   - 基本的なエクスポート/インポート処理

2. **既存コードの段階的移行**
   - EnhanceWithCrossPackageSupport関数の改良
   - Factベースの効果取得実装

3. **テストとドキュメント**
   - マルチパッケージテストケースの作成
   - 使用方法のドキュメント更新

4. **最適化とクリーンアップ**
   - パフォーマンス最適化
   - 不要になったコードの削除

### 考慮事項

1. **メソッドの扱い**
   - レシーバー付きメソッドの完全修飾名の形式
   - インターフェース経由の呼び出しへの対応

2. **循環依存**
   - パッケージ間の循環依存の検出と処理

3. **パフォーマンス**
   - 大規模プロジェクトでの効率的な動作
   - Factのサイズとメモリ使用量

この方針により、真のクロスパッケージ解析が実現し、dirtyツールの有用性が大幅に向上します。