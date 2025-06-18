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

package controllers

import (
	"encoding/json"
	"strings"

	"github.com/casdoor/casdoor/object"
)

// MergeUsers
// @Title MergeUsers
// @Tag Identity API
// @Description merge two users, delete the source user and transfer its identity bindings to target user
// @Param reserved_user_token body string true "token of the user to be reserved"
// @Param deleted_user_token body string true "token of the user to be deleted"
// @Success 200 {object} object.MergeResult The Response object
// @Failure 400 Bad request
// @Failure 401 Unauthorized
// @router /identity/merge [post]
func (c *ApiController) MergeUsers() {
	// 从 Authorization 头获取 Bearer token
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.ResponseError("Authorization header required")
		return
	}

	// 解析 Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.ResponseError("Invalid authorization header format. Expected: Bearer <token>")
		return
	}

	token := parts[1]

	// 解析token获取用户信息
	claims, err := object.ParseJwtTokenByApplication(token, nil)
	if err != nil {
		c.ResponseError("Invalid token")
		return
	}

	var request struct {
		ReservedUserToken string `json:"reserved_user_token"`
		DeletedUserToken  string `json:"deleted_user_token"`
	}

	err = json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if request.ReservedUserToken == "" || request.DeletedUserToken == "" {
		c.ResponseError("Both reserved_user_token and deleted_user_token are required")
		return
	}

	// 验证当前用户有权限进行合并操作
	// 1. 检查当前用户是否是其中一个token对应的用户
	reservedClaims, err := object.ParseJwtTokenByApplication(request.ReservedUserToken, nil)
	if err != nil {
		c.ResponseError("Invalid reserved_user_token")
		return
	}

	deletedClaims, err := object.ParseJwtTokenByApplication(request.DeletedUserToken, nil)
	if err != nil {
		c.ResponseError("Invalid deleted_user_token")
		return
	}

	// 当前用户必须是要保留的用户或者要删除的用户之一
	currentUserId := claims.User.Name
	if currentUserId != reservedClaims.User.Name && currentUserId != deletedClaims.User.Name {
		c.ResponseError("Unauthorized: You can only merge accounts you own")
		return
	}

	result, err := object.MergeUsers(request.ReservedUserToken, request.DeletedUserToken)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = map[string]interface{}{
		"status":              "ok",
		"universal_id":        result.UniversalId,
		"deleted_user_id":     result.DeletedUserId,
		"merged_auth_methods": result.MergedAuthMethods,
		"message":             "Successfully merged user accounts",
	}
	c.ServeJSON()
}

// GetIdentityInfo
// @Title GetIdentityInfo
// @Tag Identity API
// @Description get user's unified identity information including bound authentication methods
// @Success 200 {object} object The Response object
// @Failure 400 Bad request
// @Failure 401 Unauthorized
// @router /identity/info [get]
func (c *ApiController) GetIdentityInfo() {
	// 从 Authorization 头获取 Bearer token
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.ResponseError("Authorization header required")
		return
	}

	// 解析 Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.ResponseError("Invalid authorization header format. Expected: Bearer <token>")
		return
	}

	token := parts[1]

	// 解析token获取用户信息
	claims, err := object.ParseJwtTokenByApplication(token, nil)
	if err != nil {
		c.ResponseError("Invalid token")
		return
	}

	if claims.UniversalId == "" {
		c.ResponseError("User does not have a unified identity")
		return
	}

	// 获取用户的所有身份绑定
	bindings, err := object.GetUserIdentityBindingsByUniversalId(claims.UniversalId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	authMethods := []map[string]string{}
	for _, binding := range bindings {
		authMethods = append(authMethods, map[string]string{
			"auth_type":  binding.AuthType,
			"auth_value": binding.AuthValue,
		})
	}

	c.Data["json"] = map[string]interface{}{
		"universal_id":       claims.UniversalId,
		"bound_auth_methods": authMethods,
	}
	c.ServeJSON()
}

// BindAuthMethod
// @Title BindAuthMethod
// @Tag Identity API
// @Description bind a new authentication method to user's unified identity
// @Param auth_type body string true "authentication type (email, phone, github, etc.)"
// @Param auth_value body string true "authentication value"
// @Success 200 {object} object The Response object
// @Failure 400 Bad request
// @Failure 401 Unauthorized
// @router /identity/bind [post]
func (c *ApiController) BindAuthMethod() {
	// 从 Authorization 头获取 Bearer token
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.ResponseError("Authorization header required")
		return
	}

	// 解析 Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.ResponseError("Invalid authorization header format. Expected: Bearer <token>")
		return
	}

	token := parts[1]

	// 解析token获取用户信息
	claims, err := object.ParseJwtTokenByApplication(token, nil)
	if err != nil {
		c.ResponseError("Invalid token")
		return
	}

	if claims.UniversalId == "" {
		c.ResponseError("User does not have a unified identity")
		return
	}

	var request struct {
		AuthType  string `json:"auth_type"`
		AuthValue string `json:"auth_value"`
	}

	err = json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if request.AuthType == "" || request.AuthValue == "" {
		c.ResponseError("auth_type and auth_value are required")
		return
	}

	// 绑定新的认证方式
	binding, err := object.AddUserIdentityBindingForUser(claims.UniversalId, request.AuthType, request.AuthValue)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = map[string]interface{}{
		"status":  "ok",
		"message": "Authentication method bound successfully",
		"binding": map[string]string{
			"auth_type":  binding.AuthType,
			"auth_value": binding.AuthValue,
		},
	}
	c.ServeJSON()
}

// UnbindAuthMethod
// @Title UnbindAuthMethod
// @Tag Identity API
// @Description unbind an authentication method from user's unified identity
// @Param auth_type body string true "authentication type to unbind"
// @Success 200 {object} object The Response object
// @Failure 400 Bad request
// @Failure 401 Unauthorized
// @router /identity/unbind [post]
func (c *ApiController) UnbindAuthMethod() {
	// 从 Authorization 头获取 Bearer token
	authHeader := c.Ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.ResponseError("Authorization header required")
		return
	}

	// 解析 Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.ResponseError("Invalid authorization header format. Expected: Bearer <token>")
		return
	}

	token := parts[1]

	// 解析token获取用户信息
	claims, err := object.ParseJwtTokenByApplication(token, nil)
	if err != nil {
		c.ResponseError("Invalid token")
		return
	}

	if claims.UniversalId == "" {
		c.ResponseError("User does not have a unified identity")
		return
	}

	var request struct {
		AuthType string `json:"auth_type"`
	}

	err = json.Unmarshal(c.Ctx.Input.RequestBody, &request)
	if err != nil {
		c.ResponseError("Invalid request body")
		return
	}

	if request.AuthType == "" {
		c.ResponseError("auth_type is required")
		return
	}

	// 解绑认证方式
	err = object.RemoveUserIdentityBindingForUser(claims.UniversalId, request.AuthType)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	c.Data["json"] = map[string]interface{}{
		"status":  "ok",
		"message": "Authentication method unbound successfully",
	}
	c.ServeJSON()
}
