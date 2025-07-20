package sqlc

import (
	"context"
	"fmt"
)

// Example 1: 正しい例 - 必要なエフェクトをすべて宣言
// dirty: select[users], insert[audit_logs]
func GetUserWithAuditLog(ctx context.Context, q *Queries, userID int64) (*User, error) {
	// GetUserはselect[users]のエフェクトを持つ（JSONから自動検出）
	user, err := q.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// LogUserAccessはinsert[audit_logs]のエフェクトを持つ
	if err := LogUserAccess(ctx, userID, "view_profile"); err != nil {
		return nil, fmt.Errorf("failed to log access: %w", err)
	}

	return &user, nil
}

// Example 2: エラーの例 - select[users]エフェクトが宣言されていない
// dirty: insert[audit_logs]
func BrokenGetUser(ctx context.Context, q *Queries, userID int64) (*User, error) {
	// エラー: GetUserのselect[users]エフェクトが宣言されていない
	user, err := q.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := LogUserAccess(ctx, userID, "view_profile"); err != nil {
		return nil, err
	}

	return &user, nil
}

// Example 3: 複数テーブルへのエフェクト
// dirty: select[users], select[organizations], insert[audit_logs]
func GetUserFullProfile(ctx context.Context, q *Queries, userID int64) (*GetUserWithOrganizationRow, error) {
	// GetUserWithOrganizationはselect[users]とselect[organizations]の両方のエフェクトを持つ
	profile, err := q.GetUserWithOrganization(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// ログ記録
	if err := LogUserAccess(ctx, userID, "view_full_profile"); err != nil {
		return nil, fmt.Errorf("failed to log access: %w", err)
	}

	return &profile, nil
}

// Example 4: トランザクションでの複数エフェクト
// dirty: insert[users], insert[audit_logs], insert[notifications]
func CreateUserWithNotification(ctx context.Context, q *Queries, email, name string) error {
	// CreateUserWithAuditLogはinsert[users]とinsert[audit_logs]のエフェクトを持つ
	err := q.CreateUserWithAuditLog(ctx, CreateUserWithAuditLogParams{
		Email:  email,
		Name:   name,
		Action: "user_created",
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 通知を送信（手動でエフェクトを宣言）
	if err := SendWelcomeNotification(ctx, email); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// Example 5: 削除と更新の組み合わせ
// dirty: update[users], delete[sessions], insert[audit_logs]
func ArchiveUserAccount(ctx context.Context, q *Queries, userID int64) error {
	// ArchiveUserAndSessionsはupdate[users]とdelete[sessions]のエフェクトを持つ
	if err := q.ArchiveUserAndSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to archive user: %w", err)
	}

	// アーカイブ操作をログに記録
	if err := LogUserAccess(ctx, userID, "account_archived"); err != nil {
		return fmt.Errorf("failed to log archive action: %w", err)
	}

	return nil
}

// Example 6: エフェクトなしの関数（暗黙的にエフェクトが計算される）
func ImplicitEffectsExample(ctx context.Context, q *Queries, email string) error {
	// この関数には// dirty:宣言がないため、呼び出す関数のエフェクトが暗黙的に計算される
	// GetUserByEmailのselect[users]とSendPasswordResetのinsert[notifications]が検出される
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return SendPasswordReset(ctx, user.Email)
}

// Helper functions with manual effect declarations

// dirty: insert[audit_logs]
func LogUserAccess(ctx context.Context, userID int64, action string) error {
	// INSERT INTO audit_logs (user_id, action, created_at) VALUES (?, ?, NOW())
	fmt.Printf("Logging access: user=%d, action=%s\n", userID, action)
	return nil
}

// dirty: insert[notifications]
func SendWelcomeNotification(ctx context.Context, email string) error {
	// INSERT INTO notifications (email, type, sent_at) VALUES (?, 'welcome', NOW())
	fmt.Printf("Sending welcome email to: %s\n", email)
	return nil
}

// dirty: insert[notifications]
func SendPasswordReset(ctx context.Context, email string) error {
	// INSERT INTO notifications (email, type, sent_at) VALUES (?, 'password_reset', NOW())
	fmt.Printf("Sending password reset to: %s\n", email)
	return nil
}
