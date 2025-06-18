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

// 用户身份绑定结构（直接使用 User 表的 UniversalId）
type UserIdentityBinding struct {
	Id          string `xorm:"varchar(100) pk" json:"id"`
	UniversalId string `xorm:"varchar(100)" json:"universalId"`
	AuthType    string `xorm:"varchar(50)" json:"authType"`
	AuthValue   string `xorm:"varchar(255)" json:"authValue"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
}

// 用户合并结果
type MergeResult struct {
	UniversalId       string       `json:"universal_id"`
	DeletedUserId     string       `json:"deleted_user_id"`
	MergedAuthMethods []AuthMethod `json:"merged_auth_methods"`
}

// 认证方式
type AuthMethod struct {
	AuthType  string `json:"auth_type"`
	AuthValue string `json:"auth_value"`
}

// 用户身份绑定操作
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

// 检查认证方式是否存在
func checkAuthMethodExists(session *xorm.Session, universalId, authType, authValue string) (bool, error) {
	count, err := session.Where("universal_id = ? AND auth_type = ? AND auth_value = ?",
		universalId, authType, authValue).Count(&UserIdentityBinding{})
	return count > 0, err
}

// 通过统一身份ID获取用户
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

// 获取用户的认证信息（手机号和GitHub账号）
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

// 创建用户时的身份绑定
func createIdentityBindings(session *xorm.Session, user *User, universalId string, primaryProvider string) error {
	return createIdentityBindingsWithValue(session, user, universalId, primaryProvider, "")
}

// 创建用户时的身份绑定（允许指定认证值）
func createIdentityBindingsWithValue(session *xorm.Session, user *User, universalId string, primaryProvider string, providerValue string) error {
	if primaryProvider == "" {
		return fmt.Errorf("primaryProvider is required")
	}

	// 如果没有提供认证值，则尝试从用户对象获取
	if providerValue == "" {
		providerValue = getProviderValue(user, primaryProvider)
	}

	if providerValue == "" {
		return fmt.Errorf("cannot get value for provider type: %s", primaryProvider)
	}

	// 创建唯一的身份绑定记录
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

// 辅助函数：根据provider类型获取对应的值
func getProviderValue(user *User, providerType string) string {
	providerTypeLower := strings.ToLower(providerType)

	switch providerTypeLower {
	case "github":
		if user.GitHub != "" {
			return user.GitHub
		}
		// 如果用户GitHub字段为空，但有从GitHub OAuth获取的ID信息，从Properties中获取
		if user.Properties != nil {
			if githubId := user.Properties["oauth_GitHub_id"]; githubId != "" {
				return githubId
			}
			// 尝试从其他GitHub相关属性获取标识符
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
		// 首先检查用户的Custom字段
		if user.Custom != "" {
			return user.Custom
		}
		// 如果Custom字段为空，从Properties中获取
		if user.Properties != nil {
			if id := user.Properties["oauth_Custom_id"]; id != "" {
				return id
			}
		}
		return ""
	default:
		// 对于其他provider类型，尝试从Properties中获取
		if user.Properties != nil {
			if id := user.Properties[fmt.Sprintf("oauth_%s_id", providerType)]; id != "" {
				return id
			}
		}
		return ""
	}
}

// 用户合并函数
func MergeUsers(reservedUserToken, deletedUserToken string) (*MergeResult, error) {
	// 1. 验证两个用户 token
	reservedClaims, err := ParseJwtTokenByApplication(reservedUserToken, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid reserved user token: %v", err)
	}

	deletedClaims, err := ParseJwtTokenByApplication(deletedUserToken, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid deleted user token: %v", err)
	}

	// 2. 获取用户信息
	reservedUser, err := getUserByUniversalId(reservedClaims.UniversalId)
	if err != nil {
		return nil, err
	}

	deletedUser, err := getUserByUniversalId(deletedClaims.UniversalId)
	if err != nil {
		return nil, err
	}

	// 3. 验证合并条件
	if reservedUser.UniversalId == deletedUser.UniversalId {
		return nil, fmt.Errorf("cannot merge the same user")
	}

	// 4. 获取要删除用户的所有身份绑定
	deletedBindings, err := GetUserIdentityBindingsByUniversalId(deletedUser.UniversalId)
	if err != nil {
		return nil, err
	}

	// 5. 开始事务处理
	session := ormer.Engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	mergedAuthMethods := []AuthMethod{}

	// 6. 处理认证方式转移
	for _, binding := range deletedBindings {
		// 检查保留用户是否已有相同的认证方式
		exists, err := checkAuthMethodExists(session, reservedUser.UniversalId, binding.AuthType, binding.AuthValue)
		if err != nil {
			session.Rollback()
			return nil, err
		}

		if !exists {
			// 创建新的绑定记录
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

	// 7. 删除被删除用户的所有绑定记录
	_, err = session.Where("universal_id = ?", deletedUser.UniversalId).Delete(&UserIdentityBinding{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8. 清理被删除用户的相关状态
	// 8.1 删除被删除用户的所有 token
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Token{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.2 删除被删除用户的所有 session
	deletedUserId := deletedUser.GetId()
	_, err = session.Where("owner = ? AND name = ?", deletedUser.Owner, deletedUser.Name).Delete(&Session{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.3 删除被删除用户的验证记录
	_, err = session.Where("user = ?", deletedUserId).Delete(&VerificationRecord{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.4 删除被删除用户的资源记录
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Resource{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.5 删除被删除用户的支付记录
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Payment{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.6 删除被删除用户的交易记录
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Transaction{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.7 删除被删除用户的订阅记录
	_, err = session.Where("user = ?", deletedUser.Name).Delete(&Subscription{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8.8 清理被删除用户的操作记录（根据业务需求，可能需要保留用于审计）
	// 注意：Record 使用的是 casvisorsdk.Record 结构，需要特殊处理
	// 这里我们选择保留记录用于审计追踪，但可以将 User 字段清空或标记为已删除
	// _, err = session.Where("user = ?", deletedUserId).Delete(&casvisorsdk.Record{})

	// 9. 删除被删除用户记录
	_, err = session.Delete(deletedUser)
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 10. 提交事务
	if err := session.Commit(); err != nil {
		return nil, err
	}

	return &MergeResult{
		UniversalId:       reservedUser.UniversalId,
		DeletedUserId:     deletedUser.UniversalId,
		MergedAuthMethods: mergedAuthMethods,
	}, nil
}

// 通过认证方式登录
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
		// 用户名密码登录，需要先验证密码
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

	// 通过统一身份ID获取用户
	user, err := getUserByUniversalId(binding.UniversalId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// 验证用户名密码
func validateUsernamePassword(userOwnerName, password string) (*User, error) {
	// 解析 owner/name 格式
	parts := strings.Split(userOwnerName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid username format, expected: owner/name")
	}

	owner := parts[0]
	name := parts[1]

	// 使用现有的密码验证逻辑
	user, err := CheckUserPassword(owner, name, password, "en")
	if err != nil {
		return nil, err
	}

	return user, nil
}

// 用户主动绑定额外的登录方式
func AddUserIdentityBindingForUser(universalId string, authType string, authValue string) (*UserIdentityBinding, error) {
	// 检查是否已经存在相同的绑定
	existingBinding, err := GetUserIdentityBindingByAuth(authType, authValue)
	if err != nil {
		return nil, err
	}

	if existingBinding != nil {
		if existingBinding.UniversalId == universalId {
			// 已经绑定到当前用户，返回现有绑定
			return existingBinding, nil
		} else {
			// 已经绑定到其他用户，不允许重复绑定
			return nil, fmt.Errorf("此%s已被其他用户绑定", authType)
		}
	}

	// 创建新的身份绑定
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
		return nil, fmt.Errorf("创建身份绑定失败")
	}

	return binding, nil
}

// 用户解除身份绑定
func RemoveUserIdentityBindingForUser(universalId string, authType string) error {
	// 获取用户的所有身份绑定
	bindings, err := GetUserIdentityBindingsByUniversalId(universalId)
	if err != nil {
		return err
	}

	// 检查是否只剩一个身份绑定，如果是则不允许删除
	if len(bindings) <= 1 {
		return fmt.Errorf("不能删除唯一的登录方式，请先绑定其他登录方式")
	}

	// 查找要删除的身份绑定
	var targetBinding *UserIdentityBinding
	for _, binding := range bindings {
		if binding.AuthType == authType {
			targetBinding = binding
			break
		}
	}

	if targetBinding == nil {
		return fmt.Errorf("未找到要删除的身份绑定")
	}

	// 删除身份绑定
	success, err := DeleteUserIdentityBinding(targetBinding.Id)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("删除身份绑定失败")
	}

	return nil
}
