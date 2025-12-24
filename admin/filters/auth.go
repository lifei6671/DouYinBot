package filters

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"
	beego "github.com/beego/beego/v2/server/web/context"
)

func BasicAuthFilter() web.FilterFunc {
	return func(ctx *beego.Context) {
		// 只保护 /douyin 开头
		path := ctx.Input.URL()
		if !strings.HasPrefix(path, "/douyin") {
			return
		}

		user, _ := web.AppConfig.String("auth.user")
		pass, _ := web.AppConfig.String("auth.pass")

		// 防止线上误配置导致接口裸奔
		if user == "" || pass == "" {
			ctx.Output.SetStatus(http.StatusUnauthorized)
			ctx.Output.Body([]byte("basic auth not configured"))
			return
		}

		auth := ctx.Input.Header("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Basic ") {
			unauthorized(ctx)
			return
		}

		raw := strings.TrimPrefix(auth, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			unauthorized(ctx)
			return
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			unauthorized(ctx)
			return
		}

		if !secureEquals(parts[0], user) || !secureEquals(parts[1], pass) {
			unauthorized(ctx)
			return
		}

		// ✔ 鉴权通过
		// 可以把用户塞到 context，后续 controller 可用
		ctx.Input.SetData("basic_auth_user", parts[0])
	}
}

func unauthorized(ctx *beego.Context) {
	ctx.Output.Header("WWW-Authenticate", `Basic realm="Douyin API", charset="UTF-8"`)
	ctx.Output.SetStatus(http.StatusUnauthorized)
	ctx.Output.Body([]byte("unauthorized"))
}

func secureEquals(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
