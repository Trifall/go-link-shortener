package lib

const ROOT_USER_NAME = "Root User"

type Errors struct {
	Database           string
	NoSecretKey        string
	InvalidSecretKey   string
	InvalidPermissions string
}

var ERRORS = Errors{
	Database:           "database not connected",
	NoSecretKey:        "secret key required",
	InvalidSecretKey:   "invalid secret key",
	InvalidPermissions: "invalid permissions",
}
