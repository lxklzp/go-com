package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
	"github.com/wenlng/go-captcha-assets/bindata/chars"
	"github.com/wenlng/go-captcha-assets/resources/fonts/fzshengsksjw"
	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/click"
	"go-com/config"
	"go-com/core/tool"
	"go-com/internal/app"
	"log"
	"strconv"
	"strings"
)

// 行为验证码，有效期5分钟

var Behavior behavior

type behavior struct {
	click click.Captcha
}

func InitBehavior() {
	builder := click.NewBuilder(
		click.WithRangeLen(option.RangeVal{Min: 4, Max: 6}),
		click.WithRangeVerifyLen(option.RangeVal{Min: 2, Max: 4}),
		click.WithRangeThumbColors([]string{
			"#1f55c4",
			"#780592",
			"#2f6b00",
			"#910000",
			"#864401",
			"#675901",
			"#016e5c",
		}),
		click.WithRangeColors([]string{
			"#fde98e",
			"#60c1ff",
			"#fcb08e",
			"#fb88ff",
			"#b4fed4",
			"#cbfaa9",
			"#78d6f8",
		}),
	)

	fonts, err := fzshengsksjw.GetFont()
	if err != nil {
		log.Fatalln(err)
	}
	imgs, err := images.GetImages()
	if err != nil {
		log.Fatalln(err)
	}

	builder.SetResources(
		click.WithChars(chars.GetChineseChars()),
		click.WithFonts([]*truetype.Font{fonts}),
		click.WithBackgrounds(imgs),
	)

	Behavior.click = builder.Make()
}

func (c *behavior) key(key string) string {
	return config.C.App.Prefix + ":captcha-behavior:" + key
}

func (c *behavior) Generate() (string, string, string, error) {
	captData, err := c.click.Generate()
	if err != nil {
		return "", "", "", err
	}

	// 生成坐标、主图、题图
	dotData := captData.GetData()
	if dotData == nil {
		return "", "", "", errors.New("生成行为验证码失败。")
	}
	var masterImageBase64, thumbImageBase64 string
	masterImageBase64, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		return "", "", "", err
	}
	thumbImageBase64, err = captData.GetThumbImage().ToBase64()
	if err != nil {
		return "", "", "", err
	}

	// 缓存坐标
	dotsByte, _ := json.Marshal(dotData)
	key := strconv.Itoa(int(tool.SnowflakeComm.GetId()))
	ctx := context.TODO()
	err = app.Redis.Set(ctx, c.key(key), string(dotsByte), Expire).Err()
	if err != nil {
		return "", "", "", err
	}

	return key, masterImageBase64, thumbImageBase64, nil
}

func (c *behavior) Verify(key string, dots string) bool {
	if key == "" || dots == "" {
		return false
	}

	// 获取缓存的坐标数据
	ctx := context.TODO()
	key = c.key(key)
	dotsCache, _ := app.Redis.Get(ctx, key).Result()
	if dotsCache == "" {
		return false
	}
	app.Redis.Del(ctx, key)

	// 验证坐标

	src := strings.Split(dots, ",")

	var dct map[int]*click.Dot
	json.Unmarshal([]byte(dotsCache), &dct)

	if (len(dct) * 2) != len(src) {
		return false
	}

	for i := 0; i < len(dct); i++ {
		dot := dct[i]
		j := i * 2
		k := i*2 + 1
		sx, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[j]), 64)
		sy, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[k]), 64)

		if !click.CheckPoint(int64(sx), int64(sy), int64(dot.X), int64(dot.Y), int64(dot.Width), int64(dot.Height), 0) {
			return false
		}
	}

	return true
}
