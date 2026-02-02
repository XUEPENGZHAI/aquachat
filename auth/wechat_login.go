package auth

import (
	"chat/globals"
	"chat/utils"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type WechatLoginForm struct {
	Code string `json:"code" binding:"required"`
}

func buildWechatUsername(db *sql.DB, ident string) string {
	if ident == "" {
		ident = utils.GenerateChar(12)
	}

	seed := utils.Md5Encrypt(ident)
	base := "wx_" + seed[:12]
	name := base
	idx := 0
	for IsUserExist(db, name) {
		idx++
		name = fmt.Sprintf("%s_%d", base, idx)
	}

	return name
}

func createWechatUser(db *sql.DB, resp *WechatAuthResponse) (*User, error) {
	ident := resp.UnionID
	if ident == "" {
		ident = resp.OpenID
	}

	username := buildWechatUsername(db, ident)
	password := utils.GenerateChar(48) // strong random placeholder, not for login
	hash := utils.Sha2Encrypt(password)
	bindID := getMaxBindId(db) + 1
	token := utils.Sha2Encrypt(fmt.Sprintf("%s:%s:%s", resp.OpenID, resp.UnionID, utils.GenerateChar(8)))

	if _, err := globals.ExecDb(db, `
        INSERT INTO auth (username, password, wechat_openid, wechat_unionid, bind_id, token)
        VALUES (?, ?, ?, ?, ?, ?)
    `, username, hash, resp.OpenID, resp.UnionID, bindID, token); err != nil {
		// If someone else inserted same unionid/openid concurrently, treat as idempotent: fetch existing.
		if strings.Contains(err.Error(), "Duplicate entry") {
			if existing := GetUserByWechat(db, resp.UnionID, resp.OpenID); existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}

	user := &User{Username: username, Password: hash}
	OnUserCreated(db, user)

	return user, nil
}

// OnUserCreated centralizes post-creation hooks (initial quota, default API key).
func OnUserCreated(db *sql.DB, user *User) {
	if user == nil {
		return
	}
	user.CreateInitialQuota(db)
	user.CreateApiKey(db)
}

func WechatLoginAPI(c *gin.Context) {
	var form WechatLoginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "invalid_param",
		})
		return
	}

	if strings.TrimSpace(form.Code) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "invalid_code",
		})
		return
	}

	client := getWechatAuthClient()
	resp, err := client.ExchangeCode(c, form.Code)
	if err != nil || resp == nil || strings.TrimSpace(resp.OpenID) == "" {
		if err != nil {
			globals.Warn(fmt.Sprintf("wechat exchange code failed: %v", err))
		} else {
			globals.Warn("wechat exchange code failed: empty openid")
		}
		c.JSON(http.StatusBadGateway, gin.H{
			"status": false,
			"error":  "wechat_exchange_failed",
		})
		return
	}

	db := utils.GetDBFromContext(c)

	user := GetUserByWechat(db, resp.UnionID, resp.OpenID)
	if user == nil {
		if user, err = createWechatUser(db, resp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": false,
				"error":  err.Error(),
			})
			return
		}
	}

	if user.IsBanned(db) {
		c.JSON(http.StatusForbidden, gin.H{
			"status": false,
			"error":  "current user is banned",
		})
		return
	}

	token, err := user.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"token":  token,
	})
}
