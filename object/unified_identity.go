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

// 统一身份结构（极简版）
type UnifiedIdentity struct {
	Id          string `xorm:"varchar(100) pk" json:"id"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
}

// 用户身份绑定结构（极简版）
type UserIdentityBinding struct {
	Id          string `xorm:"varchar(100) pk" json:"id"`
	UnifiedId   string `xorm:"varchar(100)" json:"unifiedId"`
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

// 统一身份操作
func AddUnifiedIdentity(identity *UnifiedIdentity) (bool, error) {
	affected, err := ormer.Engine.Insert(identity)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func GetUnifiedIdentity(id string) (*UnifiedIdentity, error) {
	identity := &UnifiedIdentity{}
	has, err := ormer.Engine.Where("id = ?", id).Get(identity)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return identity, nil
}

func UpdateUnifiedIdentity(id string, identity *UnifiedIdentity) (bool, error) {
	identity.UpdatedTime = util.GetCurrentTime()
	affected, err := ormer.Engine.Where("id = ?", id).Update(identity)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func DeleteUnifiedIdentity(id string) (bool, error) {
	affected, err := ormer.Engine.Where("id = ?", id).Delete(&UnifiedIdentity{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// 用户身份绑定操作
func AddUserIdentityBinding(binding *UserIdentityBinding) (bool, error) {
	affected, err := ormer.Engine.Insert(binding)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func GetUserIdentityBindingsByUnifiedId(unifiedId string) ([]*UserIdentityBinding, error) {
	bindings := []*UserIdentityBinding{}
	err := ormer.Engine.Where("unified_id = ?", unifiedId).Find(&bindings)
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

func DeleteUserIdentityBindingsByUnifiedId(unifiedId string) (bool, error) {
	affected, err := ormer.Engine.Where("unified_id = ?", unifiedId).Delete(&UserIdentityBinding{})
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

// 检查认证方式是否存在
func checkAuthMethodExists(session *xorm.Session, unifiedId, authType, authValue string) (bool, error) {
	count, err := session.Where("unified_id = ? AND auth_type = ? AND auth_value = ?",
		unifiedId, authType, authValue).Count(&UserIdentityBinding{})
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
	err = ormer.Engine.Where("unified_id = ?", universalId).Find(&bindings)
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
func createIdentityBindings(session *xorm.Session, user *User, unifiedId string) error {
	bindings := []*UserIdentityBinding{}

	// GitHub 认证
	if user.GitHub != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "github",
			AuthValue:   user.GitHub,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	// 手机号认证
	if user.Phone != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "phone",
			AuthValue:   user.Phone,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	// 邮箱认证
	if user.Email != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "email",
			AuthValue:   user.Email,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	// 密码认证（用户名密码注册）
	if user.Password != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "password",
			AuthValue:   fmt.Sprintf("%s/%s", user.Owner, user.Name), // owner/name 作为认证值
			CreatedTime: util.GetCurrentTime(),
		})
	}

	// 其他第三方登录
	if user.Google != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "google",
			AuthValue:   user.Google,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	if user.WeChat != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "wechat",
			AuthValue:   user.WeChat,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	if user.QQ != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "qq",
			AuthValue:   user.QQ,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	if user.Facebook != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "facebook",
			AuthValue:   user.Facebook,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	if user.DingTalk != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "dingtalk",
			AuthValue:   user.DingTalk,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	if user.Weibo != "" {
		bindings = append(bindings, &UserIdentityBinding{
			Id:          util.GenerateId(),
			UnifiedId:   unifiedId,
			AuthType:    "weibo",
			AuthValue:   user.Weibo,
			CreatedTime: util.GetCurrentTime(),
		})
	}

	// 批量插入绑定记录
	for _, binding := range bindings {
		_, err := session.Insert(binding)
		if err != nil {
			return err
		}
	}

	return nil
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
	reservedUser, err := getUserByUniversalId(reservedClaims.User.UniversalId)
	if err != nil {
		return nil, err
	}

	deletedUser, err := getUserByUniversalId(deletedClaims.User.UniversalId)
	if err != nil {
		return nil, err
	}

	// 3. 验证合并条件
	if reservedUser.UniversalId == deletedUser.UniversalId {
		return nil, fmt.Errorf("cannot merge the same user")
	}

	// 4. 获取要删除用户的所有身份绑定
	deletedBindings, err := GetUserIdentityBindingsByUnifiedId(deletedUser.UniversalId)
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
				UnifiedId:   reservedUser.UniversalId,
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
	_, err = session.Where("unified_id = ?", deletedUser.UniversalId).Delete(&UserIdentityBinding{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

	// 8. 删除被删除用户的统一身份记录
	_, err = session.Where("id = ?", deletedUser.UniversalId).Delete(&UnifiedIdentity{})
	if err != nil {
		session.Rollback()
		return nil, err
	}

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
	user, err := getUserByUniversalId(binding.UnifiedId)
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
