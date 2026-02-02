package auth

import "github.com/gin-gonic/gin"

func Register(app *gin.RouterGroup) {
	app.Any("/", IndexAPI)
	app.POST("/verify", VerifyAPI)
	app.POST("/reset", ResetAPI)
	app.POST("/register", RegisterAPI) // TODO: deprecated, migrate to /auth/wechat_login
	app.POST("/login", LoginAPI)       // TODO: deprecated, migrate to /auth/wechat_login
	app.POST("/wechat_login", WechatLoginAPI)
	app.POST("/state", StateAPI)
	app.GET("/apikey", KeyAPI)
	app.GET("/userinfo", UserInfoAPI)
	app.POST("/resetkey", ResetKeyAPI)
	app.GET("/package", PackageAPI)
	app.GET("/quota", QuotaAPI)
	app.POST("/buy", BuyAPI)
	app.GET("/subscription", SubscriptionAPI)
	app.POST("/subscribe", SubscribeAPI)
	app.GET("/invite", InviteAPI)
	app.GET("/redeem", RedeemAPI)
}
