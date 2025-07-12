package whatsapp

type WhatsappService interface {
	Send(WhatsappModel) error
}
