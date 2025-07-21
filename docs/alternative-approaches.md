# クロスパッケージ解析の代替実装方針

## 現在の課題
- `singlechecker.Main`は単一パッケージのみ解析
- 依存パッケージのソースコードにアクセスできない
- Factメカニズムは複雑で、テストが困難

## 代替案の検討

### 1. go/packagesを直接使用する独自ドライバー

**概要**: analysisフレームワークを使わず、go/packagesで全パッケージを読み込む

```go
// cmd/dirty-multi/main.go
func main() {
    cfg := &packages.Config{
        Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | 
              packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
    }
    
    pkgs, err := packages.Load(cfg, patterns...)
    
    // 第1パス: 全パッケージの効果情報を収集
    allEffects := make(map[string]map[string]StringSet) // pkg -> func -> effects
    for _, pkg := range pkgs {
        effects := analyzePackage(pkg)
        allEffects[pkg.PkgPath] = effects
    }
    
    // 第2パス: 収集した情報を使って検証
    for _, pkg := range pkgs {
        verifyPackage(pkg, allEffects)
    }
}
```

**利点**:
- 完全な制御が可能
- 全パッケージの情報に同時アクセス
- シンプルで理解しやすい

**欠点**:
- analysisフレームワークの恩恵を失う
- go vetとの統合が困難
- 独自のエラー報告機構が必要

### 2. ビルドキャッシュを活用したアプローチ

**概要**: go buildのキャッシュディレクトリに効果情報を保存

```go
// 効果情報をキャッシュに保存
func cacheEffects(pkg *types.Package, effects map[string]StringSet) {
    cacheDir := filepath.Join(os.Getenv("GOCACHE"), "dirty-effects")
    hash := packageHash(pkg)
    cachePath := filepath.Join(cacheDir, hash+".json")
    saveEffects(cachePath, effects)
}

// キャッシュから効果情報を読み込み
func loadCachedEffects(pkgPath string) map[string]StringSet {
    // キャッシュから読み込み
}
```

**利点**:
- goのビルドシステムと統合
- 増分解析が可能
- 再ビルド時の高速化

**欠点**:
- キャッシュ管理が複雑
- goのバージョン間で互換性問題の可能性

### 3. プリプロセッサアプローチ

**概要**: 事前に全パッケージを解析してeffect-registry.jsonを生成

```bash
# ステップ1: 効果情報を抽出
dirty-extract ./... > .dirty-effects.json

# ステップ2: 通常の解析（生成されたJSONを使用）
dirty-check ./...
```

**利点**:
- 既存のJSONベース実装を活用
- CIでの段階的実行が容易
- デバッグが簡単

**欠点**:
- 2段階実行が必要
- リアルタイム解析ではない

### 4. go vetプラグインとして実装

**概要**: go vetの-vettoolフラグを活用

```go
// go vetは内部でmulticheckerを使用し、パッケージ間の情報共有が可能
// ただし、非公開APIを使う必要がある
```

**利点**:
- go vetの既存インフラを活用
- 標準的なワークフローに統合

**欠点**:
- go vetの内部実装に依存
- 将来的な互換性の懸念

### 5. ハイブリッドアプローチ（推奨）

**概要**: 同一モジュール内とモジュール間で異なる戦略を採用

```go
func analyzeWithHybridApproach(patterns []string) {
    // 1. 現在のモジュールを特定
    currentModule := getCurrentModule()
    
    // 2. パターンを分類
    internalPkgs, externalPkgs := classifyPackages(patterns, currentModule)
    
    // 3. 内部パッケージはgo/packagesで一括解析
    if len(internalPkgs) > 0 {
        analyzeInternalPackages(internalPkgs) // go/packages使用
    }
    
    // 4. 外部パッケージは既存のJSON方式
    if len(externalPkgs) > 0 {
        analyzeExternalPackages(externalPkgs) // JSON使用
    }
}
```

**利点**:
- 実用的で段階的に実装可能
- 既存コードの多くを再利用
- 外部依存のJSONサポートを維持

**欠点**:
- 2つの解析パスが必要
- コードがやや複雑

## 推奨方針

**短期的**: ハイブリッドアプローチ（案5）を採用
- 既存のコードベースを活かしつつ、同一モジュール内の解析を改善
- 段階的な移行が可能

**長期的**: go/packages直接使用（案1）への移行を検討
- より強力で柔軟な解析が可能
- 独自の最適化やキャッシュ戦略を実装可能

## 実装の優先順位

1. まず同一モジュール内のクロスパッケージ解析を実現（ハイブリッド）
2. 実用性を確認後、必要に応じて完全な独自ドライバーへ移行
3. パフォーマンスが問題になった場合、キャッシュ機構を追加