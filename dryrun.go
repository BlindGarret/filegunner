package filegunner

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type DryRunMailer struct {
	fileCreateFunc CreateFileFunc
	file           *os.File
}

const dryrunLogName string = "dryrun.csv"

func NewDryRunMailer(fileCreateFunc CreateFileFunc, logDir string) (*DryRunMailer, error) {
	fullPath := path.Join(logDir, dryrunLogName)
	writeHeader := false
	if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
		writeHeader = true
	}

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	if writeHeader {
		_, err = f.WriteString("Time, To, From, Bcc, Subject, Template, Vars, Attachements, Mail Service,\n")
		if err != nil {
			return nil, err
		}
	}

	return &DryRunMailer{
		fileCreateFunc: fileCreateFunc,
		file:           f,
	}, nil
}

func (m *DryRunMailer) Close() error {
	return m.file.Close()
}

func (m *DryRunMailer) Send(req MailRequest, fileName string, now time.Time) error {
	var sb strings.Builder

	_, err := sb.WriteString(strconv.FormatInt(now.Unix(), 10))
	if err != nil {
		return err
	}

	_, err = sb.WriteString(",")
	if err != nil {
		return err
	}

	err = appendQuotedString(req.To, &sb)
	if err != nil {
		return err
	}

	err = appendQuotedString(req.From, &sb)
	if err != nil {
		return err
	}

	if req.Bcc != nil {
		err = appendQuotedString(*req.Bcc, &sb)
		if err != nil {
			return err
		}
	} else {
		err = appendQuotedString("", &sb)
		if err != nil {
			return err
		}
	}

	err = appendQuotedString(req.Subject, &sb)
	if err != nil {
		return err
	}

	err = appendQuotedString(req.Template, &sb)
	if err != nil {
		return err
	}

	if req.Variables != nil {
		err = appendQuotedString(*req.Variables, &sb)
		if err != nil {
			return err
		}
	} else {
		err = appendQuotedString("", &sb)
		if err != nil {
			return err
		}
	}

	var attachmentBuilder strings.Builder
	for _, attachment := range req.Attachments {
		attachmentBuilder.WriteString(fmt.Sprintf("%s as %s;", attachment.FilePath, attachment.AttachmentName))
	}

	err = appendQuotedString(attachmentBuilder.String(), &sb)
	if err != nil {
		return err
	}

	err = appendQuotedString(req.ServiceID, &sb)
	if err != nil {
		return err
	}

	_, err = sb.WriteString("\n")
	if err != nil {
		return err
	}

	_, err = m.file.WriteString(sb.String())
	return err
}

func appendQuotedString(s string, sb *strings.Builder) error {
	_, err := sb.WriteString("\"")
	if err != nil {
		return err
	}

	_, err = sb.WriteString(strings.ReplaceAll(s, "\"", "'"))
	if err != nil {
		return err
	}

	_, err = sb.WriteString("\"")
	if err != nil {
		return err
	}

	_, err = sb.WriteString(",")
	return err
}
