# クロスパッケージ解析 設計書

## 概要

本設計書は、dirtyツールにクロスパッケージ解析機能を実装するための詳細設計を記述する。

## アーキテクチャ

### コンポーネント構成

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Entry Point                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                    Project Analyzer                          │
│  - プロジェクト全体の解析オーケストレーション                    │
│  - 依存関係グラフの構築                                       │
│  - 解析順序の決定                                            │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                   Package Analyzer                           │
│  - 個別パッケージの解析                                       │
│  - 型情報の解決                                             │
│  - エフェクトの収集と伝播                                     │
└──────┬──────────────┴────────────────────┬──────────────────┘
       │                                   │
┌──────▼────────┐                   ┌─────▼──────────┐
│ Type Resolver │                   │ Effect Cache    │
│ (go/types)    │                   │                │
└───────────────┘                   └────────────────┘
```

## データ構造

### 1. プロジェクトレベル

```go
// ProjectAnalysis はプロジェクト全体の解析状態を管理
type ProjectAnalysis struct {
    // プロジェクトルート
    RootDir string
    
    // go.mod情報
    Module *modfile.File
    
    // 解析対象パッケージ
    Packages map[string]*PackageAnalysis
    
    // 依存関係グラフ
    DependencyGraph *DependencyGraph
    
    // グローバルエフェクトレジストリ
    EffectRegistry *EffectRegistry
    
    // 型チェッカー設定
    TypeConfig *types.Config
}

// DependencyGraph はパッケージ間の依存関係を表現
type DependencyGraph struct {
    nodes map[string]*PackageNode
    edges map[string][]string // from -> []to
}

type PackageNode struct {
    Path      string
    Analysis  *PackageAnalysis
    Imports   []string
    Status    AnalysisStatus
}
```

### 2. パッケージレベル

```go
// PackageAnalysis は個別パッケージの解析結果
type PackageAnalysis struct {
    // パッケージ情報
    Package *types.Package
    
    // AST
    Files []*ast.File
    
    // 型情報
    TypeInfo *types.Info
    
    // このパッケージで定義された関数のエフェクト
    LocalEffects map[EffectKey]*EffectInfo
    
    // インポートされた関数のエフェクト（キャッシュ）
    ImportedEffects map[EffectKey]*EffectInfo
    
    // エラー
    Errors []error
}

// EffectKey は関数/メソッドの一意識別子
type EffectKey struct {
    Package  string // フルパッケージパス
    Receiver string // メソッドの場合のレシーバー型
    Name     string // 関数/メソッド名
}

// EffectInfo はエフェクト情報とメタデータ
type EffectInfo struct {
    Effects   *ast.EffectExpr
    Declared  bool        // ソースコードで宣言されているか
    Source    string      // "code", "json", "implicit"
    CallSites []CallSite  // この関数を呼び出している場所
}
```

### 3. 型解決

```go
// TypeResolver は型情報を使った識別子解決
type TypeResolver struct {
    pkg      *types.Package
    typeInfo *types.Info
    imports  map[string]*types.Package
}

// ResolveCall は関数呼び出しからEffectKeyを生成
func (r *TypeResolver) ResolveCall(call *ast.CallExpr) (*EffectKey, error) {
    switch fun := call.Fun.(type) {
    case *ast.Ident:
        // 同一パッケージ内の関数
        obj := r.typeInfo.ObjectOf(fun)
        if fn, ok := obj.(*types.Func); ok {
            return r.effectKeyFromFunc(fn), nil
        }
        
    case *ast.SelectorExpr:
        // メソッド呼び出しまたは他パッケージの関数
        obj := r.typeInfo.ObjectOf(fun.Sel)
        if fn, ok := obj.(*types.Func); ok {
            return r.effectKeyFromFunc(fn), nil
        }
    }
    return nil, fmt.Errorf("cannot resolve call")
}

