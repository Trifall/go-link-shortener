package auth

import (
	"errors"
	"go-link-shortener/database"
	"go-link-shortener/lib"
	"go-link-shortener/models"
)

/*
	All functions in this file require admin permissions
*/

// pass in the creators secret key and the new key name (for the new key)
func GenerateSecretKey(newKeyName string, isAdmin bool) (*models.SecretKey, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New(lib.ERRORS.Database)
	}

	if len(newKeyName) > 100 {
		return nil, errors.New("new key name is too long")
	}

	nameAlreadyExists := models.SearchKeyByName(db, newKeyName)

	if newKeyName == lib.ROOT_USER_NAME || nameAlreadyExists != nil {
		return nil, errors.New("key name already exists")
	}

	// create new key
	key := models.CreateSecretKey(db, newKeyName, isAdmin)

	if key == nil {
		return nil, errors.New("failed to create new key")
	}

	return key, nil
}

type UpdateKeyRequest struct {
	SecretKey         *string `json:"secret_key"`
	SecretKeyToUpdate *string `json:"secret_key_to_update"` // key to update
	Name              *string `json:"name"`
	Active            *bool   `json:"active"`
	IsAdmin           *bool   `json:"is_admin"`
}

func UpdateKey(request UpdateKeyRequest) (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", errors.New(lib.ERRORS.Database)
	}

	if request.SecretKey == nil {
		return "", errors.New(lib.ERRORS.NoSecretKey)
	}

	if request.Name != nil {
		if len(*request.Name) > 100 {
			return "", errors.New("new key name is too long")
		}
		return "", errors.New("key name required")
	}

	if request.SecretKeyToUpdate == nil {
		return "", errors.New("key to update required")
	}

	updateKeyObj := models.SearchKeyByKey(db, *request.SecretKeyToUpdate)

	if updateKeyObj.Name == lib.ROOT_USER_NAME {
		return "", errors.New("cannot update root user key")
	}

	if request.Active != nil {
		updateKeyObj.Active = *request.Active
	}

	if request.IsAdmin != nil {
		updateKeyObj.IsAdmin = *request.IsAdmin
	}

	db.Save(&updateKeyObj)

	return "Key updated successfully", nil
}

func DeleteKeyByKey(keyToDelete string) (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", errors.New(lib.ERRORS.Database)
	}

	if keyToDelete == "" {
		return "", errors.New("key to delete required")
	}

	deleteKeyObj := models.SearchKeyByKey(db, keyToDelete)

	if deleteKeyObj == nil {
		return "", errors.New("key not found")
	}

	if deleteKeyObj.Name == lib.ROOT_USER_NAME {
		return "", errors.New("cannot delete root user key")
	}

	db.Delete(&deleteKeyObj)

	return "Key deleted successfully", nil
}

func DeleteKeyByName(keyName string) (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", errors.New(lib.ERRORS.Database)
	}

	if keyName == "" {
		return "", errors.New("key name required")
	}

	deleteKeyObj := models.SearchKeyByName(db, keyName)

	if deleteKeyObj == nil {
		return "", errors.New("key not found")
	}

	if deleteKeyObj.Name == lib.ROOT_USER_NAME {
		return "", errors.New("cannot delete root user key")
	}

	db.Delete(&deleteKeyObj)

	return "Key deleted successfully", nil
}

func GetKeys(secretKey string) ([]models.SecretKey, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New(lib.ERRORS.Database)
	}

	var keys []models.SecretKey
	db.Find(&keys)

	return keys, nil
}

func ValidateKey(secretKey string) (*models.SecretKey, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New(lib.ERRORS.Database)
	}

	if secretKey == "" {
		return nil, errors.New(lib.ERRORS.NoSecretKey)
	}

	authKeyObj := models.SearchKeyByKey(db, secretKey)

	if authKeyObj == nil {
		return nil, errors.New(lib.ERRORS.InvalidSecretKey)
	}

	return authKeyObj, nil
}
