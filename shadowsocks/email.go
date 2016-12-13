package shadowsocks

import (
	"os"
	"strconv"

	"golog"

	mailer "github.com/kataras/go-mailer"
)

type Email struct {
	Host     string `json:"host"`
	Username string `json:"user"`
	Password string `json:"passwd"`
	Port     int    `json:"port"`
	Dest     string `json:"dest"`
}

// SendEmail 发送邮件到指定邮箱
func SendEmail(content string) error {

	p, err := strconv.Atoi(os.Getenv("SS_PORT"))
	if err != nil {
		p = 587
	}

	email := &Email{
		Host:     os.Getenv("SS_EMAIL_HOST"),
		Username: os.Getenv("SS_USER_NAME"),
		Password: os.Getenv("SS_PASS_WORD"),
		Port:     p,
		Dest:     os.Getenv("SS_DEST_EMAIL"),
	}

	if email.Host == "" ||
		email.Username == "" ||
		email.Password == "" {
		golog.Debug("EMAIL 信息不完整")
		return nil
	}

	cfg := mailer.Config{
		Host:     email.Host,
		Username: email.Username,
		Password: email.Password,
		Port:     email.Port,
	}

	mailService := mailer.New(cfg)
	var to = []string{email.Dest}

	err = mailService.Send("FROM "+os.Getenv("SS_ID"), content, to...)
	if err != nil {
		return err
	}

	return nil
}