func (r *TypeResolver) effectKeyFromFunc(fn *types.Func) *EffectKey {
    key := &EffectKey{
        Package: fn.Pkg().Path(),
        Name:    fn.Name(),
    }
    
    // メソッドの場合
    if recv := fn.Type().(*types.Signature).Recv(); recv != nil {
        key.Receiver = r.formatReceiver(recv.Type())
    }
    
    return key
}
```

## 解析フロー

### 1. 初期化フェーズ

```go
func analyzeProject(rootDir string) (*ProjectAnalysis, error) {
    // 1. go.modの読み込み
    modFile, err := loadGoMod(rootDir)
    if err != nil {
        return nil, err
    }
    
    // 2. 解析対象パッケージの列挙
    packages, err := listPackages(rootDir)
    if err != nil {
        return nil, err
    }
    
    // 3. 依存関係グラフの構築
    depGraph := buildDependencyGraph(packages)
    
    // 4. JSON Effects V2の読み込み
    registry, err := loadEffectRegistry(rootDir)
    if err != nil {
        return nil, err
    }
    
    return &ProjectAnalysis{
        RootDir:         rootDir,
        Module:          modFile,
        Packages:        make(map[string]*PackageAnalysis),
        DependencyGraph: depGraph,
        EffectRegistry:  registry,
    }, nil
}
```

### 2. 解析実行フェーズ

```go
func (p *ProjectAnalysis) Analyze() error {
    // トポロジカルソートで解析順序を決定
    order, err := p.DependencyGraph.TopologicalSort()
    if err != nil {
        return err // 循環依存
    }
    
    // 依存順に解析
    for _, pkgPath := range order {
        if err := p.analyzePackage(pkgPath); err != nil {
            // エラーを記録して続行
            p.recordError(pkgPath, err)
        }
    }
    
    // 最終的な整合性チェック
    return p.validateConsistency()
}

func (p *ProjectAnalysis) analyzePackage(pkgPath string) error {
    // 1. パッケージのロードと型チェック
    pkg, typeInfo, err := p.loadPackageWithTypes(pkgPath)
    if err != nil {
        return err
    }
    
    // 2. ローカルエフェクトの収集
    localEffects := p.collectLocalEffects(pkg, typeInfo)
    
    // 3. 呼び出しグラフの構築とエフェクト伝播
    analysis := &PackageAnalysis{
        Package:      pkg,
        TypeInfo:     typeInfo,
        LocalEffects: localEffects,
    }
    
    // 4. インポートされた関数のエフェクト解決
    if err := p.resolveImportedEffects(analysis); err != nil {
        return err
    }
    
    // 5. 暗黙的エフェクトの計算
    if err := p.propagateEffects(analysis); err != nil {
        return err
    }
    
    p.Packages[pkgPath] = analysis
    return nil
}
```

### 3. エフェクト解決

```go
func (p *ProjectAnalysis) resolveImportedEffects(pkg *PackageAnalysis) error {
    resolver := &TypeResolver{
        pkg:      pkg.Package,
        typeInfo: pkg.TypeInfo,
        imports:  pkg.Package.Imports(),
    }
    
    // 各ファイルのASTを走査
    for _, file := range pkg.Files {
        ast.Inspect(file, func(n ast.Node) bool {
            if call, ok := n.(*ast.CallExpr); ok {
                key, err := resolver.ResolveCall(call)
                if err != nil {
                    return true
                }
                
                // エフェクトの検索優先順位:
                // 1. 解析済みパッケージ
                // 2. JSON Effects V2
                // 3. デフォルト（エフェクトなし）
                effects := p.lookupEffects(key)
                pkg.ImportedEffects[*key] = effects
            }
            return true
        })
    }
    
    return nil
}

func (p *ProjectAnalysis) lookupEffects(key *EffectKey) *EffectInfo {
    // 1. 既に解析済みのパッケージから検索
    if pkg, ok := p.Packages[key.Package]; ok {
        if info, ok := pkg.LocalEffects[*key]; ok {
            return info
        }
    }
    
    // 2. JSON Effects V2から検索
    if effects := p.EffectRegistry.Lookup(key); effects != nil {
        return &EffectInfo{
            Effects:  effects,
            Declared: false,
            Source:   "json",
        }
    }
    
    // 3. デフォルト
    return &EffectInfo{
        Effects:  ast.NewLiteralSet(),
        Declared: false,
        Source:   "default",
    }
}
```

## キャッシュシステム

```go
// EffectCache はパッケージのエフェクト情報をキャッシュ
type EffectCache struct {
    dir string
    mu  sync.RWMutex
}

type CachedPackageEffects struct {
    Version   string                 // パッケージのバージョン/ハッシュ
    Effects   map[EffectKey]*EffectInfo
    Timestamp time.Time
}

func (c *EffectCache) Load(pkgPath string) (*CachedPackageEffects, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    cachePath := filepath.Join(c.dir, url.QueryEscape(pkgPath)+".json")
    data, err := os.ReadFile(cachePath)
    if err != nil {
        return nil, err
    }
    
    var cached CachedPackageEffects
    if err := json.Unmarshal(data, &cached); err != nil {
        return nil, err
    }
    
    return &cached, nil
}

