package account

// User Settings table name
const (
	userSettingsTableName = "user_settings"
)

// User Settings CRUD Queries
const (
	insertUserSettingsQuery = `
		INSERT INTO user_settings (
			id, user_id, language, timezone, email_notifications, created_at, updated_at
		) VALUES (
			:id, :user_id, :language, :timezone, :email_notifications, :created_at, :updated_at
		)`

	selectUserSettingsByUserIDQuery = `
		SELECT id, user_id, language, timezone, email_notifications, created_at, updated_at
		FROM user_settings 
		WHERE user_id = $1`

	updateUserSettingsQuery = `
		UPDATE user_settings 
		SET language = :language, timezone = :timezone, email_notifications = :email_notifications, 
			updated_at = :updated_at
		WHERE user_id = :user_id`

	deleteUserSettingsQuery = `
		DELETE FROM user_settings WHERE user_id = $1`
)
