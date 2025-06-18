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
		RedirectURL:  "http://localhost:8080/auth/github/callback",
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
	// 检查是否启用调试模式
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		RunDebugTests()
		return
	}

	// 检查环境变量
	if githubOauthConfig.ClientID == "" || githubOauthConfig.ClientSecret == "" {
		log.Fatal("请设置 GITHUB_CLIENT_ID 和 GITHUB_CLIENT_SECRET 环境变量")
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/auth/github/callback", handleGitHubCallback)
	http.HandleFunc("/callback", handleGitHubCallback)
	http.HandleFunc("/health", handleHealth)

	fmt.Println("🚀 GitHub OAuth 回调处理服务启动在 http://localhost:8080")
	fmt.Println("📝 回调接口: POST/GET http://localhost:8080/auth/github/callback 和 /callback")
	fmt.Println("🧠 智能重定向URL检测: 自动尝试多个可能的回调URL")
	fmt.Println("⚙️  环境变量:")
	fmt.Printf("   GITHUB_CLIENT_ID: %s\n", githubOauthConfig.ClientID)
	fmt.Printf("   GITHUB_CLIENT_SECRET: %s\n", maskSecret(githubOauthConfig.ClientSecret))
	fmt.Println("")
	fmt.Println("📋 API 接口说明:")
	fmt.Println("   GET  /                           - 服务状态页面")
	fmt.Println("   GET  /health                     - 健康检查")
	fmt.Println("   POST /callback                   - GitHub OAuth 回调处理（Casdoor风格）")
	fmt.Println("   GET  /callback                   - GitHub OAuth 回调处理（Casdoor风格）")
	fmt.Println("   POST /auth/github/callback       - GitHub OAuth 回调处理（测试服务风格）")
	fmt.Println("   GET  /auth/github/callback       - GitHub OAuth 回调处理（测试服务风格）")
	fmt.Println("")
	fmt.Println("💡 提示: 请确保GitHub OAuth应用包含以下回调URL:")
	fmt.Println("   - http://localhost:8000/callback")
	fmt.Println("   - http://localhost:8080/auth/github/callback")
	fmt.Println("   - http://127.0.0.1:8000/callback")
	fmt.Println("   - http://127.0.0.1:8080/auth/github/callback")
	fmt.Println("")
	fmt.Println("🐛 调试模式: go run main.go --debug")

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

	// 智能检测正确的重定向URL
	// 尝试多个可能的重定向URL，直到找到有效的一个
	possibleRedirectURLs := []string{
		"http://localhost:8000/callback",             // Casdoor默认
		"http://localhost:8080/auth/github/callback", // 测试服务默认
		"http://127.0.0.1:8000/callback",             // Casdoor localhost变种
		"http://127.0.0.1:8080/auth/github/callback", // 测试服务localhost变种
	}

	// 如果有环境变量指定，优先使用
	if customURL := os.Getenv("GITHUB_REDIRECT_URL"); customURL != "" {
		possibleRedirectURLs = append([]string{customURL}, possibleRedirectURLs...)
	}

	var token *oauth2.Token
	var err error
	var successRedirectURL string

	for _, redirectURL := range possibleRedirectURLs {
		log.Printf("🔍 尝试重定向URL: %s", redirectURL)

		config := *githubOauthConfig
		config.RedirectURL = redirectURL

		token, err = config.Exchange(context.Background(), code)
		if err == nil {
			successRedirectURL = redirectURL
			log.Printf("✅ 成功使用重定向URL: %s", redirectURL)
			break
		} else {
			log.Printf("❌ 重定向URL失败 %s: %v", redirectURL, err)
		}
	}

	if err != nil {
		log.Printf("❌ 所有重定向URL都失败了，最后一个错误: %v", err)
		sendErrorResponse(w, "获取访问令牌失败", fmt.Sprintf("尝试了所有可能的重定向URL都失败了。最后错误: %v", err))
		return
	}

	log.Printf("✅ 获取访问令牌成功: %s (使用重定向URL: %s)", maskToken(token.AccessToken), successRedirectURL)

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

// DebugConfig 调试配置信息
func DebugConfig() {
	fmt.Println("🔧 === GitHub OAuth 调试信息 ===")
	fmt.Printf("GITHUB_CLIENT_ID: %s\n", os.Getenv("GITHUB_CLIENT_ID"))
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if len(clientSecret) > 8 {
		fmt.Printf("GITHUB_CLIENT_SECRET: %s...%s (长度: %d)\n",
			clientSecret[:4], clientSecret[len(clientSecret)-4:], len(clientSecret))
	} else {
		fmt.Printf("GITHUB_CLIENT_SECRET: %s (长度: %d)\n", clientSecret, len(clientSecret))
	}
	fmt.Println()
}

// TestGitHubAPI 测试GitHub API连接
func TestGitHubAPI() error {
	fmt.Println("🌐 测试GitHub API连接...")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return fmt.Errorf("无法连接到GitHub API: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ GitHub API响应状态: %d\n", resp.StatusCode)
	if resp.StatusCode == 401 {
		fmt.Println("✅ GitHub API连接正常（未授权响应符合预期）")
	}
	return nil
}

// TestTokenExchange 测试令牌交换（使用无效code）
func TestTokenExchange() {
	fmt.Println("🔑 测试OAuth配置...")

	config := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:8000/callback",
	}

	// 使用一个明显无效的code来测试配置
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	_, err := config.Exchange(ctx, "invalid_test_code_12345")
	duration := time.Since(start)

	fmt.Printf("⏱️  令牌交换耗时: %v\n", duration)

	if err != nil {
		// 分析错误类型
		if duration > 5*time.Second {
			fmt.Printf("⚠️  响应过慢 (>5s)，可能存在网络问题\n")
		}

		errStr := err.Error()
		if strings.Contains(errStr, "invalid_grant") || strings.Contains(errStr, "bad_verification_code") {
			fmt.Println("✅ OAuth配置正确（收到预期的无效授权码错误）")
		} else if strings.Contains(errStr, "invalid_client") {
			fmt.Println("❌ Client ID或Secret错误")
		} else if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
			fmt.Println("❌ 网络超时，检查网络连接")
		} else {
			fmt.Printf("❓ 未知错误: %v\n", err)
		}
	}
}