func (c *EffectCache) Save(pkgPath string, effects *CachedPackageEffects) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    cachePath := filepath.Join(c.dir, url.QueryEscape(pkgPath)+".json")
    data, err := json.MarshalIndent(effects, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(cachePath, data, 0644)
}
```

## エラーハンドリング

```go
// AnalysisError は解析中のエラーを表現
type AnalysisError struct {
    Package string
    File    string
    Pos     token.Pos
    Message string
    Type    ErrorType
}

type ErrorType int

const (
    ErrorTypeEffect ErrorType = iota
    ErrorTypeImport
    ErrorTypeCircular
    ErrorTypeType
)

// ErrorCollector はエラーを収集し、最終的にレポート
type ErrorCollector struct {
    errors []AnalysisError
    mu     sync.Mutex
}

func (ec *ErrorCollector) Add(err AnalysisError) {
    ec.mu.Lock()
    defer ec.mu.Unlock()
    ec.errors = append(ec.errors, err)
}

func (ec *ErrorCollector) Report() error {
    if len(ec.errors) == 0 {
        return nil
    }
    
    // エラーをパッケージごとにグループ化してレポート
    byPackage := make(map[string][]AnalysisError)
    for _, err := range ec.errors {
        byPackage[err.Package] = append(byPackage[err.Package], err)
    }
    
    // フォーマットして出力
    var buf strings.Builder
    for pkg, errs := range byPackage {
        fmt.Fprintf(&buf, "Package %s:\n", pkg)
        for _, err := range errs {
            fmt.Fprintf(&buf, "  %s: %s\n", err.File, err.Message)
        }
    }
    
    return errors.New(buf.String())
}
```

## パフォーマンス最適化

### 1. 並列解析

```go
func (p *ProjectAnalysis) AnalyzeParallel() error {
    // 依存関係に基づいてレイヤーを作成
    layers := p.DependencyGraph.Layers()
    
    for _, layer := range layers {
        // 同一レイヤー内のパッケージは並列解析可能
        var wg sync.WaitGroup
        errors := make(chan error, len(layer))
        
        for _, pkgPath := range layer {
            wg.Add(1)
            go func(path string) {
                defer wg.Done()
                if err := p.analyzePackage(path); err != nil {
                    errors <- err
                }
            }(pkgPath)
        }
        
        wg.Wait()
        close(errors)
        
        // エラー処理
        for err := range errors {
            if err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

### 2. インクリメンタル解析

```go
func (p *ProjectAnalysis) AnalyzeIncremental(changedFiles []string) error {
    // 変更されたファイルが属するパッケージを特定
    affectedPackages := p.findAffectedPackages(changedFiles)
    
    // 影響を受けるパッケージを再解析
    for _, pkgPath := range affectedPackages {
        // キャッシュを無効化
        p.invalidateCache(pkgPath)
        
        // 再解析
        if err := p.analyzePackage(pkgPath); err != nil {
            return err
        }
        
        // 依存パッケージも再解析が必要かチェック
        for _, dep := range p.DependencyGraph.GetDependents(pkgPath) {
            if p.needsReanalysis(dep, pkgPath) {
                affectedPackages = append(affectedPackages, dep)
            }
        }
    }
    
    return nil
}
```

## 統合テスト戦略

```go
// testdata/crosspackage/
// ├── go.mod
// ├── pkg1/
// │   └── effects.go    // dirty: 宣言あり
// ├── pkg2/
// │   └── caller.go     // pkg1を呼び出す
// └── pkg3/
//     └── transitive.go // pkg2を呼び出す（推移的）

func TestCrossPackageAnalysis(t *testing.T) {
    proj, err := analyzeProject("testdata/crosspackage")
    if err != nil {
        t.Fatal(err)
    }
    
    // pkg3の関数がpkg1のエフェクトを正しく要求するか確認
    pkg3 := proj.Packages["testdata/crosspackage/pkg3"]
    funcInfo := pkg3.LocalEffects[EffectKey{
        Package: "testdata/crosspackage/pkg3",
        Name:    "ProcessTransitive",
    }]
    
    // pkg1のエフェクトが伝播していることを確認
    expected := "{ select[users] | transform }"
    if funcInfo.Effects.String() != expected {
        t.Errorf("Expected effects %s, got %s", expected, funcInfo.Effects)
    }
}
```

## まとめ

この設計により、dirtyツールは以下を実現する：

1. **正確な型解決** - go/typesを使用した完全修飾名での関数識別
2. **効率的な解析** - 依存順序での解析と並列処理
3. **スケーラビリティ** - キャッシュとインクリメンタル解析
4. **拡張性** - JSON Effects V2との統合

実装は3つのフェーズで段階的に行い、各フェーズで実用的な機能を提供する。