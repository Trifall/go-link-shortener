package lib

const ROOT_USER_NAME = "Root User"

type Errors struct {
	Database                string
	NoSecretKey             string
	InvalidSecretKey        string
	InvalidPermissions      string
	NoNewFields             string
	KeyNotFound             string
	KeyRequired             string
	KeyNameRequired         string
	KeyNameAlreadyExists    string
	NewKeyNameTooLong       string
	CannotUpdateRootUserKey string
	FailedKeyCreation       string
}

var ERRORS = Errors{
	Database:                "database not connected",
	NoSecretKey:             "secret key required",
	InvalidSecretKey:        "invalid secret key",
	InvalidPermissions:      "invalid permissions",
	NoNewFields:             "no fields to update",
	KeyNotFound:             "key not found",
	KeyRequired:             "key required",
	KeyNameRequired:         "key name required",
	KeyNameAlreadyExists:    "key name already exists",
	NewKeyNameTooLong:       "new key name is too long",
	CannotUpdateRootUserKey: "cannot update root user key",
	FailedKeyCreation:       "failed to create new key",
}
