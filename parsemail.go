package parsemail

import (
	"net/mail"
	"io"
	"strings"
	"mime/multipart"
	"mime"
	"fmt"
	"errors"
	"io/ioutil"
	"time"
	"encoding/base64"
	"bytes"
)

const content_type_multipart_mixed = "multipart/mixed"
const content_type_multipart_alternative = "multipart/alternative"
const content_type_multipart_related = "multipart/related"
const content_type_text_html = "text/html"
const content_type_text_plain = "text/plain"

func Parse(r io.Reader) (email Email, err error) {
	msg, err := mail.ReadMessage(r);
	if err != nil {
		return
	}

	email, err = createEmailFromHeader(msg.Header)
	if err != nil {
		return
	}

	contentType, params, err := parseContentType(msg.Header.Get("Content-Type"))
	if err != nil {
		return
	}

	switch contentType {
	case content_type_multipart_mixed:
		email.TextBody, email.HTMLBody, email.Attachments, email.EmbeddedFiles, err = parseMultipartMixed(msg.Body, params["boundary"])
	case content_type_multipart_alternative:
		email.TextBody, email.HTMLBody, email.EmbeddedFiles, err = parseMultipartAlternative(msg.Body, params["boundary"])
	case content_type_text_plain:
		message, _ := ioutil.ReadAll(msg.Body)
		email.TextBody = strings.TrimSuffix(string(message[:]), "\n")
	case content_type_text_html:
		message, _ := ioutil.ReadAll(msg.Body)
		email.HTMLBody = strings.TrimSuffix(string(message[:]), "\n")
	case "":
		err = errors.New("No top level mime type specified")
	default:
		err = errors.New(fmt.Sprintf("Unknown top level mime type: %s", contentType))
	}

	return
}

func createEmailFromHeader(header mail.Header) (email Email, err error) {
	email.Subject = header.Get("Subject")

	email.From, err = parseAddressList(header.Get("From"))
	if err != nil {
		return
	}

	email.Sender, err = parseAddress(header.Get("Sender"))
	if err != nil {
		return
	}

	email.ReplyTo, err = parseAddressList(header.Get("Reply-To"))
	if err != nil {
		return
	}

	email.To, err = parseAddressList(header.Get("To"))
	if err != nil {
		return
	}

	email.Cc, err = parseAddressList(header.Get("Cc"))
	if err != nil {
		return
	}

	email.Bcc, err = parseAddressList(header.Get("Bcc"))
	if err != nil {
		return
	}

	email.Date, err = parseTime(header.Get("Date"))
	if err != nil {
		return
	}

	email.ResentFrom, err = parseAddressList(header.Get("Resent-From"))
	if err != nil {
		return
	}

	email.ResentSender, err = parseAddress(header.Get("Resent-Sender"))
	if err != nil {
		return
	}

	email.ResentTo, err = parseAddressList(header.Get("Resent-To"))
	if err != nil {
		return
	}

	email.ResentCc, err = parseAddressList(header.Get("Resent-Cc"))
	if err != nil {
		return
	}

	email.ResentBcc, err = parseAddressList(header.Get("Resent-Bcc"))
	if err != nil {
		return
	}

	if header.Get("Resent-Date") == "" {
		email.ResentDate = time.Time{}
	} else {
		email.ResentDate, err = parseTime(header.Get("Resent-Date"))
		if err != nil {
			return
		}
	}

	email.ResentMessageID = parseMessageId(header.Get("Resent-Message-ID"))
	email.MessageID = parseMessageId(header.Get("Message-ID"))
	email.InReplyTo = parseMessageIdList(header.Get("In-Reply-To"))
	email.References = parseMessageIdList(header.Get("References"))

	//decode whole header for easier access to extra fields
	//todo: should we decode? aren't only standard fields mime encoded?
	email.Header, err = decodeHeaderMime(header)
	if err != nil {
		return
	}

	return
}

func parseContentType(contentTypeHeader string) (contentType string, params map[string]string, err error) {
	if contentTypeHeader == "" {
		contentType = content_type_text_plain
		return
	}

	return mime.ParseMediaType(contentTypeHeader)
}

func parseAddress(s string) (*mail.Address, error) {
	if strings.Trim(s, " \n") != "" {
		return mail.ParseAddress(s)
	}

	return nil, nil
}

func parseAddressList(s string) ([]*mail.Address, error) {
	if strings.Trim(s, " \n") != "" {
		return mail.ParseAddressList(s)
	}

	return []*mail.Address{}, nil
}

func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC1123Z, s)
	if err == nil {
		return t, err
	}

	return time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", s)
}

func parseMessageId(s string) string {
	return strings.Trim(s, "<> ")
}

func parseMessageIdList(s string) (result []string) {
	for _, p := range(strings.Split(s, " ")) {
		if strings.Trim(p, " \n") != "" {
			result = append(result, parseMessageId(p))
		}
	}

	return
}

