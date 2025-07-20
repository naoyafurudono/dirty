package example

// Base effects
//dirty: select[config]
func LoadConfig() error {
	return nil
}

//dirty: select[user]
func GetUserData(id int64) error {
	return nil
}

//dirty: insert[audit]
func WriteAuditLog(msg string) error {
	return nil
}

// チェーン1: 設定を読み込んでユーザーを取得
func LoadUserWithConfig(id int64) error {
	if err := LoadConfig(); err != nil {
		return err
	}
	return GetUserData(id)
}

// チェーン2: ユーザーを取得して監査ログを書く
func GetUserWithAudit(id int64) error {
	if err := GetUserData(id); err != nil {
		return err
	}
	return WriteAuditLog("user accessed")
}

// チェーン3: すべてを組み合わせる
func ComplexOperation(id int64) error {
	if err := LoadUserWithConfig(id); err != nil {
		return err
	}
	return GetUserWithAudit(id)
}

// エラー: 深い呼び出しチェーンのエフェクトが不足
//dirty: insert[audit]
func BrokenDeepChain(id int64) error {
	// ComplexOperationは以下のエフェクトを持つ:
	// - select[config] (LoadConfig経由)
	// - select[user] (GetUserData経由、2回)
	// - insert[audit] (WriteAuditLog経由)
	return ComplexOperation(id)
}
