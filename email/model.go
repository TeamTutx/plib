package email

//EmailModel : Structure to define an email
type EmailModel struct {
	FromMail    string
	ToMail      []string
	CcMail      []string
	BccMail     []string
	Subject     string
	Body        string
	SenderName  string
	ReplyTo     []string
	Attachments []AttachmentDetails
}

//AttachmentDetails : Structure to define the attachment files
type AttachmentDetails struct {
	FileName       string
	RenameFileName string
}
