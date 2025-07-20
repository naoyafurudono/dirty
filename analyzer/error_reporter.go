package analyzer

import (
	"fmt"
	"go/token"
	"strings"
)

// EffectError represents a detailed effect violation error
type EffectError struct {
	CallSite        token.Pos
	Caller          string
	Callee          string
	CallerEffects   []string
	CalleeEffects   []string
	MissingEffects  []string
	PropagationPath []PropagationStep
}

// PropagationStep represents one step in the effect propagation chain
type PropagationStep struct {
	Function string
	Effects  []string
	Source   string // どこからエフェクトが来たか
}

// Format formats the error message with detailed information
func (e *EffectError) Format() string {
	var b strings.Builder

	// メインエラーメッセージ
	b.WriteString(fmt.Sprintf("function calls %s which has effects [%s] not declared in this function\n",
		e.Callee, strings.Join(e.CalleeEffects, ", ")))

	// 詳細情報
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Called function '%s' requires:\n", e.Callee))
	for _, effect := range e.CalleeEffects {
		b.WriteString(fmt.Sprintf("    - %s\n", effect))
	}

	b.WriteString("\n")
	if len(e.CallerEffects) > 0 {
		b.WriteString(fmt.Sprintf("  Function '%s' declares:\n", e.Caller))
		for _, effect := range e.CallerEffects {
			b.WriteString(fmt.Sprintf("    - %s\n", effect))
		}
	} else {
		b.WriteString(fmt.Sprintf("  Function '%s' declares no effects\n", e.Caller))
	}

	b.WriteString("\n")
	b.WriteString("  Missing effects:\n")
	for _, effect := range e.MissingEffects {
		b.WriteString(fmt.Sprintf("    - %s\n", effect))
	}

	// エフェクト伝播経路（暗黙的エフェクトの場合）
	if len(e.PropagationPath) > 0 {
		b.WriteString("\n")
		b.WriteString("  Effect propagation path:\n")
		for i, step := range e.PropagationPath {
			indent := strings.Repeat("  ", i+2)
			if i == 0 {
				b.WriteString(fmt.Sprintf("%s%s\n", indent, step.Function))
			} else {
				b.WriteString(fmt.Sprintf("%s└─ %s (from %s)\n", indent, step.Function, step.Source))
			}
			if len(step.Effects) > 0 {
				b.WriteString(fmt.Sprintf("%s   effects: [%s]\n", indent, strings.Join(step.Effects, ", ")))
			}
		}
	}

	// 修正提案
	b.WriteString("\n")
	b.WriteString("  To fix, add the missing effects to the function declaration:\n")
	allEffects := combineEffects(e.CallerEffects, e.MissingEffects)
	b.WriteString(fmt.Sprintf("    //dirty: %s\n", strings.Join(allEffects, ", ")))

	return b.String()
}

// combineEffects merges two effect slices without duplicates
func combineEffects(existing, missing []string) []string {
	effectSet := make(map[string]bool)
	var result []string

	// Add existing effects
	for _, e := range existing {
		if !effectSet[e] {
			effectSet[e] = true
			result = append(result, e)
		}
	}

	// Add missing effects
	for _, e := range missing {
		if !effectSet[e] {
			effectSet[e] = true
			result = append(result, e)
		}
	}

	return result
}

// BuildPropagationPath builds the effect propagation path for a function
func BuildPropagationPath(funcName string, functions map[string]*FunctionInfo, visited map[string]bool) []PropagationStep {
	if visited[funcName] {
		return nil
	}
	visited[funcName] = true

	fn, ok := functions[funcName]
	if !ok {
		return nil
	}

	var path []PropagationStep

	// 自分自身を追加
	step := PropagationStep{
		Function: funcName,
		Effects:  fn.DeclaredEffects.ToSlice(),
	}

	// 宣言がない場合は計算されたエフェクトを表示
	if !fn.HasDeclaration && len(fn.ComputedEffects) > 0 {
		step.Effects = fn.ComputedEffects.ToSlice()
		step.Source = "computed"
	}

	path = append(path, step)

	// 呼び出し先の関数を追加
	for _, call := range fn.CallSites {
		if _, ok := functions[call.Callee]; ok {
			subPath := BuildPropagationPath(call.Callee, functions, visited)
			if len(subPath) > 0 {
				// sourceを設定
				for i := range subPath {
					if i == 0 {
						subPath[i].Source = funcName
					}
				}
				path = append(path, subPath...)
			}
		}
	}

	return path
}
