package example

// データベースアクセスを行う関数
// dirty: select[user]
func GetUserByID(id int64) error {
	// SELECT * FROM users WHERE id = ?
	return nil
}

// ログを記録する関数
// dirty: insert[log]
func WriteLog(message string) error {
	// INSERT INTO logs (message) VALUES (?)
	return nil
}

// 正しい例: 必要なエフェクトをすべて宣言
// dirty: select[user], insert[log]
func ProcessUser(id int64) error {
	if err := GetUserByID(id); err != nil {
		return err
	}
	return WriteLog("user processed")
}

// エラーになる例: select[user]エフェクトが不足
// dirty: insert[log]
func ProcessUserBroken(id int64) error {
	if err := GetUserByID(id); err != nil {
		return err
	}
	return WriteLog("user processed")
}

// 宣言なし関数（チェック対象外だが、暗黙的にエフェクトを持つ）
func HelperFunction(id int64) error {
	return GetUserByID(id)
}

// エラーになる例: HelperFunctionの暗黙的エフェクトが不足
// dirty: insert[log]
func UseHelper(id int64) error {
	if err := HelperFunction(id); err != nil {
		return err
	}
	return WriteLog("done")
}
