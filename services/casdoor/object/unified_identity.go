// Copyright 2024 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/xorm"
)

// User identity binding structure (directly using User table's UniversalId)
type UserIdentityBinding struct {
	Id          string `xorm:"varchar(100) pk" json:"id"`
	UniversalId string `xorm:"varchar(100)" json:"universalId"`
	AuthType    string `xorm:"varchar(50)" json:"authType"`
	AuthValue   string `xorm:"varchar(255)" json:"authValue"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
}

// User merge result
type MergeResult struct {
	UniversalId       string       `json:"universal_id"`
	DeletedUserId     string       `json:"deleted_user_id"`
	MergedAuthMethods []AuthMethod `json:"merged_auth_methods"`
}

// Authentication method
type AuthMethod struct {
	AuthType  string `json:"auth_type"`
	AuthValue string `json:"auth_value"`
}

// User identity binding operations
func AddUserIdentityBinding(binding *UserIdentityBinding) (bool, error) {
	affected, err := ormer.Engine.Insert(binding)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func GetUserIdentityBindingsByUniversalId(universalId string) ([]*UserIdentityBinding, error) {
	bindings := []*UserIdentityBinding{}
	err := ormer.Engine.Where("universal_id = ?", universalId).Find(&bindings)
	return bindings, err
}

func GetUserIdentityBindingByAuth(authType, authValue string) (*UserIdentityBinding, error) {
	binding := &UserIdentityBinding{}
	has, err := ormer.Engine.Where("auth_type = ? AND auth_value = ?", authType, authValue).Get(binding)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return binding, nil
}

func DeleteUserIdentityBinding(id string) (bool, error) {
	affected, err := ormer.Engine.Where("id = ?", id).Delete(&UserIdentityBinding{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteUserIdentityBindingsByUniversalId(universalId string) (bool, error) {
	affected, err := ormer.Engine.Where("universal_id = ?", universalId).Delete(&UserIdentityBinding{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// Check if authentication method exists
func checkAuthMethodExists(session *xorm.Session, universalId, authType, authValue string) (bool, error) {
	count, err := session.Where("universal_id = ? AND auth_type = ? AND auth_value = ?",
		universalId, authType, authValue).Count(&UserIdentityBinding{})
	return count > 0, err
}

// Get user by universal ID
func getUserByUniversalId(universalId string) (*User, error) {
	user := &User{}
	has, err := ormer.Engine.Where("universal_id = ?", universalId).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("user not found for universal_id: %s", universalId)
	}
	return user, nil
}

// Get user's authentication information (phone number and GitHub account)
func getUserAuthInfo(universalId string) (phoneNumber string, githubAccount string, err error) {
	bindings := []*UserIdentityBinding{}
	err = ormer.Engine.Where("universal_id = ?", universalId).Find(&bindings)
	if err != nil {
		return "", "", err
	}

	for _, binding := range bindings {
		switch binding.AuthType {
		case "phone":
			phoneNumber = binding.AuthValue
		case "github":
			githubAccount = binding.AuthValue
		}
	}

	return phoneNumber, githubAccount, nil
}

// Identity binding when creating a user
func createIdentityBindings(session *xorm.Session, user *User, universalId string, primaryProvider string) error {
	return createIdentityBindingsWithValue(session, user, universalId, primaryProvider, "")
}

// Identity binding when creating a user (allow specifying authentication value)
func createIdentityBindingsWithValue(session *xorm.Session, user *User, universalId string, primaryProvider string, providerValue string) error {
	if primaryProvider == "" {
		return fmt.Errorf("primaryProvider is required")
	}

	// If no authentication value is provided, try to get it from the user object
	if providerValue == "" {
		providerValue = getProviderValue(user, primaryProvider)
	}

	// If still empty, try to auto-detect a valid provider type based on user data
	if providerValue == "" {
		autoDetectedProvider, autoDetectedValue := autoDetectProviderType(user)
		if autoDetectedProvider != "" && autoDetectedValue != "" {
			primaryProvider = autoDetectedProvider
			providerValue = autoDetectedValue
		}
	}

	if providerValue == "" {
		return fmt.Errorf("cannot get value for provider type: %s", primaryProvider)
	}

	// Create a unique identity binding record
	binding := &UserIdentityBinding{
		Id:          util.GenerateId(),
		UniversalId: universalId,
		AuthType:    strings.ToLower(primaryProvider),
		AuthValue:   providerValue,
		CreatedTime: util.GetCurrentTime(),
	}

	_, err := session.Insert(binding)
	if err != nil {
		return err
	}

	return nil
}

// Auto-detect provider type based on user data
func autoDetectProviderType(user *User) (string, string) {
	// Priority order: email, phone, username+password
	if user.Email != "" {
		return "email", user.Email
	}
	if user.Phone != "" {
		return "phone", user.Phone
	}
	if user.Password != "" {
		return "password", fmt.Sprintf("%s/%s", user.Owner, user.Name)
	}

	// Check other OAuth providers
	if user.GitHub != "" {
		return "github", user.GitHub
	}
	if user.Google != "" {
		return "google", user.Google
	}
	if user.WeChat != "" {
		return "wechat", user.WeChat
	}
	if user.QQ != "" {
		return "qq", user.QQ
	}
	if user.Facebook != "" {
		return "facebook", user.Facebook
	}
	if user.DingTalk != "" {
		return "dingtalk", user.DingTalk
	}
	if user.Weibo != "" {
		return "weibo", user.Weibo
	}
	if user.Ldap != "" {
		return "ldap", user.Ldap
	}
	if user.Custom != "" {
		return "custom", user.Custom
	}

	return "", ""
}

// Helper function: Get value corresponding to provider type
func getProviderValue(user *User, providerType string) string {
	providerTypeLower := strings.ToLower(providerType)

	switch providerTypeLower {
	case "github":
		if user.GitHub != "" {
			return user.GitHub
		}
		// If user GitHub field is empty but has ID information from GitHub OAuth, get it from Properties
		if user.Properties != nil {
			if githubId := user.Properties["oauth_GitHub_id"]; githubId != "" {
				return githubId
			}
			// Try to get identifier from other GitHub related attributes
			if githubUsername := user.Properties["oauth_GitHub_username"]; githubUsername != "" {
				return githubUsername
			}
		}
		return ""
	case "google":
		return user.Google
	case "wechat":
		return user.WeChat
	case "qq":
		return user.QQ
	case "facebook":
		return user.Facebook
	case "dingtalk":
		return user.DingTalk
	case "weibo":
		return user.Weibo
	case "email":
		return user.Email
	case "phone":
		return user.Phone
	case "password":
		if user.Password != "" {
			return fmt.Sprintf("%s/%s", user.Owner, user.Name)
		}
		return ""
	case "ldap":
		return user.Ldap
	case "custom":
		// First check user's Custom field
		if user.Custom != "" {
			return user.Custom
		}
		// If Custom field is empty, get it from Properties
		if user.Properties != nil {
			if id := user.Properties["oauth_Custom_id"]; id != "" {
				return id
			}
		}
		return ""
	default:
		// For other provider types, try to get it from Properties
		if user.Properties != nil {
			if id := user.Properties[fmt.Sprintf("oauth_%s_id", providerType)]; id != "" {
				return id
			}
		}
		return ""
	}
}

// User merge function
func MergeUsers(reservedUserToken, deletedUserToken string) (*MergeResult, error) {
	// 1. Verify two user tokens
	reservedClaims, err := ParseJwtTokenByApplication(reservedUserToken, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid reserved user token: %v", err)
	}

	deletedClaims, err := ParseJwtTokenByApplication(deletedUserToken, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid deleted user token: %v", err)
	}

	// 2. Check if users exist and get user information
	reservedUser, err := getUserByUniversalId(reservedClaims.UniversalId)
	if err != nil {
		return nil, fmt.Errorf("reserved account does not exist (universalId: %s): %v", reservedClaims.UniversalId, err)
	}
	if reservedUser == nil {
		return nil, fmt.Errorf("reserved account does not exist (universalId: %s)", reservedClaims.UniversalId)
	}

	deletedUser, err := getUserByUniversalId(deletedClaims.UniversalId)
	if err != nil {
		return nil, fmt.Errorf("account to be deleted does not exist (universalId: %s): %v", deletedClaims.UniversalId, err)
	}
	if deletedUser == nil {
		return nil, fmt.Errorf("account to be deleted does not exist (universalId: %s)", deletedClaims.UniversalId)
	}

	// 2.1 Check if users are marked as deleted
	if reservedUser.IsDeleted {
		return nil, fmt.Errorf("reserved account has been deleted and cannot be merged (user: %s)", reservedUser.GetId())
	}
	if deletedUser.IsDeleted {
		return nil, fmt.Errorf("account to be deleted has been deleted and cannot be merged (user: %s)", deletedUser.GetId())
	}

	// 3. Verify merge conditions
	if reservedUser.UniversalId == deletedUser.UniversalId {
		return nil, fmt.Errorf("cannot merge the same user")
	}

	// 4. Get all identity bindings of the user to be deleted
	deletedBindings, err := GetUserIdentityBindingsByUniversalId(deletedUser.UniversalId)
	if err != nil {
		return nil, err
	}

	// 5. Start transaction processing
	session := ormer.Engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	mergedAuthMethods := []AuthMethod{}

	// 6. Handle authentication method transfer
	for _, binding := range deletedBindings {
		// Check if the reserved user already has the same authentication method
		exists, err := checkAuthMethodExists(session, reservedUser.UniversalId, binding.AuthType, binding.AuthValue)
		if err != nil {
			session.Rollback()
			return nil, err
		}

		if !exists {
			// Create new binding record
			newBinding := &UserIdentityBinding{
				Id:          util.GenerateId(),
				UniversalId: reservedUser.UniversalId,
				AuthType:    binding.AuthType,
				AuthValue:   binding.AuthValue,
				CreatedTime: util.GetCurrentTime(),
			}
			_, err = session.Insert(newBinding)
			if err != nil {
				session.Rollback()
				return nil, err
			}

			mergedAuthMethods = append(mergedAuthMethods, AuthMethod{
				AuthType:  binding.AuthType,
				AuthValue: binding.AuthValue,
			})
		}
	}

	// 7. Delete all bindings of the deleted user
	_, err = session.Where("universal_id = ?", deletedUser.UniversalId).Delete(&UserIdentityBinding{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8. Clean up related status of the deleted user
	// 8.1 Delete all tokens of the deleted user
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Token{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.2 Delete all sessions of the deleted user
	deletedUserId := deletedUser.GetId()
	_, err = session.Where("owner = ? AND name = ?", deletedUser.Owner, deletedUser.Name).Delete(&Session{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.3 Delete verification records of the deleted user
	_, err = session.Where("user = ?", deletedUserId).Delete(&VerificationRecord{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.4 Delete resource records of the deleted user
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Resource{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.5 Delete payment records of the deleted user
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Payment{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.6 Delete transaction records of the deleted user
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Transaction{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.7 Delete subscription records of the deleted user
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Subscription{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.8 Clean up operation records of the deleted user (according to business needs, it may be necessary to retain for audit)
	// Note: Record uses casvisorsdk.Record structure, which needs special handling
	// Here we choose to retain records for audit tracking, but can clear or mark User field as deleted
	// _, err = session.Where("user = ?", deletedUserId).Delete(&casvisorsdk.Record{})

	// 9. Delete deleted user record
	_, err = session.Delete(deletedUser)
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 10. Commit transaction
	if err := session.Commit(); err != nil {
		return nil, err
	}

	return &MergeResult{
		UniversalId:       reservedUser.UniversalId,
		DeletedUserId:     deletedUser.UniversalId,
		MergedAuthMethods: mergedAuthMethods,
	}, nil
}

// Login with unified identity
func LoginWithUnifiedIdentity(authType, authValue, password string) (*User, error) {
	var binding *UserIdentityBinding
	var err error

	switch authType {
	case "github":
		binding, err = GetUserIdentityBindingByAuth("github", authValue)
	case "phone":
		binding, err = GetUserIdentityBindingByAuth("phone", authValue)
	case "email":
		binding, err = GetUserIdentityBindingByAuth("email", authValue)
	case "password":
		// User name password login, need to verify password first
		user, err := validateUsernamePassword(authValue, password)
		if err != nil || user == nil {
			return nil, err
		}
		binding, err = GetUserIdentityBindingByAuth("password", fmt.Sprintf("%s/%s", user.Owner, user.Name))
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	if err != nil {
		return nil, err
	}

	if binding == nil {
		return nil, fmt.Errorf("authentication failed")
	}

	// Get user by unified identity ID
	user, err := getUserByUniversalId(binding.UniversalId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Verify username password
func validateUsernamePassword(userOwnerName, password string) (*User, error) {
	// Parse owner/name format
	parts := strings.Split(userOwnerName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid username format, expected: owner/name")
	}

	owner := parts[0]
	name := parts[1]

	// Use existing password verification logic
	user, err := CheckUserPassword(owner, name, password, "en")
	if err != nil {
		return nil, err
	}

	return user, nil
}

// User actively binds additional login methods
func AddUserIdentityBindingForUser(universalId string, authType string, authValue string) (*UserIdentityBinding, error) {
	// Check if it already exists
	existingBinding, err := GetUserIdentityBindingByAuth(authType, authValue)
	if err != nil {
		return nil, err
	}

	if existingBinding != nil {
		if existingBinding.UniversalId == universalId {
			// Already bound to current user, return existing binding
			return existingBinding, nil
		} else {
			// Already bound to other user, not allowed to repeat binding
			return nil, fmt.Errorf("this %s has been bound to other users", authType)
		}
	}

	// Create new identity binding
	binding := &UserIdentityBinding{
		Id:          util.GenerateId(),
		UniversalId: universalId,
		AuthType:    authType,
		AuthValue:   authValue,
		CreatedTime: util.GetCurrentTime(),
	}

	success, err := AddUserIdentityBinding(binding)
	if err != nil {
		return nil, err
	}

	if !success {
		return nil, fmt.Errorf("failed to create identity binding")
	}

	return binding, nil
}

// User removes identity binding
func RemoveUserIdentityBindingForUser(universalId string, authType string) error {
	// Get all identity bindings of the user
	bindings, err := GetUserIdentityBindingsByUniversalId(universalId)
	if err != nil {
		return err
	}

	// Check if there is only one identity binding, if so, not allowed to delete
	if len(bindings) <= 1 {
		return fmt.Errorf("cannot delete the only login method, please bind other login methods first")
	}

	// Find the identity binding to be deleted
	var targetBinding *UserIdentityBinding
	for _, binding := range bindings {
		if binding.AuthType == authType {
			targetBinding = binding
			break
		}
	}

	if targetBinding == nil {
		return fmt.Errorf("identity binding to be deleted not found")
	}

	// Delete identity binding
	success, err := DeleteUserIdentityBinding(targetBinding.Id)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("failed to delete identity binding")
	}

	return nil
}
