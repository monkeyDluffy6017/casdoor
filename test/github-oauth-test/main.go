package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	githubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}
)

// GitHubUser 代表从 GitHub API 获取的用户信息
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
	Blog      string `json:"blog"`
}

// GitHubEmail 代表从 GitHub API 获取的邮箱信息
type GitHubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// CallbackResponse 回调响应结构
type CallbackResponse struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
	User        *GitHubUser `json:"user,omitempty"`
	AccessToken string      `json:"access_token,omitempty"`
	Error       string      `json:"error,omitempty"`
}

func main() {
	// 检查环境变量
	if githubOauthConfig.ClientID == "" || githubOauthConfig.ClientSecret == "" {
		log.Fatal("请设置 GITHUB_CLIENT_ID 和 GITHUB_CLIENT_SECRET 环境变量")
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/auth/github/callback", handleGitHubCallback)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("🚀 GitHub OAuth 回调处理服务启动在 http://localhost:8080")
	fmt.Println("📝 回调接口: POST/GET http://localhost:8080/auth/github/callback")
	fmt.Println("⚙️  环境变量:")
	fmt.Printf("   GITHUB_CLIENT_ID: %s\n", githubOauthConfig.ClientID)
	fmt.Printf("   GITHUB_CLIENT_SECRET: %s\n", maskSecret(githubOauthConfig.ClientSecret))
	fmt.Println("")
	fmt.Println("📋 API 接口说明:")
	fmt.Println("   GET  /                           - 服务状态页面")
	fmt.Println("   GET  /health                     - 健康检查")
	fmt.Println("   POST /auth/github/callback       - GitHub OAuth 回调处理")
	fmt.Println("   GET  /auth/github/callback       - GitHub OAuth 回调处理")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return strings.Repeat("*", len(secret))
	}
	return secret[:4] + strings.Repeat("*", len(secret)-8) + secret[len(secret)-4:]
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>GitHub OAuth 回调服务</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 50px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            max-width: 800px;
            margin: 0 auto;
        }
        .endpoint {
            background: #f6f8fa;
            padding: 15px;
            border-radius: 5px;
            margin: 10px 0;
            border-left: 4px solid #0366d6;
        }
        .method {
            background: #28a745;
            color: white;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 12px;
            margin-right: 10px;
        }
        .method.get { background: #17a2b8; }
        .method.post { background: #28a745; }
        h1 { color: #24292e; }
        code {
            background: #f6f8fa;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
        }
        pre {
            background: #f6f8fa;
            padding: 15px;
            border-radius: 6px;
            overflow-x: auto;
            border: 1px solid #e1e4e8;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔗 GitHub OAuth 回调处理服务</h1>
        <p>这是一个专门处理 GitHub OAuth 2.0 授权回调的服务。</p>

        <h2>📋 API 接口</h2>

        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/health</code>
            <p>健康检查接口，返回服务状态。</p>
        </div>

        <div class="endpoint">
            <span class="method get">GET</span>
            <span class="method post">POST</span>
            <code>/auth/github/callback</code>
            <p>GitHub OAuth 回调处理接口。</p>
            <strong>参数：</strong>
            <ul>
                <li><code>code</code> - GitHub 授权码</li>
                <li><code>state</code> - 状态参数（可选）</li>
            </ul>
        </div>

        <h2>📝 响应格式</h2>
        <p>成功响应示例：</p>
        <pre>{
  "success": true,
  "message": "GitHub OAuth 处理成功",
  "user": {
    "id": 12345,
    "login": "username",
    "name": "User Name",
    "email": "user@example.com",
    "avatar_url": "https://avatars.githubusercontent.com/...",
    "company": "Company Name",
    "location": "Location",
    "bio": "User bio",
    "blog": "https://blog.example.com"
  },
  "access_token": "gho_xxxxxxxxxxxx"
}</pre>

        <p>错误响应示例：</p>
        <pre>{
  "success": false,
  "message": "处理失败",
  "error": "错误详情"
}</pre>

        <h2>🔧 测试方法</h2>
        <p>您可以使用 curl 测试回调接口：</p>
        <pre>curl -X POST "http://localhost:8080/auth/github/callback" \
     -d "code=YOUR_GITHUB_AUTH_CODE"</pre>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "github-oauth-callback",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// 支持 GET 和 POST 请求
	var code, state string

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			sendErrorResponse(w, "解析表单数据失败", err.Error())
			return
		}
		code = r.FormValue("code")
		state = r.FormValue("state")
	} else {
		code = r.URL.Query().Get("code")
		state = r.URL.Query().Get("state")
	}

	log.Printf("📨 收到 GitHub 回调请求: method=%s, code=%s, state=%s", r.Method, maskCode(code), state)

	// 检查授权码
	if code == "" {
		log.Printf("❌ 未收到授权码")
		sendErrorResponse(w, "未收到授权码", "缺少 code 参数")
		return
	}

	log.Printf("✅ 收到授权码: %s", maskCode(code))

	// 使用授权码获取访问令牌
	token, err := githubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("❌ 获取访问令牌失败: %v", err)
		sendErrorResponse(w, "获取访问令牌失败", err.Error())
		return
	}

	log.Printf("✅ 获取访问令牌成功: %s", maskToken(token.AccessToken))

	// 使用访问令牌获取用户信息
	userInfo, err := getUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("❌ 获取用户信息失败: %v", err)
		sendErrorResponse(w, "获取用户信息失败", err.Error())
		return
	}

	log.Printf("✅ 获取用户信息成功: %s (%s)", userInfo.Login, userInfo.Email)

	// 返回成功响应
	response := CallbackResponse{
		Success:     true,
		Message:     "GitHub OAuth 处理成功",
		User:        userInfo,
		AccessToken: token.AccessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, message, error string) {
	response := CallbackResponse{
		Success: false,
		Message: message,
		Error:   error,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func maskCode(code string) string {
	if len(code) <= 10 {
		return strings.Repeat("*", len(code))
	}
	return code[:5] + strings.Repeat("*", len(code)-10) + code[len(code)-5:]
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return strings.Repeat("*", len(token))
	}
	return token[:5] + strings.Repeat("*", len(token)-10) + token[len(token)-5:]
}

func getUserInfo(accessToken string) (*GitHubUser, error) {
	// 创建 HTTP 客户端
	client := &http.Client{Timeout: 10 * time.Second}

	// 获取用户基本信息
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API 错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// 如果用户的公开邮箱为空，尝试获取私有邮箱
	if user.Email == "" {
		email, err := getUserEmail(client, accessToken)
		if err != nil {
			log.Printf("⚠️ 获取用户邮箱失败: %v", err)
		} else {
			user.Email = email
		}
	}

	return &user, nil
}

func getUserEmail(client *http.Client, accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API 错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	// 优先返回主邮箱，其次返回已验证的邮箱
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("未找到已验证的邮箱")
}