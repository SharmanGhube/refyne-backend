package user

// Table Names
const (
	userTableName            = "users"
	userPreferencesTableName = "user_preferences"
)

// User CRUD Queries
const (
	insertUserQuery = `
		INSERT INTO users (
			id, email, username, password_hash, status, 
			is_active, is_verified, created_at, updated_at
		) VALUES (
			:id, :email, :username, :password_hash, :status,
			:is_active, :is_verified, :created_at, :updated_at
		)`

	selectUserByEmailQuery = `
		SELECT id, email, username, password_hash, status, is_active, is_verified,
			   last_login_at, last_login_ip, created_at, updated_at, deleted_at
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL`

	selectUserByUsernameQuery = `
		SELECT id, email, username, password_hash, status, is_active, is_verified,
			   last_login_at, last_login_ip, created_at, updated_at, deleted_at
		FROM users 
		WHERE username = $1 AND deleted_at IS NULL`

	checkUserExistsByEmailQuery = `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	checkUserExistsByUsernameQuery = `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`
)
