package whatsapp

type Template struct {
}

type WhatsappModel struct {
	Text string `json:"text"`
	To   string `json:"to"`
}
