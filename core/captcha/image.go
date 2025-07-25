package captcha

import (
	"context"
	"github.com/mojocn/base64Captcha"
	"go-com/config"
	"go-com/internal/app"
	"strings"
)

/**
图形验证码，有效期5分钟

使用方式：

// 初始化
captcha.InitImage()

// 生成前端需要的数据
captcha.Image.Generate()
// 验证前端提交的数据
captcha.Image.Verify()

*/

var Image image

type image struct {
	bc           *base64Captcha.Captcha
	driverString *base64Captcha.DriverString
	driverMath   *base64Captcha.DriverMath
	store        base64Captcha.Store
}

func InitImage() {
	Image.driverString = &base64Captcha.DriverString{
		Height:          50,
		Width:           150,
		NoiseCount:      2,
		ShowLineOptions: 2,
		Length:          5,
		Source:          "abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ123456789",
		BgColor:         nil,
		Fonts:           nil,
	}
	Image.store = &imageRedisStore{}
	Image.bc = base64Captcha.NewCaptcha(Image.driverString.ConvertFonts(), Image.store)
}

func (c *image) Generate() (id, b64s, answer string, err error) {
	return c.bc.Generate()
}

func (c *image) Verify(id, answer string) bool {
	return c.bc.Store.Verify(id, answer, true)
}

type imageRedisStore struct {
}

func (s *imageRedisStore) key(key string) string {
	return config.C.App.Prefix + ":captcha-image:" + key
}

func (s *imageRedisStore) Set(id string, value string) error {
	ctx := context.TODO()
	key := s.key(id)
	return app.Redis.Set(ctx, key, value, Expire).Err()
}

func (s *imageRedisStore) Get(id string, clear bool) string {
	ctx := context.TODO()
	key := s.key(id)
	value, _ := app.Redis.Get(ctx, key).Result()
	if clear && value != "" {
		app.Redis.Del(ctx, key)
	}

	return value
}

func (s *imageRedisStore) Verify(id, answer string, clear bool) bool {
	if id == "" || answer == "" {
		return false
	}
	v := s.Get(id, clear)
	return strings.EqualFold(v, answer)
}
