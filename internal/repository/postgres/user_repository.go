package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
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

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reset the sequence
	_, err = tx.ExecContext(ctx, "SELECT reset_users_id_seq()")
	if err != nil {
		log.Printf("Error resetting sequence: %v", err)
		return err
	}

	query := `INSERT INTO users (name, email, password_hash, date_of_birth, phone_number, is_blocked, is_email_verified, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, query,
		user.Name, user.Email, user.PasswordHash, user.DOB, user.PhoneNumber, user.IsBlocked, user.IsEmailVerified, time.Now()).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation error code
			return usecase.ErrDuplicateEmail
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	// SQL query to select user details by email
	query := `SELECT id, name, email, password_hash, date_of_birth, phone_number, is_blocked, created_at, updated_at, last_login ,is_email_verified
              FROM users WHERE email = $1`

	var user domain.User
	var lastLogin sql.NullTime //nulltime : nullable timestamp value
	//it holds two values : Timestamp and Valid (bool, indicates whether the value is null or not)

	// Execute the query and scan the result into the user struct
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.DOB,
		&user.PhoneNumber, &user.IsBlocked, &user.CreatedAt, &user.UpdatedAt, &lastLogin, &user.IsEmailVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no user is found, return a specific error
			return nil, repository.ErrUserNotFound
		}
		// For any other error, return it as is
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
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
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
			return nil, repository.ErrOTPNotFound
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
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), userID)
	return err
}

func (r *userRepository) CreateVerificationEntry(ctx context.Context, entry *domain.VerificationEntry) error {

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
		time.Now()).Scan(&entry.ID)

	return err
}

func (r *userRepository) GetVerificationEntryByEmail(ctx context.Context, email string) (*domain.VerificationEntry, error) {
	query := `SELECT id, email, otp_code, user_data, password_hash, expires_at, is_verified, created_at
              FROM verification_entries
              WHERE email = $1 AND is_verified = false
              ORDER BY created_at DESC
              LIMIT 1`

	var entry domain.VerificationEntry
	var userDataJSON []byte
	var expiresAt time.Time //later used to verify the time is in UTC
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&entry.ID, &entry.Email, &entry.OTPCode, &userDataJSON, &entry.PasswordHash, &expiresAt, &entry.IsVerified, &entry.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrVerificationEntryNotFound
		}
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

func (r *userRepository) UpdateVerificationEntry(ctx context.Context, entry *domain.VerificationEntry) error {
	query := `UPDATE verification_entries
              SET is_verified = $1
              WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, entry.IsVerified, entry.ID)
	return err
}

func (r *userRepository) DeleteExpiredVerificationEntries(ctx context.Context) error {
	query := `DELETE FROM verification_entries
              WHERE expires_at < $1 AND is_verified = false`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}

func (r *userRepository) DeleteVerificationEntry(ctx context.Context, email string) error {
	query := `DELETE FROM verification_entries WHERE email = $1`
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

func (r *userRepository) UpdateVerificationEntryAfterResendOTP(ctx context.Context, entry *domain.VerificationEntry) error {
	query := `UPDATE verification_entries
              SET is_verified = $1, otp_code = $2, expires_at = $3
              WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, entry.IsVerified, entry.OTPCode, entry.ExpiresAt, entry.ID)
	if err != nil {
		log.Printf("Error updating verification entry: %v", err)
	}
	return err
}
