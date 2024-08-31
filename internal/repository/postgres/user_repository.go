package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

// this code sets up a structure that will handle database operations related to users.
type userRepository struct {
	db *sql.DB // pointer to sql db object, represents a connection to the database.
}

// constructor function for creating new userRepository instances.
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db} //creates a new userRepository with the provided database connection and returns it.
}

//This approach follows the dependency injection principle, where the database connection is provided from outside rather than created within the repository. This makes the code more flexible and easier to test.

func (r *userRepository) CreateUserSignUpVerifcationEntry(ctx context.Context, entry *domain.VerificationEntry) error {

	query := `INSERT INTO verification_entries (email, otp_code, user_data, password_hash, expires_at, is_verified, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7)
              RETURNING id`

	userDataJSON, err := json.Marshal(entry.UserData)
	if err != nil {
		log.Printf("Error marshaling user data: %v", err)
		return err
	}

	err = r.db.QueryRowContext(ctx, query,
		entry.Email,
		entry.OTPCode,
		userDataJSON,
		entry.PasswordHash,
		entry.ExpiresAt,
		entry.IsVerified,
		time.Now().UTC()).Scan(&entry.ID)

	if err != nil {
		log.Printf("Error creating signup verification entry: %v", err)
		return err
	}
	return nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email, password_hash, date_of_birth, phone_number, is_blocked, is_email_verified, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		user.Name, user.Email, user.PasswordHash, user.DOB,
		user.PhoneNumber, user.IsBlocked, user.IsEmailVerified, time.Now().UTC()).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation error code
			return utils.ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	// SQL query to select user details by email
	query := `SELECT id, name, email, password_hash, date_of_birth, phone_number, is_blocked, created_at, updated_at, last_login ,is_email_verified
              FROM users 
			  WHERE email = $1 AND deleted_at IS NULL`

	var user domain.User
	var lastLogin sql.NullTime //nulltime : nullable timestamp value. it holds two values : Timestamp and Valid (bool, indicates whether the value is null or not)

	// Execute the query and scan the result into the user struct
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.DOB,
		&user.PhoneNumber, &user.IsBlocked, &user.CreatedAt, &user.UpdatedAt,
		&lastLogin, &user.IsEmailVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no user is found, return a specific error
			return nil, utils.ErrUserNotFound
		}
		// For any other error, return it as is
		log.Printf("error while retrieving user by email : %v", err)
		return nil, err
	}

	// If lastLogin is not null, that means there is valid time stored in the lastLogin.Time
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	// Return the user struct and nil error
	return &user, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	// SQL query to update the last_login field for a user
	query := `UPDATE users SET last_login = $1 WHERE id = $2`

	// Execute the update query
	_, err := r.db.ExecContext(ctx, query, time.Now().UTC(), userID)
	return err
}

func (r *userRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	// SQL query to insert the blacklisted token
	query := `INSERT INTO blacklisted_tokens (token, expires_at) VALUES ($1, $2)`
	// Execute the insert query
	_, err := r.db.ExecContext(ctx, query, token, expiresAt)
	return err
}

func (r *userRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, token).Scan(&exists)
	return exists, err
}

func (r *userRepository) CreateOTP(ctx context.Context, otp *domain.OTP) error {
	query := `INSERT INTO user_otps (user_id, email, otp, expires_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := r.db.QueryRowContext(ctx, query, otp.UserID, otp.Email, otp.OTPCode, otp.ExpiresAt).Scan(&otp.ID, &otp.CreatedAt)
	return err
}

func (r *userRepository) GetOTPByEmail(ctx context.Context, email string) (*domain.OTP, error) {
	query := `SELECT id, user_id, email, otp, expires_at, created_at FROM user_otps WHERE email = $1`
	var otp domain.OTP
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&otp.ID, &otp.UserID, &otp.Email, &otp.OTPCode, &otp.ExpiresAt, &otp.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrOTPNotFound
		}
		return nil, err
	}
	return &otp, nil
}

func (r *userRepository) DeleteOTP(ctx context.Context, email string) error {
	query := `DELETE FROM user_otps WHERE email = $1`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

func (r *userRepository) UpdateEmailVerificationStatus(ctx context.Context, userID int64, status bool) error {
	query := `UPDATE users SET is_email_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now().UTC(), userID)
	return err
}

