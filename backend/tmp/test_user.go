package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://127.0.0.1:8080/api"
	
	// 先登录获取 token
	adminLoginReq := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}
	
	adminLoginBody, _ := json.Marshal(adminLoginReq)
	adminLoginResp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(adminLoginBody))
	if err != nil {
		fmt.Printf("管理员登录失败: %v\n", err)
		return
	}
	defer adminLoginResp.Body.Close()
	
	adminLoginRespBody, _ := io.ReadAll(adminLoginResp.Body)
	if adminLoginResp.StatusCode != 200 {
		fmt.Printf("管理员登录失败: %s\n", string(adminLoginRespBody))
		return
	}
	
	var adminLoginResult map[string]interface{}
	json.Unmarshal(adminLoginRespBody, &adminLoginResult)
	token, ok := adminLoginResult["token"].(string)
	if !ok {
		fmt.Println("无法获取 token")
		return
	}
	fmt.Println("✓ 管理员登录成功")
	
	// 测试创建用户
	testUsername := fmt.Sprintf("testuser_%d", time.Now().Unix())
	createUserReq := map[string]interface{}{
		"username": testUsername,
		"password": "123456",
		"email":    testUsername + "@test.com",
		"status":   1,
		"role_id":  1,
	}
	
	createBody, _ := json.Marshal(createUserReq)
	createReq, _ := http.NewRequest("POST", baseURL+"/users", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	
	createResp, err := http.DefaultClient.Do(createReq)
	if err != nil {
		fmt.Printf("创建用户请求失败: %v\n", err)
		return
	}
	defer createResp.Body.Close()
	
	createRespBody, _ := io.ReadAll(createResp.Body)
	fmt.Printf("创建用户响应: %s\n", string(createRespBody))
	
	if createResp.StatusCode != 200 {
		fmt.Printf("创建用户失败，状态码: %d\n", createResp.StatusCode)
		return
	}
	fmt.Println("✓ 创建用户成功")
	
	// 测试登录
	loginReq := map[string]interface{}{
		"username": testUsername,
		"password": "123456",
	}
	
	loginBody, _ := json.Marshal(loginReq)
	loginResp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		fmt.Printf("登录请求失败: %v\n", err)
		return
	}
	defer loginResp.Body.Close()
	
	loginRespBody, _ := io.ReadAll(loginResp.Body)
	fmt.Printf("登录响应: %s\n", string(loginRespBody))
	
	if loginResp.StatusCode == 200 {
		fmt.Println("✓ 创建和登录测试成功！")
	} else {
		fmt.Printf("✗ 登录失败，状态码: %d\n", loginResp.StatusCode)
	}
}
