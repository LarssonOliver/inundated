package model

type User struct {
	ID          int    // Primary key
	Name        string // Unique user name, from OIDC
	DisplayName string // Display name of the user, fetch this from OIDC
}