func (r *userRepository) UpdateSignUpVerificationEntry(ctx context.Context, entry *domain.VerificationEntry) error {
	query := `UPDATE verification_entries
              SET is_verified = $1
              WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, entry.IsVerified, entry.ID)
	if err != nil {
		log.Printf("error while updating the verification entry table after verifying otp : %v", err)
	}
	return err
}

func (r *userRepository) DeleteExpiredVerificationEntries(ctx context.Context) error {
	query := `DELETE FROM verification_entries
              WHERE expires_at < $1 AND is_verified = false`

	_, err := r.db.ExecContext(ctx, query, time.Now().UTC())
	return err
}

func (r *userRepository) DeleteSignUpVerificationEntry(ctx context.Context, email string) error {
	query := `DELETE FROM verification_entries WHERE email = $1`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

func (r *userRepository) DeletePasswordResetVerificationEntry(ctx context.Context, email string) error {
	query := `DELETE FROM password_reset_entries WHERE email = $1`
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

func (r *userRepository) GetOTPResendInfo(ctx context.Context, email string) (int, time.Time, error) {
	query := `SELECT resend_count, last_resend_time FROM otp_resend_info WHERE email = $1`
	var count int
	var lastResendTime time.Time
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count, &lastResendTime)
	if err == sql.ErrNoRows {
		return 0, time.Time{}, nil
	}
	return count, lastResendTime, err
}

func (r *userRepository) UpdateOTPResendInfo(ctx context.Context, email string) error {
	query := `
        INSERT INTO otp_resend_info (email, resend_count, last_resend_time)
        VALUES ($1, 1, NOW())
        ON CONFLICT (email) DO UPDATE
        SET resend_count = otp_resend_info.resend_count + 1, last_resend_time = NOW()
    `
	_, err := r.db.ExecContext(ctx, query, email)
	return err
}

func (r *userRepository) UpdateSignUpVerificationEntryAfterResendOTP(ctx context.Context, entry *domain.VerificationEntry) error {
	query := `UPDATE verification_entries
              SET is_verified = $1, otp_code = $2, expires_at = $3
              WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, entry.IsVerified, entry.OTPCode, entry.ExpiresAt, entry.ID)
	if err != nil {
		log.Printf("Error updating verification entry: %v", err)
	}
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, name, email, password_hash, date_of_birth, phone_number, is_blocked, is_email_verified, created_at, updated_at, last_login
              FROM users 
              WHERE id = $1 AND deleted_at IS NULL`

	var user domain.User
	var lastLogin sql.NullTime //time.Time can't hold null values

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.DOB,
		&user.PhoneNumber, &user.IsBlocked, &user.IsEmailVerified,
		&user.CreatedAt, &user.UpdatedAt, &lastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrUserNotFound
		}
		log.Printf("failed to retrieve the user by id  : %v ", err)
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users 	SET name=$1, phone_number=$2,updated_at= NOW() WHERE id=$3`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.PhoneNumber, user.ID)
	if err != nil {
		log.Printf("Failed to update the user data: %v", err)
		return err
	}
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID int64, newPasswordHash string) error {
	query := `UPDATE users SET password_hash= $1, updated_at= NOW() WHERE id= $2`
	_, err := r.db.ExecContext(ctx, query, newPasswordHash, userID)
	if err != nil {
		log.Printf("error while updating the password_hash : %v", err)
	}
	return err
}

