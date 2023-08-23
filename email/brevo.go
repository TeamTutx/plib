package email

import (
	"net/mail"
	"os"

	"gitlab.com/g-harshit/plib/conf"
	"gitlab.com/g-harshit/plib/constant"
	"gitlab.com/g-harshit/plib/perror"
	"gopkg.in/gomail.v2"
)

type BrevoEmailService struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var brevoEmailServiceObj *BrevoEmailService

func NewBrevoEmailService() {
	brevoEmailServiceObj = &BrevoEmailService{
		Host:     conf.String("smtp.brevo.host", ""),
		Port:     conf.Int("smtp.brevo.port", 0),
		Username: conf.String("smtp.brevo.username", ""),
		Password: conf.String("smtp.brevo.password", ""),
	}
}

func GetBrevoEmailService() *BrevoEmailService {
	return brevoEmailServiceObj
}

func (b *BrevoEmailService) Send(data EmailModel) (err error) {
	var mail *gomail.Message
	if mail, err = buildMail(data); err == nil {
		dialer := gomail.NewDialer(b.Host, b.Port, b.Username, b.Password)
		if err = dialer.DialAndSend(mail); err != nil {
			err = perror.CustomError("error in SendEmail:" + err.Error())
		}
	}
	return
}

//buildMail : build the email structure
func buildMail(emailData EmailModel) (message *gomail.Message, err error) {
	message = gomail.NewMessage()
	header := map[string][]string{
		"From":     {getEmailFrom(emailData)},
		"To":       emailData.ToMail,
		"Cc":       emailData.CcMail,
		"Bcc":      emailData.BccMail,
		"Subject":  {emailData.Subject},
		"Reply-to": emailData.ReplyTo,
	}
	message.SetHeaders(header)
	message.SetBody("text/html", emailData.Body)
	for _, attachment := range emailData.Attachments {
		if attachment.FileName != "" {
			if existsAtLocation(attachment.FileName) {
				message.Attach(attachment.FileName)
			} else {
				err = perror.CustomError("BuildMail: error in attachment | file_path:" + attachment.FileName)
			}
		}
	}
	return
}

//getEmailFrom : Get email From property
func getEmailFrom(email EmailModel) (from string) {
	if email.FromMail == "" {
		email.FromMail = constant.EmailFrom
	}
	address := mail.Address{
		Name:    email.SenderName,
		Address: email.FromMail,
	}
	from = address.String()
	return
}

//existsAtLocation : check if file exists
func existsAtLocation(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}
