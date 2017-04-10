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

func Parse(r io.Reader) (Email, error) {
	email := Email{}

	msg, err := mail.ReadMessage(r);
	if err != nil {
		return email, err
	}

	var body []byte
	_,err = msg.Body.Read(body);
	if err != nil {
		return email, err
	}

	email.Header, err = decodeHeaderMime(msg.Header)
	if err != nil {
		return email, err
	}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		return email, err
	}

	if mediaType == "" {
		return email, errors.New("No top level mime type specified")
	} else if strings.HasPrefix(mediaType, "multipart/mixed") {
		email.TextBody, email.HTMLBody, email.Attachments, email.EmbeddedFiles, err = parseMultipartMixed(msg.Body, params["boundary"])
		if err != nil {
			return email, err
		}
	} else if strings.HasPrefix(mediaType, "multipart/alternative") {
		email.TextBody, email.HTMLBody, email.EmbeddedFiles, err = parseMultipartAlternative(msg.Body, params["boundary"])
		if err != nil {
			return email, err
		}
	} else if strings.HasPrefix(mediaType, "text/plain") {
		message, _ := ioutil.ReadAll(msg.Body)
		email.TextBody = strings.TrimSuffix(string(message[:]), "\n")
	} else if strings.HasPrefix(mediaType, "text/html") {
		message, _ := ioutil.ReadAll(msg.Body)
		email.HTMLBody = strings.TrimSuffix(string(message[:]), "\n")
	} else {
		return email, errors.New(fmt.Sprintf("Unknown top level mime type: %s", mediaType))
	}

	return email, nil
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

func parseMultipartAlternative(msg io.Reader, boundary string) (textBody, htmlBody string, embeddedFiles []EmbeddedFile, err error) {
	pmr := multipart.NewReader(msg, boundary)
	for {

		pp, err := pmr.NextPart()

		if err == io.EOF {
			break
		}
		if err != nil {
			return textBody, htmlBody, embeddedFiles, err
		}

		ppMediaType, ppParams, err := mime.ParseMediaType(pp.Header.Get("Content-Type"))

		if ppMediaType == "text/plain" {
			ppContent, err := ioutil.ReadAll(pp)
			if err != nil {
				return textBody, htmlBody, embeddedFiles, err
			}

			textBody += strings.TrimSuffix(string(ppContent[:]), "\n")
		} else if ppMediaType == "text/html" {
			ppContent, err := ioutil.ReadAll(pp)
			if err != nil {
				return textBody, htmlBody, embeddedFiles, err
			}

			htmlBody += strings.TrimSuffix(string(ppContent[:]), "\n")
		} else if ppMediaType == "multipart/related" {
			var tb, hb string
			var ef []EmbeddedFile
			tb, hb, ef, err = parseMultipartAlternative(pp, ppParams["boundary"])
			htmlBody += hb
			textBody += tb
			embeddedFiles = append(embeddedFiles, ef...)
		} else if pp.Header.Get("Content-Transfer-Encoding") != "" {
			reference, err := decodeMimeSentence(pp.Header.Get("Content-Id"));
			if err != nil {
				return textBody, htmlBody, embeddedFiles, err
			}
			reference = strings.Trim(reference, "<>")

			decoded, err := decodePartData(pp)
			if err != nil {
				return textBody, htmlBody, embeddedFiles, err
			}

			embeddedFiles = append(embeddedFiles, EmbeddedFile{reference, decoded})
		} else {
			return textBody, htmlBody, embeddedFiles, errors.New(fmt.Sprintf("Can't process multipart/alternative inner mime type: %s", ppMediaType))
		}
	}

	return textBody, htmlBody, embeddedFiles, err
}

func parseMultipartMixed(msg io.Reader, boundary string) (textBody, htmlBody string, attachments []Attachment, embeddedFiles []EmbeddedFile, err error) {
	mr := multipart.NewReader(msg, boundary)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return textBody, htmlBody, attachments, embeddedFiles, err
		}

		pMediaType, pParams, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
		if err != nil {
			return textBody, htmlBody, attachments, embeddedFiles, err
		}

		if strings.HasPrefix(pMediaType, "multipart/alternative") {
			textBody, htmlBody, embeddedFiles, err = parseMultipartAlternative(p, pParams["boundary"])
			if err != nil {
				return textBody, htmlBody, attachments, embeddedFiles, err
			}
		} else if p.FileName() != "" {

			filename, err := decodeMimeSentence(p.FileName());
			if err != nil {
				return textBody, htmlBody, attachments, embeddedFiles, err
			}

			decoded, err := decodePartData(p)
			if err != nil {
				return textBody, htmlBody, attachments, embeddedFiles, err
			}

			attachments = append(attachments, Attachment{filename, decoded})
		} else {
			return textBody, htmlBody, attachments, embeddedFiles, errors.New(fmt.Sprintf("Unknown multipart/mixed nested mime type: %s", pMediaType))
		}
	}

	return textBody, htmlBody, attachments, embeddedFiles, err
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

type Attachment struct {
	Filename string
	Data io.Reader
}

type EmbeddedFile struct {
	CID string
	Data io.Reader
}

type Email struct {
	Header mail.Header
	HTMLBody string
	TextBody string
	Attachments []Attachment
	EmbeddedFiles []EmbeddedFile
}

func (e *Email) Subject() string {
	return e.Header.Get("Subject")
}

func (e *Email) Sender() string {
	return e.Header.Get("Sender")
}

func (e *Email) From() []string {
	result := []string{}

	for _, v := range(strings.Split(e.Header.Get("From"), ",")) {
		t := strings.Trim(v, " ")
		if t != "" {
			result = append(result, t)
		}
	}

	return result
}

func (e *Email) To() []string {
	result := []string{}

	for _, v := range(strings.Split(e.Header.Get("To"), ",")) {
		t := strings.Trim(v, " ")
		if t != "" {
			result = append(result, t)
		}
	}

	return result
}

func (e *Email) Cc() []string {
	result := []string{}

	for _, v := range(strings.Split(e.Header.Get("Cc"), ",")) {
		t := strings.Trim(v, " ")
		if t != "" {
			result = append(result, t)
		}
	}

	return result
}

func (e *Email) Bcc() []string {
	result := []string{}

	for _, v := range(strings.Split(e.Header.Get("Bcc"), ",")) {
		t := strings.Trim(v, " ")
		if t != "" {
			result = append(result, t)
		}
	}

	return result
}

func (e *Email) ReplyTo() []string {
	result := []string{}

	for _, v := range(strings.Split(e.Header.Get("Reply-To"), ",")) {
		t := strings.Trim(v, " ")
		if t != "" {
			result = append(result, t)
		}
	}

	return result
}

func (e *Email) Date() (time.Time, error) {
	t, err := time.Parse(time.RFC1123Z, e.Header.Get("Date"))
	if err == nil {
		return t, err
	}

	return time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", e.Header.Get("Date"))
}

func (e *Email) MessageID() string {
	return strings.Trim(e.Header.Get("Message-ID"), "<>")
}

func (e *Email) InReplyTo() []string {
	result := []string{}

	for _, v := range(strings.Split(e.Header.Get("In-Reply-To"), " ")) {
		if v != "" {
			result = append(result, strings.Trim(v, "<> "))
		}
	}

	return result
}

func (e *Email) References() []string {
	result := []string{}

	for _, v := range(strings.Split(e.Header.Get("References"), " ")) {
		if v != "" {
			result = append(result, strings.Trim(v, "<> "))
		}
	}

	return result
}
