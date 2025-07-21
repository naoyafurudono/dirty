# Phase 2: 基本的なクロスパッケージ解析 - 実装計画

## 概要

go/typesを使わず、現在のAST解析ベースのアプローチを拡張して、基本的なクロスパッケージ解析を実現する。

## 方針

1. **シンプルな実装を優先**
   - 現在のanalyzer.goの構造を大きく変えない
   - パッケージパスを含む関数識別子を導入
   - 複数パッケージを順次解析

2. **制限事項を受け入れる**
   - インターフェース経由の呼び出しは追跡しない
   - 型エイリアスは考慮しない
   - ベンダリングは考慮しない

## 実装ステップ

### Step 1: テストケースの作成

```
testdata/src/crosspackage/
├── go.mod
├── pkg1/
│   └── db.go          // エフェクト宣言あり
├── pkg2/
│   └── service.go     // pkg1を使用
└── pkg3/
    └── handler.go     // pkg2を使用（推移的）
```

### Step 2: 関数識別子の拡張

現在の実装:
```go
// 関数名のみ
"GetUser" 
```

拡張後:
```go
// パッケージパス付き
"github.com/example/pkg1.GetUser"
"github.com/example/pkg2.(*UserService).GetUser"
```

### Step 3: パッケージ解析の拡張

```go
type CrossPackageAnalyzer struct {
    // 解析済みパッケージのエフェクト情報
    packageEffects map[string]map[string]*EffectSet
    
    // 現在解析中のパッケージ
    currentPackage string
    
    // インポート情報
    imports map[string]string // alias -> package path
}
```

### Step 4: インポート解析

```go
func (a *CrossPackageAnalyzer) analyzeImports(file *ast.File) {
    for _, imp := range file.Imports {
        path := strings.Trim(imp.Path.Value, `"`)
        alias := ""
        if imp.Name != nil {
            alias = imp.Name.Name
        } else {
            // パッケージ名を推測
            alias = filepath.Base(path)
        }
        a.imports[alias] = path
    }
}
```

### Step 5: 関数呼び出しの解決

```go
func (a *CrossPackageAnalyzer) resolveCallExpr(call *ast.CallExpr) string {
    switch fun := call.Fun.(type) {
    case *ast.Ident:
        // 同一パッケージ内
        return a.currentPackage + "." + fun.Name
        
    case *ast.SelectorExpr:
        // 他パッケージまたはメソッド
        if ident, ok := fun.X.(*ast.Ident); ok {
            if pkgPath, isImport := a.imports[ident.Name]; isImport {
                // インポートされたパッケージの関数
                return pkgPath + "." + fun.Sel.Name
            }
            // メソッド呼び出し（簡易版）
            return a.currentPackage + ".Method:" + fun.Sel.Name
        }
    }
    return ""
}
```

### Step 6: マルチパッケージ解析

```go
func AnalyzeProject(rootDir string) error {
    analyzer := &CrossPackageAnalyzer{
        packageEffects: make(map[string]map[string]*EffectSet),
    }
    
    // 1. パッケージリストの取得
    packages := listPackages(rootDir)
    
    // 2. 各パッケージを解析（簡易的な依存順）
    for _, pkg := range packages {
        if err := analyzer.analyzePackage(pkg); err != nil {
            return err
        }
    }
    
    // 3. 整合性チェック
    return analyzer.validateAllPackages()
}
```

## テストシナリオ

### 基本シナリオ

```go
// pkg1/db.go
package pkg1

// dirty: { select[users] }
func GetUser(id int) User { ... }

// pkg2/service.go
package pkg2

import "crosspackage/pkg1"

// dirty: { select[users] | transform }  // pkg1.GetUserのエフェクトも必要
func ProcessUser(id int) ProcessedUser {
    user := pkg1.GetUser(id)
    return transform(user)
}
```

### エラーケース

```go
// pkg2/service.go
// dirty: { transform }  // エラー: select[users]が不足
func ProcessUser(id int) ProcessedUser {
    user := pkg1.GetUser(id)  // requires { select[users] }
    return transform(user)
}
```

## 制限事項と今後の課題

1. **型情報なしの制限**
   - メソッドのレシーバー型を正確に識別できない
   - インターフェース経由の呼び出しを追跡できない

2. **パフォーマンス**
   - 全パッケージを毎回解析
   - キャッシュ機構なし

3. **スケーラビリティ**
   - 大規模プロジェクトでは遅い
   - メモリ使用量が多い

これらはPhase 1（型システム統合）やPhase 3（高度な機能）で対処する。

## 実装優先順位

1. テストケースの作成
2. 基本的な動作確認
3. エラーメッセージの改善
4. ドキュメント更新

この実装により、基本的なクロスパッケージ解析が可能になり、実用的な使用が開始できる。