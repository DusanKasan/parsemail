package parsemail_test

import (
	"testing"
	"github.com/DusanKasan/parsemail"
	"strings"
	"time"
	"net/mail"
	"encoding/base64"
	"io/ioutil"
)

func TestParseEmail(t *testing.T) {
	var testData = []struct{
		mailData string

		subject string
		from []string
		sender string
		to []string
		replyTo []string
		cc []string
		bcc []string
		messageID string
		inReplyTo []string
		references []string
		date time.Time
		htmlBody string
		textBody string
		attachments []attachmentData
		embeddedFiles []embeddedFileData
		headerCheck func (mail.Header, *testing.T)
	}{
		{
			mailData: Data1,
			subject: "Test Subject 1",
			from: []string{"Peter Paholík <peter.paholik@gmail.com>"},
			to: []string{"dusan@kasan.sk"},
			messageID: "CACtgX4kNXE7T5XKSKeH_zEcfUUmf2vXVASxYjaaK9cCn-3zb_g@mail.gmail.com",
			date: parseDate("Fri, 07 Apr 2017 09:17:26 +0200"),
			htmlBody: "<div dir=\"ltr\"><br></div>",
			attachments: []attachmentData{
				{
					filename: "Peter Paholík 1 4 2017 2017-04-07.pdf",
					base64data: "JVBERi0xLjQNCiW1tbW1DQoxIDAgb2JqDQo8PC9UeXBlL0NhdGFsb2cvUGFnZXMgMiAwIFIvTGFuZyhlbi1VUykgL1N0cnVjdFRyZWVSb290IDY3IDAgUi9NYXJrSW5mbzw8L01hcmtlZCB0cnVlPj4vT3V0cHV0SW50ZW50c1s8PC9UeXBlL091dHB1dEludGVudC9TL0dUU19QREZBMS9PdXRwdXRDb25kZXYgMzk1MzYyDQo+Pg0Kc3RhcnR4cmVmDQo0MTk4ODUNCiUlRU9GDQo=",
				},
			},
		},
		{
			mailData: Data2,
			subject: "Re: Test Subject 2",
			from: []string{"Sender Man <sender@domain.com>"},
			to: []string{"info@receiver.com"},
			cc: []string{"Cc Man <ccman@gmail.com>"},
			messageID: "0e9a21b4-01dc-e5c1-dcd6-58ce5aa61f4f@receiver.com",
			inReplyTo: []string{"9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@receiver.eu"},
			references: []string{"2f6b7595-c01e-46e5-42bc-f263e1c4282d@receiver.com", "9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@domain.com"},
			date: parseDate("Fri, 07 Apr 2017 12:59:55 +0200"),
			htmlBody: `<html>data<img src="part2.9599C449.04E5EC81@develhell.com"/></html>`,
			textBody: `First level
> Second level
>> Third level
>
`,
			embeddedFiles: []embeddedFileData{
				{
					cid: "part2.9599C449.04E5EC81@develhell.com",
					base64data: "iVBORw0KGgoAAAANSUhEUgAAAQEAAAAYCAIAAAB1IN9NAAAACXBIWXMAAAsTAAALEwEAmpwYYKUKF+Os3baUndC0pDnwNAmLy1SUr2Gw0luxQuV/AwC6cEhVV5VRrwAAAABJRU5ErkJggg==",
				},
			},
		},
	}

	for _, td := range testData {
		e, err := parsemail.Parse(strings.NewReader(td.mailData))
		if err != nil {
			t.Error(err)
		}

		if td.subject != e.Subject() {
			t.Errorf("Wrong subject. Expected: %s, Got: %s", td.subject, e.Subject())
		}

		if td.sender != e.Sender() {
			t.Errorf("Wrong sender. Expected: %s, Got: %s", td.sender, e.Sender())
		}

		if !assertSliceEq(td.from, e.From()) {
			t.Errorf("Wrong from. Expected: %s, Got: %s", td.from, e.From())
		}

		if !assertSliceEq(td.inReplyTo, e.InReplyTo()) {
			t.Errorf("Wrong in reply to. Expected: %s, Got: %s", td.inReplyTo, e.InReplyTo())
		}

		if !assertSliceEq(td.references, e.References()) {
			t.Errorf("Wrong references. Expected: %s, Got: %s", td.references, e.References())
		}

		if !assertSliceEq(td.to, e.To()) {
			t.Errorf("Wrong to. Expected: %s, Got: %s", td.to, e.To())
		}

		if !assertSliceEq(td.replyTo, e.ReplyTo()) {
			t.Errorf("Wrong reply to. Expected: %s, Got: %s", td.replyTo, e.ReplyTo())
		}

		if !assertSliceEq(td.cc, e.Cc()) {
			t.Errorf("Wrong cc. Expected: %s, Got: %s", td.cc, e.Cc())
		}

		if !assertSliceEq(td.bcc, e.Bcc()) {
			t.Errorf("Wrong cc. Expected: %s, Got: %s", td.cc, e.Cc())
		}

		date, err := e.Date()
		if err != nil {
			t.Error(err)
		} else if td.date != date {
			t.Errorf("Wrong date. Expected: %v, Got: %v", td.date, date)
		}

		if td.htmlBody != e.HTMLBody {
			t.Errorf("Wrong html body. Expected: '%s', Got: '%s'", td.htmlBody, e.HTMLBody)
		}

		if td.textBody != e.TextBody {
			t.Errorf("Wrong text body. Expected: '%s', Got: '%s'", td.textBody, e.TextBody)
		}

		if td.messageID != e.MessageID() {
			t.Errorf("Wrong messageID. Expected: '%s', Got: '%s'", td.messageID, e.MessageID())
		}

		if len(td.attachments) != len(e.Attachments) {
			t.Errorf("Incorrect number of attachments! Expected: %v, Got: %v.", len(td.attachments), len(e.Attachments))
		} else {
			attachs := e.Attachments[:]

			for _, ad := range(td.attachments) {
				found := false

				for i, ra := range(attachs) {
					b, err := ioutil.ReadAll(ra.Data)
					if err != nil {
						t.Error(err)
					}

					encoded := base64.StdEncoding.EncodeToString(b)
					if ra.Filename == ad.filename && encoded == ad.base64data {
						found = true
						attachs = append(attachs[:i], attachs[i+1:]...)
					}
				}

				if !found {
					t.Errorf("Attachment not found: %s", ad.filename)
				}
			}

			if len(attachs) != 0 {
				t.Errorf("Email contains %v unexpected attachments: %v", len(attachs), attachs)
			}
		}

		if len(td.embeddedFiles) != len(e.EmbeddedFiles) {
			t.Errorf("Incorrect number of embedded files! Expected: %s, Got: %s.", len(td.embeddedFiles), len(e.EmbeddedFiles))
		} else {
			embeds := e.EmbeddedFiles[:]

			for _, ad := range(td.embeddedFiles) {
				found := false

				for i, ra := range(embeds) {
					b, err := ioutil.ReadAll(ra.Data)
					if err != nil {
						t.Error(err)
					}

					encoded := base64.StdEncoding.EncodeToString(b)

					if ra.CID == ad.cid && encoded == ad.base64data {
						found = true
						embeds = append(embeds[:i], embeds[i+1:]...)
					}
				}

				if !found {
					t.Errorf("Embedded file not found: %s", ad.cid)
				}
			}

			if len(embeds) != 0 {
				t.Errorf("Email contains %v unexpected embedded files: %v", len(embeds), embeds)
			}
		}
	}
}

