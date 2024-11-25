package database

import "time"

// GitHubProfile model
type GitHubProfile struct {
	ID           int64     `pg:"id,pk"`
	UserID       int64     `pg:"user_id,unique"`
	GitHubID     int64     `pg:"github_id,unique"`
	AccessToken  string    `pg:"access_token"`
	RefreshToken string    `pg:"refresh_token"`
	Username     string    `pg:"username"`
	ProfileURL   string    `pg:"profile_url"`
	AvatarURL    string    `pg:"avatar_url"`
	TokenExpiry  time.Time `pg:"token_expiry"`
	LinkedAt     time.Time `pg:"linked_at"`
}
