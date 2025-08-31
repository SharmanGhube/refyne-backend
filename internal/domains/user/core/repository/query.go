package user

// Table names
const (
	userTableName         = "users"
	userSettingsTableName = "user_settings"
)

// CRUD Queries

const (
	insertUserQuery = `
		INSERT INTO users (
			id,
			email, 
			password_hash, 
			first_name, 
			last_name, 
			username, 
			status, 
			is_active, 
			is_verified,
			created_at,
			updated_at
		) VALUES (
			:id, :email, :password_hash, :first_name, :last_name, :username, :status, :is_active, :is_verified, :created_at, :updated_at
		)`

	selectUserByIDQuery = `
		SELECT id, email, password_hash, first_name, last_name, username,
			   status, is_active, is_verified, last_login, last_login_ip,
			   created_at, updated_at, deleted_at
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`

	selectUserByEmailQuery = `
		SELECT id, email, password_hash, first_name, last_name, username,
			   status, is_active, is_verified, last_login, last_login_ip,
			   created_at, updated_at, deleted_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`

	selectUserByUsernameQuery = `
		SELECT id, email, password_hash, first_name, last_name, username,
			   status, is_active, is_verified, last_login, last_login_ip,
			   created_at, updated_at, deleted_at
		FROM users 
		WHERE username = $1 AND deleted_at IS NULL
	`

	updateUserQuery = `
		UPDATE users 
		SET first_name = $2, 
			last_name = $3, 
			username = $4,
			status = $5,
			is_active = $6,
			is_verified = $7,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, first_name, last_name, username,
				  status, is_active, is_verified, last_login, last_login_ip,
				  created_at, updated_at, deleted_at
	`

	updateUserPasswordQuery = `
		UPDATE users 
		SET password_hash = $2,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, first_name, last_name, username,
				  status, is_active, is_verified, last_login, last_login_ip,
				  created_at, updated_at, deleted_at
	`

	updateUserLoginInfoQuery = `
		UPDATE users 
		SET last_login = $2,
			last_login_ip = $3,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	updateUserStatusQuery = `
		UPDATE users 
		SET status = $2,
			is_active = $3,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, first_name, last_name, username,
				  status, is_active, is_verified, last_login, last_login_ip,
				  created_at, updated_at, deleted_at
	`

	updateUserVerificationQuery = `
		UPDATE users 
		SET is_verified = $2,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, email, password_hash, first_name, last_name, username,
				  status, is_active, is_verified, last_login, last_login_ip,
				  created_at, updated_at, deleted_at
	`

	softDeleteUserQuery = `
		UPDATE users 
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	hardDeleteUserQuery = `
		DELETE FROM users 
		WHERE id = $1
	`

	listUsersQuery = `
		SELECT id, email, password_hash, first_name, last_name, username,
			   status, is_active, is_verified, last_login, last_login_ip,
			   created_at, updated_at, deleted_at
		FROM users 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	countUsersQuery = `
		SELECT COUNT(*) 
		FROM users 
		WHERE deleted_at IS NULL
	`

	verifyUserQuery = `
		UPDATE users 
		SET is_verified = true, 
			is_active = true, 
			status = 'active',
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
)
