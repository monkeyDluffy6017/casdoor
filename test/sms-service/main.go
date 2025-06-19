package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// SMSRequest 短信请求结构体
type SMSRequest struct {
	Phone       string `json:"phone"`       // 手机号
	PhoneNumber string `json:"phoneNumber"` // 手机号 (Casdoor格式)
	Code        string `json:"code"`        // 验证码
}

// SMSResponse 短信响应结构体
type SMSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 模拟短信发送服务
func sendSMSHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 处理预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 只接受POST请求
	if r.Method != "POST" {
		response := SMSResponse{
			Success: false,
			Message: "只支持POST请求",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 解析请求体
	var smsReq SMSRequest

	// 首先尝试解析表单数据
	r.ParseForm()
	log.Printf("所有表单参数: %v", r.Form)

	if len(r.Form) > 0 {
		// 有表单数据，从表单获取参数
		// Casdoor发送的参数名是phoneNumber和code
		smsReq.Phone = r.FormValue("phoneNumber") // 修改：从phoneNumber获取手机号
		if smsReq.Phone == "" {
			smsReq.Phone = r.FormValue("phone") // 兼容：如果phoneNumber为空，尝试phone
		}
		smsReq.Code = r.FormValue("code")
		log.Printf("从表单获取: phone=%s, code=%s", smsReq.Phone, smsReq.Code)
	} else {
		// 没有表单数据，尝试JSON解析
		err := json.NewDecoder(r.Body).Decode(&smsReq)
		log.Printf("解析JSON结果: err=%v, phone=%s, phoneNumber=%s, code=%s", err, smsReq.Phone, smsReq.PhoneNumber, smsReq.Code)

		// 统一手机号字段：优先使用PhoneNumber，其次使用Phone
		if smsReq.PhoneNumber != "" {
			smsReq.Phone = smsReq.PhoneNumber
		}
	}

	// 记录请求日志
	log.Printf("收到短信发送请求 - 手机号: %s, 验证码: %s", smsReq.Phone, smsReq.Code)

	// 验证手机号
	if smsReq.Phone == "" {
		response := SMSResponse{
			Success: false,
			Message: "手机号不能为空",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证验证码
	if smsReq.Code == "" {
		response := SMSResponse{
			Success: false,
			Message: "验证码不能为空",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 模拟短信发送过程
	log.Printf("正在向手机号 %s 发送验证码: %s", smsReq.Phone, smsReq.Code)

	// 模拟网络延迟
	time.Sleep(100 * time.Millisecond)

	// 这里是模拟发送，实际场景中会调用真实的短信API
	// 比如阿里云短信、腾讯云短信等

	// 模拟发送成功
	response := SMSResponse{
		Success: true,
		Message: fmt.Sprintf("验证码已成功发送到手机号 %s", smsReq.Phone),
		Data: map[string]interface{}{
			"phone":     smsReq.Phone,
			"code":      smsReq.Code,
			"timestamp": time.Now().Unix(),
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("短信发送成功 - 手机号: %s", smsReq.Phone)
}

// 健康检查接口
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "ok",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
		"service": "SMS验证码服务",
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 设置路由
	http.HandleFunc("/oidc_auth/send/sms", sendSMSHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
		<h1>短信验证码服务</h1>
		<p>服务运行正常</p>
		<p>短信发送接口: POST /oidc_auth/send/sms</p>
		<p>健康检查接口: GET /health</p>
		<p>当前时间: %s</p>
		`, time.Now().Format("2006-01-02 15:04:05"))
	})

	port := ":8083"
	log.Printf("短信验证码服务启动，监听端口: %s", port)
	log.Println("短信发送接口: POST http://localhost:8083/oidc_auth/send/sms")
	log.Println("健康检查接口: GET http://localhost:8083/health")

	// 启动HTTP服务器
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("服务启动失败:", err)
	}
}
