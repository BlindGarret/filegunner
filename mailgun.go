package filegunner

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
)

type Attachment struct {
	FilePath       string
	AttachmentName string
}

type MailRequest struct {
	From        string
	To          string
	Bcc         *string
	Subject     string
	Template    string
	Variables   *string
	ServiceID   string
	Attachments []Attachment
}

type MailgunMailer struct {
	client        HttpClient
	readFileFunc  ReadFileFunc
	serviceLookup map[string]MailgunService
}

func NewMailgunMailer(client HttpClient, readFileFunc ReadFileFunc, serviceLookup map[string]MailgunService) *MailgunMailer {
	return &MailgunMailer{
		client:        client,
		readFileFunc:  readFileFunc,
		serviceLookup: serviceLookup,
	}
}

func (m *MailgunMailer) Send(mailReq MailRequest, fileName string) error {
	var bs bytes.Buffer
	w := multipart.NewWriter(&bs)

	err := w.WriteField("from", mailReq.From)
	if err != nil {
		return err
	}

	err = w.WriteField("to", mailReq.To)
	if err != nil {
		return err
	}

	if mailReq.Bcc != nil {
		err = w.WriteField("bcc", *mailReq.Bcc)
		if err != nil {
			return err
		}
	}

	err = w.WriteField("subject", mailReq.Subject)
	if err != nil {
		return err
	}

	err = w.WriteField("template", mailReq.Template)
	if err != nil {
		return err
	}

	if mailReq.Variables != nil {
		err = w.WriteField("h:X-Mailgun-Variables", *mailReq.Variables)
		if err != nil {
			return err
		}
	}

	if mailReq.Attachments != nil {
		for _, attachment := range mailReq.Attachments {
			attachmentWriter, err := w.CreateFormFile("attachment", attachment.AttachmentName)
			if err != nil {
				return err
			}
			bs, err := m.readFileFunc(attachment.FilePath)
			if err != nil {
				return err
			}
			_, err = attachmentWriter.Write(bs)
			if err != nil {
				return err
			}
		}
	}

	err = w.Close()
	if err != nil {
		return err
	}

	service, ok := m.serviceLookup[mailReq.ServiceID]
	if !ok {
		return fmt.Errorf("service not found: %s", mailReq.ServiceID)
	}

	httpReq, err := http.NewRequest("POST", service.Url, &bs)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", w.FormDataContentType())
	httpReq.SetBasicAuth("api", service.ApiKey)
	resp, err := m.client.Do(httpReq)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email with status code: %d", resp.StatusCode)
	}

	return nil
}
