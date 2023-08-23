package email

type EmailService interface {
	Send(EmailModel) error
}
