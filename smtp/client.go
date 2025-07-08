package smtp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net"
	"net/smtp"
	"strings"

	"github.com/spf13/viper"
)

type Client struct {
	From     string
	Password string
	SmtpHost string
	SmtpPort string
}

func CreateClient() *Client {
	return &Client{
		From:     viper.GetString("SMTP_FROM"),
		Password: "yB$9f66v1",
		SmtpHost: viper.GetString("SMTP_HOST"),
		SmtpPort: viper.GetString("SMTP_PORT"),
	}
}

type loginAuth struct {
	username, password string
}

func LoginAuth() smtp.Auth {
	return &loginAuth{username: viper.GetString("SMTP_USERNAME"), password: "yB$9f66v1"}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown from server")
		}
	}
	return nil, nil
}

func (c Client) Send(input EmailInput, param map[string]interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error Email Sending", r)
		}
	}()

	fmt.Println(input)

	templatePath := "files/email-template/" + input.Template + ".html"
	t, tmpErr := template.ParseFiles(templatePath)
	if tmpErr != nil {
		fmt.Println("Template error", tmpErr.Error())
		return tmpErr
	}

	tlsconfig := &tls.Config{
		ServerName: c.SmtpHost,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := net.Dial("tcp", c.SmtpHost+":"+c.SmtpPort)
	if err != nil {
		fmt.Println("TLS DIAL ERROR", err.Error())
		log.Panic(err)
		return err
	}

	client, err := smtp.NewClient(conn, c.SmtpHost)
	if err != nil {
		fmt.Println("New client error", err.Error())
		log.Panic(err)
		return err
	}

	if err = client.StartTLS(tlsconfig); err != nil {
		fmt.Println("StartTLS error", err.Error())
		log.Panic(err)
		return err
	}

	auth := LoginAuth()

	if err = client.Auth(auth); err != nil {
		fmt.Println("Client auth", err.Error())
		log.Panic(err.Error())
		return err
	}

	if err = client.Mail(c.From); err != nil {
		fmt.Println("Client mail", err.Error())
		log.Panic(err)
		return err
	}

	if err = client.Rcpt(input.Email); err != nil {
		fmt.Println("RCPT ERROR", err.Error())
		log.Panic(err)
		return err
	}

	if len(input.MultiBcc) > 0 {
		for _, bccEmail := range input.MultiBcc {
			if bccEmail != "" {
				if err = client.Rcpt(bccEmail); err != nil {
					fmt.Println("MULTI BCC RCPT ERROR for", bccEmail, err.Error())
					log.Panic(err)
					return err
				}
			}
		}
	}

	w, err := client.Data()
	if err != nil {
		log.Panic(err)
		return err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	headers := map[string]string{
		"From":         fmt.Sprintf("FIBO GLOBAL <%s>", c.From),
		"To":           input.Email,
		"MIME-Version": "1.0",
		"Content-Type": "multipart/mixed; boundary=" + writer.Boundary(),
		"Reply-To":     c.From,
		"Return-Path":  c.From,
		"X-Mailer":     "Go SMTP Client",
		"Subject":      "test",
	}

	if len(input.MultiBcc) > 0 {
		validBccs := []string{}
		for _, bcc := range input.MultiBcc {
			if bcc != "" {
				validBccs = append(validBccs, bcc)
			}
		}
		if len(validBccs) > 0 {
			headers["Bcc"] = strings.Join(validBccs, ", ")
		}
	}

	for key, value := range headers {
		body.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	body.WriteString("\r\n")

	htmlPart, _ := writer.CreatePart(map[string][]string{
		"Content-Type": {"text/html; charset=\"UTF-8\""},
	})
	if err := t.Execute(htmlPart, param); err != nil {
		fmt.Println("HTML template execution error:", err)
		return err
	}

	writer.Close()

	_, err = w.Write(body.Bytes())
	if err != nil {
		log.Panic(err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
		return err
	}

	client.Quit()

	log.Println("Mail sent successfully")
	return nil
}