func parseMultipartAlternative(msg io.Reader, boundary string) (textBody, htmlBody string, embeddedFiles []EmbeddedFile, err error) {
	pmr := multipart.NewReader(msg, boundary)
	for {
		part, err := pmr.NextPart()

		if err == io.EOF {
			break
		} else if err != nil {
			return textBody, htmlBody, embeddedFiles, err
		}

		contentType, params, err := mime.ParseMediaType(part.Header.Get("Content-Type"))

		switch contentType {
		case content_type_text_plain:
			ppContent, err := ioutil.ReadAll(part)
			if err != nil {
				return textBody, htmlBody, embeddedFiles, err
			}

			textBody += strings.TrimSuffix(string(ppContent[:]), "\n")
		case content_type_text_html:
			ppContent, err := ioutil.ReadAll(part)
			if err != nil {
				return textBody, htmlBody, embeddedFiles, err
			}

			htmlBody += strings.TrimSuffix(string(ppContent[:]), "\n")
		case content_type_multipart_related:
			var tb, hb string
			var ef []EmbeddedFile
			tb, hb, ef, err = parseMultipartAlternative(part, params["boundary"])
			htmlBody += hb
			textBody += tb
			embeddedFiles = append(embeddedFiles, ef...)
		default:
			if isEmbeddedFile(part) {
				ef, err := decodeEmbeddedFile(part)
				if err != nil {
					return textBody, htmlBody, embeddedFiles, err
				}

				embeddedFiles = append(embeddedFiles, ef)
			} else {
				return textBody, htmlBody, embeddedFiles, errors.New(fmt.Sprintf("Can't process multipart/alternative inner mime type: %s", contentType))
			}
		}
	}

	return textBody, htmlBody, embeddedFiles, err
}

func parseMultipartMixed(msg io.Reader, boundary string) (textBody, htmlBody string, attachments []Attachment, embeddedFiles []EmbeddedFile, err error) {
	mr := multipart.NewReader(msg, boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return textBody, htmlBody, attachments, embeddedFiles, err
		}

		contentType, params, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			return textBody, htmlBody, attachments, embeddedFiles, err
		}

		if contentType == content_type_multipart_alternative {
			textBody, htmlBody, embeddedFiles, err = parseMultipartAlternative(part, params["boundary"])
			if err != nil {
				return textBody, htmlBody, attachments, embeddedFiles, err
			}
		} else if isAttachment(part) {
			at, err := decodeAttachment(part)
			if err != nil {
				return textBody, htmlBody, attachments, embeddedFiles, err
			}

			attachments = append(attachments, at)
		} else {
			return textBody, htmlBody, attachments, embeddedFiles, errors.New(fmt.Sprintf("Unknown multipart/mixed nested mime type: %s", contentType))
		}
	}

	return textBody, htmlBody, attachments, embeddedFiles, err
}

func decodeMimeSentence(s string) (string, error) {
	result := []string{}
	ss := strings.Split(s, " ")

	for _, word := range ss {
		dec := new(mime.WordDecoder)
		w, err := dec.Decode(word)
		if err != nil {
			if len(result) == 0 {
				w = word
			} else {
				w = " " + word
			}
		}

		result = append(result, w)
	}

	return strings.Join(result, ""), nil
}

func decodeHeaderMime(header mail.Header) (mail.Header, error)  {
	parsedHeader := map[string][]string{}

	for headerName, headerData := range header {

		parsedHeaderData := []string{}
		for _, headerValue := range headerData {
			decodedHeaderValue, err := decodeMimeSentence(headerValue)
			if err != nil {
				return mail.Header{}, err
			}
			parsedHeaderData = append(parsedHeaderData, decodedHeaderValue)
		}

		parsedHeader[headerName] = parsedHeaderData
	}

	return mail.Header(parsedHeader), nil
}

func decodePartData(part *multipart.Part) (io.Reader, error) {
	encoding := part.Header.Get("Content-Transfer-Encoding")

	if encoding == "base64" {
		dr := base64.NewDecoder(base64.StdEncoding, part)
		dd, err := ioutil.ReadAll(dr)
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(dd), nil
	} else {
		return nil, errors.New(fmt.Sprintf("Unknown encoding: %s", encoding))
	}
}

func isEmbeddedFile(part *multipart.Part) bool {
	return part.Header.Get("Content-Transfer-Encoding") != ""
}

func decodeEmbeddedFile(part *multipart.Part) (ef EmbeddedFile, err error) {
	cid, err := decodeMimeSentence(part.Header.Get("Content-Id"));
	if err != nil {
		return
	}

	decoded, err := decodePartData(part)
	if err != nil {
		return
	}

	ef.CID = strings.Trim(cid, "<>")
	ef.Data = decoded
	ef.ContentType = part.Header.Get("Content-Type")

	return
}

func isAttachment(part *multipart.Part) bool {
	return part.FileName() != ""
}

func decodeAttachment(part *multipart.Part) (at Attachment, err error) {
	filename, err := decodeMimeSentence(part.FileName());
	if err != nil {
		return
	}

	decoded, err := decodePartData(part)
	if err != nil {
		return
	}

	at.Filename = filename
	at.Data = decoded
	at.ContentType = strings.Split(part.Header.Get("Content-Type"), ";")[0]

	return
}

type Attachment struct {
	Filename string
	ContentType string
	Data io.Reader
}

type EmbeddedFile struct {
	CID string
	ContentType string
	Data io.Reader
}

type MessageID string

type Email struct {
	Header mail.Header

	Subject string
	Sender *mail.Address
	From []*mail.Address
	ReplyTo []*mail.Address
	To []*mail.Address
	Cc []*mail.Address
	Bcc []*mail.Address
	Date time.Time
	MessageID string
	InReplyTo []string
	References []string

	ResentFrom []*mail.Address
	ResentSender *mail.Address
	ResentTo []*mail.Address
	ResentDate time.Time
	ResentCc []*mail.Address
	ResentBcc []*mail.Address
	ResentMessageID string

	Received string

	HTMLBody string
	TextBody string

	Attachments []Attachment
	EmbeddedFiles []EmbeddedFile
}