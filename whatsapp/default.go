package whatsapp

import (
	"fmt"

	"github.com/TeamTutx/plib/ally"
	"github.com/TeamTutx/plib/conf"
	"github.com/TeamTutx/plib/constant"
	"github.com/TeamTutx/plib/perror"
	"github.com/TeamTutx/plib/phttp"
)

type DefaultWhatsappService struct {
	host    string
	authKey string
}

var defaultWhatsappServiceObj *DefaultWhatsappService

func NewDefaultWhatsappService() {
	defaultWhatsappServiceObj = &DefaultWhatsappService{
		host:    conf.String("whatsapp.wasenderapi.host", ""),
		authKey: conf.String("whatsapp.wasenderapi.api_key", ""),
	}
}

func GetDefaultWhatsappService() *DefaultWhatsappService {
	return defaultWhatsappServiceObj
}

func (w *DefaultWhatsappService) Send(requestPayload WhatsappModel) (err error) {
	var resp phttp.HTTPRes

	url := fmt.Sprintf("%s/%s", w.host, "api/send-message")
	req := phttp.NewReq("POST", url)
	req.Body = requestPayload
	req.Header = map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + w.authKey,
	}
	if resp, err = req.RequestHTTP(); err != nil {
		return
	}

	if resp.StatusCode != 200 {
		if resp.StatusCode == 429 {
			err = perror.BadReqError(nil, "expeccting 200 got 429").SetMsg("Too many request. Please try after some time")
			return
		}
		err = perror.HTTPError(nil, "error while sending notificalion, error code : "+ally.IfToA(resp.StatusCode)).SetMsg(constant.SomethingWentWrong)
		return
	}

	return
}