// ValidateEnvironment 验证环境变量
func ValidateEnvironment() bool {
	fmt.Println("🔍 验证环境变量...")

	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	if clientID == "" {
		fmt.Println("❌ GITHUB_CLIENT_ID 未设置")
		return false
	}

	if clientSecret == "" {
		fmt.Println("❌ GITHUB_CLIENT_SECRET 未设置")
		return false
	}

	// 验证Client ID格式（GitHub的Client ID通常以Iv开头）
	if len(clientID) < 16 || !strings.Contains(clientID, "Iv") {
		fmt.Printf("⚠️  Client ID格式可能不正确: %s\n", clientID)
	} else {
		fmt.Println("✅ Client ID格式正确")
	}

	// 验证Client Secret长度
	if len(clientSecret) != 40 {
		fmt.Printf("⚠️  Client Secret长度异常: %d (期望40)\n", len(clientSecret))
	} else {
		fmt.Println("✅ Client Secret长度正确")
	}

	return true
}

// RunDebugTests 运行所有调试测试
func RunDebugTests() {
	fmt.Println("🐛 === GitHub OAuth 问题诊断 ===\n")

	DebugConfig()

	if !ValidateEnvironment() {
		fmt.Println("❌ 环境变量验证失败，请检查配置")
		return
	}

	fmt.Println()
	if err := TestGitHubAPI(); err != nil {
		fmt.Printf("❌ GitHub API测试失败: %v\n", err)
	}

	fmt.Println()
	TestTokenExchange()

	fmt.Println("\n💡 建议:")
	fmt.Println("1. 如果OAuth配置正确但仍然失败，请获取新的授权码")
	fmt.Println("2. 确保授权码获取后立即使用（10分钟内）")
	fmt.Println("3. 检查网络连接和防火墙设置")
	fmt.Println("4. 确认GitHub OAuth应用状态正常")
}
