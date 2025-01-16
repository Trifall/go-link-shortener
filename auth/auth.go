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
		return nil, errors.New(lib.ERRORS.NewKeyNameTooLong)
	}

	nameAlreadyExists := models.SearchKeyByName(db, newKeyName)

	if newKeyName == lib.ROOT_USER_NAME || nameAlreadyExists != nil {
		return nil, errors.New(lib.ERRORS.KeyNameAlreadyExists)
	}

	// create new key
	key := models.CreateSecretKey(db, newKeyName, isAdmin)

	if key == nil {
		return nil, errors.New(lib.ERRORS.FailedKeyCreation)
	}

	return key, nil
}

type UpdateKeyS struct {
	Name     *string
	Key      *string
	IsActive *bool
	IsAdmin  *bool
}

func UpdateKey(request UpdateKeyS) (string, *models.SecretKey, error) {
	db := database.GetDB()
	if db == nil {
		return "", nil, errors.New(lib.ERRORS.Database)
	}

	if request.Key == nil {
		return "", nil, errors.New(lib.ERRORS.KeyRequired)
	}

	updateKeyObj := models.SearchKeyByKey(db, *request.Key)
	if updateKeyObj == nil {
		return "", nil, errors.New(lib.ERRORS.KeyNotFound)
	}
	if updateKeyObj.Name == lib.ROOT_USER_NAME {
		return "", nil, errors.New(lib.ERRORS.CannotUpdateRootUserKey)
	}

	if request.Name != nil {
		if len(*request.Name) > 100 {
			return "", nil, errors.New(lib.ERRORS.NewKeyNameTooLong)
		}
		updateKeyObj.Name = *request.Name
	}

	if request.IsActive != nil {
		updateKeyObj.IsActive = *request.IsActive
	}

	if request.IsAdmin != nil {
		updateKeyObj.IsAdmin = *request.IsAdmin
	}

	// if all of the fields except for the key are nil, return an error with message "no fields to update"
	if request.Name == nil && request.IsActive == nil && request.IsAdmin == nil {
		return "", nil, errors.New(lib.ERRORS.NoNewFields)
	}

	// TODO: could update this to return custom error
	if err := db.Save(&updateKeyObj).Error; err != nil {
		return "", nil, err
	}

	return "Key updated successfully", updateKeyObj, nil
}

func DeleteKeyByKey(keyToDelete string) (string, error) {
	db := database.GetDB()
	if db == nil {
		return "", errors.New(lib.ERRORS.Database)
	}

	if keyToDelete == "" {
		return "", errors.New(lib.ERRORS.KeyRequired)
	}

	deleteKeyObj := models.SearchKeyByKey(db, keyToDelete)

	if deleteKeyObj == nil {
		return "", errors.New(lib.ERRORS.KeyNotFound)
	}

	if deleteKeyObj.Name == lib.ROOT_USER_NAME {
		return "", errors.New(lib.ERRORS.CannotUpdateRootUserKey)
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
		return "", errors.New(lib.ERRORS.KeyNameRequired)
	}

	deleteKeyObj := models.SearchKeyByName(db, keyName)

	if deleteKeyObj == nil {
		return "", errors.New(lib.ERRORS.KeyNotFound)
	}

	if deleteKeyObj.Name == lib.ROOT_USER_NAME {
		return "", errors.New(lib.ERRORS.CannotUpdateRootUserKey)
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