func (r *userRepository) CreatePasswordResetEntry(ctx context.Context, entry *domain.PasswordResetEntry) error {
	query := `INSERT INTO password_reset_entries (email, otp_code, expires_at, is_verified, created_at)
              VALUES ($1, $2, $3, $4, $5)
              RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		entry.Email,
		entry.OTPCode,
		entry.ExpiresAt,
		entry.IsVerified,
		time.Now().UTC()).Scan(&entry.ID)

	if err != nil {
		log.Printf("Error creating password reset entry: %v", err)
		return err
	}
	return nil
}

func (r *userRepository) FindSignUpVerificationEntryByEmail(ctx context.Context, email string) (*domain.VerificationEntry, error) {
	query := `SELECT id, email, otp_code, user_data, password_hash, expires_at, is_verified, created_at
              FROM verification_entries
              WHERE email = $1 AND is_verified = false
              ORDER BY created_at DESC
              LIMIT 1`

	var entry domain.VerificationEntry
	var userDataJSON []byte
	var expiresAt time.Time //later used to verify the time is in UTC
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&entry.ID,
		&entry.Email,
		&entry.OTPCode,
		&userDataJSON,
		&entry.PasswordHash,
		&expiresAt,
		&entry.IsVerified,
		&entry.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrVerificationEntryNotFound
		}
		log.Printf("error while retrieving verification entries : %v", err)
		return nil, err
	}

	// Ensure the time is in UTC
	entry.ExpiresAt = expiresAt.UTC()

	err = json.Unmarshal(userDataJSON, &entry.UserData)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *userRepository) FindPasswordResetEntryByEmail(ctx context.Context, email string) (*domain.PasswordResetEntry, error) {
	query := `SELECT id, email, otp_code , expires_at , is_verified, created_at
				FROM password_reset_entries 
				WHERE email=$1 AND  is_verified=false
				ORDER BY created_at DESC
				LIMIT 1`

	var entry domain.PasswordResetEntry
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&entry.ID,
		&entry.Email,
		&entry.OTPCode,
		&expiresAt,
		&entry.IsVerified,
		&entry.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrVerificationEntryNotFound
		}
		log.Printf("error while retrieving verification entries : %v", err)
		return nil, err
	}

	// Ensure the time is in UTC
	entry.ExpiresAt = expiresAt.UTC()

	return &entry, nil
}

func (r *userRepository) UserAddressExists(ctx context.Context, address *domain.UserAddress) (bool, error) {
	query := `SELECT EXISTS (
	SELECT 1 FROM user_address 
	WHERE user_id =$1 AND address_line1 = $2 AND city=$3 AND state=$4 AND pincode=$5 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, address.UserID,
		address.AddressLine1,
		address.City,
		address.State,
		address.PinCode).Scan(&exists)
	if err != nil {
		log.Printf("error while checking user exists : %v", err)
	}
	return exists, err
}

func (r *userRepository) AddUserAddress(ctx context.Context, address *domain.UserAddress) error {
	query := `
				INSERT INTO user_address (user_id, address_line1, address_line2, state, city, pincode, landmark, phone_number)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id,created_at, updated_at
				`
	err := r.db.QueryRowContext(ctx, query, address.UserID, address.AddressLine1, address.AddressLine2,
		address.State, address.City, address.PinCode, address.Landmark,
		address.PhoneNumber).Scan(&address.ID, &address.CreatedAt, &address.UpdatedAt)

	if err != nil {
		log.Printf("error while adding address data into user_address table : %v", err)
	}
	return err
}

func (r *userRepository) GetUserAddressByID(ctx context.Context, addressID int64) (*domain.UserAddress, error) {
	query := `
		SELECT id, user_id, address_line1, address_line2, state, city, pincode, landmark, phone_number, created_at, updated_at
		FROM user_address
		WHERE id=$1 AND deleted_at IS NULL		
		`
	var address domain.UserAddress
	err := r.db.QueryRowContext(ctx, query, addressID).Scan(
		&address.ID, &address.UserID, &address.AddressLine1,
		&address.AddressLine2, &address.State, &address.City,
		&address.PinCode, &address.Landmark, &address.PhoneNumber,
		&address.CreatedAt, &address.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrAddressNotFound
		}
		return nil, err
	}

	return &address, nil
}

func (r *userRepository) UpdateUserAddress(ctx context.Context, address *domain.UserAddress) error {
	query := `
			UPDATE user_address SET
			address_line1= $1, address_line2 = $2, city = $3, state = $4, 
			pincode = $5, landmark = $6, phone_number = $7, updated_at = $8
			WHERE id= $9 AND user_id= $10 AND deleted_at IS NULL
			  `
	_, err := r.db.ExecContext(ctx, query,
		address.AddressLine1, address.AddressLine2, address.City,
		address.State, address.PinCode, address.Landmark, address.PhoneNumber,
		time.Now().UTC(), address.ID, address.UserID)
	if err != nil {
		log.Printf("error while updating user address : %v", err)
	}
	return err
}

func (r *userRepository) GetUserAddresses(ctx context.Context, userID int64) ([]*domain.UserAddress, error) {
	query := `
        SELECT id, user_id, address_line1, address_line2, state, city, pincode, landmark, phone_number, created_at, updated_at
        FROM user_address
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []*domain.UserAddress
	for rows.Next() {
		var addr domain.UserAddress
		err := rows.Scan(
			&addr.ID, &addr.UserID, &addr.AddressLine1, &addr.AddressLine2, &addr.State,
			&addr.City, &addr.PinCode, &addr.Landmark, &addr.PhoneNumber,
			&addr.CreatedAt, &addr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, &addr)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func (r *userRepository) GetUserAddressCount(ctx context.Context, userID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_address
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepository) DeleteUserAddress(ctx context.Context, addressID int64) error {
	query := `
		UPDATE user_address
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, addressID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrAddressNotFound
	}

	return nil
}
