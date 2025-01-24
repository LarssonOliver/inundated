package model

type User struct {
	Name        string // Unique user name, likely comes from OIDC provider
	DisplayName string // Display name of the user
}
