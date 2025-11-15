package utils

import (
	"context"
	"fmt"
	"math/rand"
	"haodun_manage/backend/database"
	"time"
)

const (
	CaptchaLength = 4
	CaptchaTTL    = 5 * time.Minute
)

func GenerateCaptcha() (string, string) {
	// 生成4位随机数字验证码
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%04d", rand.Intn(10000))
	
	// 生成唯一ID
	captchaID := fmt.Sprintf("captcha:%d", time.Now().UnixNano())
	
	// 存储到Redis，5分钟过期
	ctx := context.Background()
	database.RedisClient.Set(ctx, captchaID, code, CaptchaTTL)
	
	return captchaID, code
}

func VerifyCaptcha(captchaID, code string) bool {
	ctx := context.Background()
	storedCode, err := database.RedisClient.Get(ctx, captchaID).Result()
	if err != nil {
		return false
	}
	
	// 验证后删除
	database.RedisClient.Del(ctx, captchaID)
	
	return storedCode == code
}