func parseDate(in string) time.Time {
	out, err := time.Parse(time.RFC1123Z, in)
	if err != nil {
		panic(err)
	}

	return out
}

type attachmentData struct{
	filename string
	base64data string
}

type embeddedFileData struct{
	cid string
	base64data string
}

func assertSliceEq(a, b []string) bool {
	if len(a) == len(b) && len(a) == 0 {
		return true
	}

	if a == nil && b == nil {
		return true;
	}

	if a == nil || b == nil {
		return false;
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

var Data1 = `From: =?UTF-8?Q?Peter_Pahol=C3=ADk?= <peter.paholik@gmail.com>
Date: Fri, 7 Apr 2017 09:17:26 +0200
Message-ID: <CACtgX4kNXE7T5XKSKeH_zEcfUUmf2vXVASxYjaaK9cCn-3zb_g@mail.gmail.com>
Subject: Test Subject 1
To: dusan@kasan.sk
Content-Type: multipart/mixed; boundary=f403045f1dcc043a44054c8e6bbf

--f403045f1dcc043a44054c8e6bbf
Content-Type: multipart/alternative; boundary=f403045f1dcc043a3f054c8e6bbd

--f403045f1dcc043a3f054c8e6bbd
Content-Type: text/plain; charset=UTF-8



--f403045f1dcc043a3f054c8e6bbd
Content-Type: text/html; charset=UTF-8

<div dir="ltr"><br></div>

--f403045f1dcc043a3f054c8e6bbd--
--f403045f1dcc043a44054c8e6bbf
Content-Type: application/pdf;
	name="=?UTF-8?Q?Peter_Paholi=CC=81k_1?=
	=?UTF-8?Q?_4_2017_2017=2D04=2D07=2Epdf?="
Content-Disposition: attachment;
	filename="=?UTF-8?Q?Peter_Paholi=CC=81k_1?=
	=?UTF-8?Q?_4_2017_2017=2D04=2D07=2Epdf?="
Content-Transfer-Encoding: base64
X-Attachment-Id: f_j17i0f0d0

JVBERi0xLjQNCiW1tbW1DQoxIDAgb2JqDQo8PC9UeXBlL0NhdGFsb2cvUGFnZXMgMiAwIFIvTGFu
Zyhlbi1VUykgL1N0cnVjdFRyZWVSb290IDY3IDAgUi9NYXJrSW5mbzw8L01hcmtlZCB0cnVlPj4v
T3V0cHV0SW50ZW50c1s8PC9UeXBlL091dHB1dEludGVudC9TL0dUU19QREZBMS9PdXRwdXRDb25k
ZXYgMzk1MzYyDQo+Pg0Kc3RhcnR4cmVmDQo0MTk4ODUNCiUlRU9GDQo=
--f403045f1dcc043a44054c8e6bbf--
`

var Data2 = `Subject: Re: Test Subject 2
To: info@receiver.com
References: <2f6b7595-c01e-46e5-42bc-f263e1c4282d@receiver.com>
 <9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@domain.com>
Cc: Cc Man <ccman@gmail.com>
From: Sender Man <sender@domain.com>
Message-ID: <0e9a21b4-01dc-e5c1-dcd6-58ce5aa61f4f@receiver.com>
Date: Fri, 7 Apr 2017 12:59:55 +0200
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:45.0)
 Gecko/20100101 Thunderbird/45.8.0
MIME-Version: 1.0
In-Reply-To: <9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@receiver.eu>
Content-Type: multipart/alternative;
 boundary="------------C70C0458A558E585ACB75FB4"

This is a multi-part message in MIME format.
--------------C70C0458A558E585ACB75FB4
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 8bit

First level
> Second level
>> Third level
>


--------------C70C0458A558E585ACB75FB4
Content-Type: multipart/related;
 boundary="------------5DB4A1356834BB602A5F88B2"


--------------5DB4A1356834BB602A5F88B2
Content-Type: text/html; charset=utf-8
Content-Transfer-Encoding: 8bit

<html>data<img src="part2.9599C449.04E5EC81@develhell.com"/></html>

--------------5DB4A1356834BB602A5F88B2
Content-Type: image/png
Content-Transfer-Encoding: base64
Content-ID: <part2.9599C449.04E5EC81@develhell.com>

iVBORw0KGgoAAAANSUhEUgAAAQEAAAAYCAIAAAB1IN9NAAAACXBIWXMAAAsTAAALEwEAmpwY
YKUKF+Os3baUndC0pDnwNAmLy1SUr2Gw0luxQuV/AwC6cEhVV5VRrwAAAABJRU5ErkJggg==
--------------5DB4A1356834BB602A5F88B2

--------------C70C0458A558E585ACB75FB4--
`