package main

import (
	"encoding/json"
	"fmt"

	"github.com/movsb/taoblog/auth"

	"github.com/movsb/taoblog/protocols"
)

func (c *Client) Settings(args []string) {
	if len(args) == 0 {
		return
	}
	option := protocols.Option{}
	resp := c.mustGet(fmt.Sprintf("/options/%s", args[0]))
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&option); err != nil {
		panic(err)
	}
	switch args[0] {
	case "name":
		c.changeName(&option)
	case "login":
		c.changeLogin(&option)
	}
}

func (c *Client) changeName(old *protocols.Option) {
	fmt.Printf("旧的博客名：%s\n", old.Value)
	now := c.line.MustReadLine("新的博客名：")
	if now != "" {
		old.Value = now
		resp := c.mustPostJSON(fmt.Sprintf("/options/%s", old.Name), old)
		resp.Body.Close()
	}
}

func (c *Client) changeLogin(old *protocols.Option) {
	savedAuth := auth.SavedAuth{}
	if err := json.Unmarshal([]byte(old.Value), &savedAuth); err != nil {
		panic(err)
	}
	fmt.Printf("原用户用：%s\n", savedAuth.Username)
	fmt.Printf("原谷歌ClientID：%s\n", savedAuth.GoogleClientID)
	fmt.Printf("原管理员谷歌UserID：%s\n", savedAuth.AdminGoogleID)

	fmt.Println()

	savedAuth.Username = c.line.MustReadLine("新用户名：")
	newPassword := c.line.MustReadLine("新密码：")
	savedAuth.Password = auth.HashPassword(newPassword)
	savedAuth.GoogleClientID = c.line.MustReadLine("新谷歌ClientID：")
	savedAuth.AdminGoogleID = c.line.MustReadLine("新管理员谷歌UserID：")

	old.Value = savedAuth.Encode()

	resp := c.mustPostJSON(fmt.Sprintf("/options/%s", old.Name), old)
	defer resp.Body.Close()
}
