package parsemail

import (
	"encoding/base64"
	"io/ioutil"
	"net/mail"
	"strings"
	"testing"
	"time"
)

func TestParseEmail(t *testing.T) {
	var testData = map[int]struct {
		mailData string

		contentType     string
		content         string
		subject         string
		date            time.Time
		from            []mail.Address
		sender          mail.Address
		to              []mail.Address
		replyTo         []mail.Address
		cc              []mail.Address
		bcc             []mail.Address
		messageID       string
		resentDate      time.Time
		resentFrom      []mail.Address
		resentSender    mail.Address
		resentTo        []mail.Address
		resentReplyTo   []mail.Address
		resentCc        []mail.Address
		resentBcc       []mail.Address
		resentMessageID string
		inReplyTo       []string
		references      []string
		htmlBody        string
		textBody        string
		attachments     []attachmentData
		embeddedFiles   []embeddedFileData
		headerCheck     func(mail.Header, *testing.T)
	}{
		1: {
			mailData: rfc5322exampleA11,
			subject:  "Saying Hello",
			from: []mail.Address{
				{
					Name:    "John Doe",
					Address: "jdoe@machine.example",
				},
			},
			to: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@example.net",
				},
			},
			sender: mail.Address{
				Name:    "Michael Jones",
				Address: "mjones@machine.example",
			},
			messageID: "1234@local.machine.example",
			date:      parseDate("Fri, 21 Nov 1997 09:55:06 -0600"),
			textBody: `This is a message just to say hello.
So, "Hello".`,
		},
		2: {
			mailData: rfc5322exampleA12,
			from: []mail.Address{
				{
					Name:    "Joe Q. Public",
					Address: "john.q.public@example.com",
				},
			},
			to: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@x.test",
				},
				{
					Name:    "",
					Address: "jdoe@example.org",
				},
				{
					Name:    "Who?",
					Address: "one@y.test",
				},
			},
			cc: []mail.Address{
				{
					Name:    "",
					Address: "boss@nil.test",
				},
				{
					Name:    "Giant; \"Big\" Box",
					Address: "sysservices@example.net",
				},
			},
			messageID: "5678.21-Nov-1997@example.com",
			date:      parseDate("Tue, 01 Jul 2003 10:52:37 +0200"),
			textBody:  `Hi everyone.`,
		},
		3: {
			mailData: rfc5322exampleA2a,
			subject:  "Re: Saying Hello",
			from: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@example.net",
				},
			},
			replyTo: []mail.Address{
				{
					Name:    "Mary Smith: Personal Account",
					Address: "smith@home.example",
				},
			},
			to: []mail.Address{
				{
					Name:    "John Doe",
					Address: "jdoe@machine.example",
				},
			},
			messageID:  "3456@example.net",
			inReplyTo:  []string{"1234@local.machine.example"},
			references: []string{"1234@local.machine.example"},
			date:       parseDate("Fri, 21 Nov 1997 10:01:10 -0600"),
			textBody:   `This is a reply to your hello.`,
		},
		4: {
			mailData: rfc5322exampleA2b,
			subject:  "Re: Saying Hello",
			from: []mail.Address{
				{
					Name:    "John Doe",
					Address: "jdoe@machine.example",
				},
			},
			to: []mail.Address{
				{
					Name:    "Mary Smith: Personal Account",
					Address: "smith@home.example",
				},
			},
			messageID:  "abcd.1234@local.machine.test",
			inReplyTo:  []string{"3456@example.net"},
			references: []string{"1234@local.machine.example", "3456@example.net"},
			date:       parseDate("Fri, 21 Nov 1997 11:00:00 -0600"),
			textBody:   `This is a reply to your reply.`,
		},
		5: {
			mailData: rfc5322exampleA3,
			subject:  "Saying Hello",
			from: []mail.Address{
				{
					Name:    "John Doe",
					Address: "jdoe@machine.example",
				},
			},
			to: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@example.net",
				},
			},
			messageID: "1234@local.machine.example",
			date:      parseDate("Fri, 21 Nov 1997 09:55:06 -0600"),
			resentFrom: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@example.net",
				},
			},
			resentTo: []mail.Address{
				{
					Name:    "Jane Brown",
					Address: "j-brown@other.example",
				},
			},
			resentMessageID: "78910@example.net",
			resentDate:      parseDate("Mon, 24 Nov 1997 14:22:01 -0800"),
			textBody: `This is a message just to say hello.
So, "Hello".`,
		},
		6: {
			mailData: data1,
			contentType: `multipart/mixed; boundary=f403045f1dcc043a44054c8e6bbf`,
			content: "",
			subject:  "Peter Paholík",
			from: []mail.Address{
				{
					Name:    "Peter Paholík",
					Address: "peter.paholik@gmail.com",
				},
			},
			to: []mail.Address{
				{
					Name:    "",
					Address: "dusan@kasan.sk",
				},
			},
			messageID: "CACtgX4kNXE7T5XKSKeH_zEcfUUmf2vXVASxYjaaK9cCn-3zb_g@mail.gmail.com",
			date:      parseDate("Fri, 07 Apr 2017 09:17:26 +0200"),
			htmlBody:  "<div dir=\"ltr\"><br></div>",
			attachments: []attachmentData{
				{
					filename:    "Peter Paholík 1 4 2017 2017-04-07.pdf",
					contentType: "application/pdf",
					base64data:  "JVBERi0xLjQNCiW1tbW1DQoxIDAgb2JqDQo8PC9UeXBlL0NhdGFsb2cvUGFnZXMgMiAwIFIvTGFuZyhlbi1VUykgL1N0cnVjdFRyZWVSb290IDY3IDAgUi9NYXJrSW5mbzw8L01hcmtlZCB0cnVlPj4vT3V0cHV0SW50ZW50c1s8PC9UeXBlL091dHB1dEludGVudC9TL0dUU19QREZBMS9PdXRwdXRDb25kZXYgMzk1MzYyDQo+Pg0Kc3RhcnR4cmVmDQo0MTk4ODUNCiUlRU9GDQo=",
				},
			},
		},
		7: {
			mailData: data2,
			contentType: `multipart/alternative; boundary="------------C70C0458A558E585ACB75FB4"`,
			content: "",
			subject:  "Re: Test Subject 2",
			from: []mail.Address{
				{
					Name:    "Sender Man",
					Address: "sender@domain.com",
				},
			},
			to: []mail.Address{
				{
					Name:    "",
					Address: "info@receiver.com",
				},
			},
			cc: []mail.Address{
				{
					Name:    "Cc Man",
					Address: "ccman@gmail.com",
				},
			},
			messageID:  "0e9a21b4-01dc-e5c1-dcd6-58ce5aa61f4f@receiver.com",
			inReplyTo:  []string{"9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@receiver.eu"},
			references: []string{"2f6b7595-c01e-46e5-42bc-f263e1c4282d@receiver.com", "9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@domain.com"},
			date:       parseDate("Fri, 07 Apr 2017 12:59:55 +0200"),
			htmlBody:   `<html>data<img src="part2.9599C449.04E5EC81@develhell.com"/></html>`,
			textBody: `First level
> Second level
>> Third level
>
`,
			embeddedFiles: []embeddedFileData{
				{
					cid:         "part2.9599C449.04E5EC81@develhell.com",
					contentType: "image/png",
					base64data:  "iVBORw0KGgoAAAANSUhEUgAAAQEAAAAYCAIAAAB1IN9NAAAACXBIWXMAAAsTAAALEwEAmpwYYKUKF+Os3baUndC0pDnwNAmLy1SUr2Gw0luxQuV/AwC6cEhVV5VRrwAAAABJRU5ErkJggg==",
				},
			},
		},
		8: {
			mailData: imageContentExample,
			subject:  "Saying Hello",
			from: []mail.Address{
				{
					Name:    "John Doe",
					Address: "jdoe@machine.example",
				},
			},
			to: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@example.net",
				},
			},
			sender: mail.Address{
				Name:    "Michael Jones",
				Address: "mjones@machine.example",
			},
			messageID: "1234@local.machine.example",
			date:      parseDate("Fri, 21 Nov 1997 09:55:06 -0600"),
			contentType: `image/jpeg; x-unix-mode=0644; name="image.gif"`,
			content: `GIF89a;`,
		},
	}

	for index, td := range testData {
		e, err := Parse(strings.NewReader(td.mailData))
		if err != nil {
			t.Error(err)
		}

		if td.contentType != e.ContentType {
			t.Errorf("[Test Case %v] Wrong content type. Expected: %s, Got: %s", index, td.contentType, e.ContentType)
		}

		if td.content != "" {
			b, err := ioutil.ReadAll(e.Content)
			if err != nil {
				t.Error(err)
			} else if td.content != string(b) {
				t.Errorf("[Test Case %v] Wrong content. Expected: %s, Got: %s", index, td.content, string(b))
			}
		}

		if td.subject != e.Subject {
			t.Errorf("[Test Case %v] Wrong subject. Expected: %s, Got: %s", index, td.subject, e.Subject)
		}

		if td.messageID != e.MessageID {
			t.Errorf("[Test Case %v] Wrong messageID. Expected: '%s', Got: '%s'", index, td.messageID, e.MessageID)
		}

		if !td.date.Equal(e.Date) {
			t.Errorf("[Test Case %v] Wrong date. Expected: %v, Got: %v", index, td.date, e.Date)
		}

		d := dereferenceAddressList(e.From)
		if !assertAddressListEq(td.from, d) {
			t.Errorf("[Test Case %v] Wrong from. Expected: %s, Got: %s", index, td.from, d)
		}

		var sender mail.Address
		if e.Sender != nil {
			sender = *e.Sender
		}
		if td.sender != sender {
			t.Errorf("[Test Case %v] Wrong sender. Expected: %s, Got: %s", index, td.sender, sender)
		}

		d = dereferenceAddressList(e.To)
		if !assertAddressListEq(td.to, d) {
			t.Errorf("[Test Case %v] Wrong to. Expected: %s, Got: %s", index, td.to, d)
		}

		d = dereferenceAddressList(e.Cc)
		if !assertAddressListEq(td.cc, d) {
			t.Errorf("[Test Case %v] Wrong cc. Expected: %s, Got: %s", index, td.cc, d)
		}

		d = dereferenceAddressList(e.Bcc)
		if !assertAddressListEq(td.bcc, d) {
			t.Errorf("[Test Case %v] Wrong bcc. Expected: %s, Got: %s", index, td.bcc, d)
		}

		if td.resentMessageID != e.ResentMessageID {
			t.Errorf("[Test Case %v] Wrong resent messageID. Expected: '%s', Got: '%s'", index, td.resentMessageID, e.ResentMessageID)
		}

		if !td.resentDate.Equal(e.ResentDate) && !td.resentDate.IsZero() && !e.ResentDate.IsZero() {
			t.Errorf("[Test Case %v] Wrong resent date. Expected: %v, Got: %v", index, td.resentDate, e.ResentDate)
		}

		d = dereferenceAddressList(e.ResentFrom)
		if !assertAddressListEq(td.resentFrom, d) {
			t.Errorf("[Test Case %v] Wrong resent from. Expected: %s, Got: %s", index, td.resentFrom, d)
		}

		var resentSender mail.Address
		if e.ResentSender != nil {
			resentSender = *e.ResentSender
		}
		if td.resentSender != resentSender {
			t.Errorf("[Test Case %v] Wrong resent sender. Expected: %s, Got: %s", index, td.resentSender, resentSender)
		}

		d = dereferenceAddressList(e.ResentTo)
		if !assertAddressListEq(td.resentTo, d) {
			t.Errorf("[Test Case %v] Wrong resent to. Expected: %s, Got: %s", index, td.resentTo, d)
		}

		d = dereferenceAddressList(e.ResentCc)
		if !assertAddressListEq(td.resentCc, d) {
			t.Errorf("[Test Case %v] Wrong resent cc. Expected: %s, Got: %s", index, td.resentCc, d)
		}

		d = dereferenceAddressList(e.ResentBcc)
		if !assertAddressListEq(td.resentBcc, d) {
			t.Errorf("[Test Case %v] Wrong resent bcc. Expected: %s, Got: %s", index, td.resentBcc, d)
		}

		if !assertSliceEq(td.inReplyTo, e.InReplyTo) {
			t.Errorf("[Test Case %v] Wrong in reply to. Expected: %s, Got: %s", index, td.inReplyTo, e.InReplyTo)
		}

		if !assertSliceEq(td.references, e.References) {
			t.Errorf("[Test Case %v] Wrong references. Expected: %s, Got: %s", index, td.references, e.References)
		}

		d = dereferenceAddressList(e.ReplyTo)
		if !assertAddressListEq(td.replyTo, d) {
			t.Errorf("[Test Case %v] Wrong reply to. Expected: %s, Got: %s", index, td.replyTo, d)
		}

		if td.htmlBody != e.HTMLBody {
			t.Errorf("[Test Case %v] Wrong html body. Expected: '%s', Got: '%s'", index, td.htmlBody, e.HTMLBody)
		}

		if td.textBody != e.TextBody {
			t.Errorf("[Test Case %v] Wrong text body. Expected: '%s', Got: '%s'", index, td.textBody, e.TextBody)
		}

		if len(td.attachments) != len(e.Attachments) {
			t.Errorf("[Test Case %v] Incorrect number of attachments! Expected: %v, Got: %v.", index, len(td.attachments), len(e.Attachments))
		} else {
			attachs := e.Attachments[:]

			for _, ad := range td.attachments {
				found := false

				for i, ra := range attachs {
					b, err := ioutil.ReadAll(ra.Data)
					if err != nil {
						t.Error(err)
					}

					encoded := base64.StdEncoding.EncodeToString(b)
					if ra.Filename == ad.filename && encoded == ad.base64data && ra.ContentType == ad.contentType {
						found = true
						attachs = append(attachs[:i], attachs[i+1:]...)
					}
				}

				if !found {
					t.Errorf("[Test Case %v] Attachment not found: %s", index, ad.filename)
				}
			}

			if len(attachs) != 0 {
				t.Errorf("[Test Case %v] Email contains %v unexpected attachments: %v", index, len(attachs), attachs)
			}
		}

		if len(td.embeddedFiles) != len(e.EmbeddedFiles) {
			t.Errorf("[Test Case %v] Incorrect number of embedded files! Expected: %v, Got: %v.", index, len(td.embeddedFiles), len(e.EmbeddedFiles))
		} else {
			embeds := e.EmbeddedFiles[:]

			for _, ad := range td.embeddedFiles {
				found := false

				for i, ra := range embeds {
					b, err := ioutil.ReadAll(ra.Data)
					if err != nil {
						t.Error(err)
					}

					encoded := base64.StdEncoding.EncodeToString(b)

					if ra.CID == ad.cid && encoded == ad.base64data && ra.ContentType == ad.contentType {
						found = true
						embeds = append(embeds[:i], embeds[i+1:]...)
					}
				}

				if !found {
					t.Errorf("[Test Case %v] Embedded file not found: %s", index, ad.cid)
				}
			}

			if len(embeds) != 0 {
				t.Errorf("[Test Case %v] Email contains %v unexpected embedded files: %v", index, len(embeds), embeds)
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

type attachmentData struct {
	filename    string
	contentType string
	base64data  string
}

type embeddedFileData struct {
	cid         string
	contentType string
	base64data  string
}

func assertSliceEq(a, b []string) bool {
	if len(a) == len(b) && len(a) == 0 {
		return true
	}

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
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

func assertAddressListEq(a, b []mail.Address) bool {
	if len(a) == len(b) && len(a) == 0 {
		return true
	}

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
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

func dereferenceAddressList(al []*mail.Address) (result []mail.Address) {
	for _, a := range al {
		result = append(result, *a)
	}

	return
}

var data1 = `From: =?UTF-8?Q?Peter_Pahol=C3=ADk?= <peter.paholik@gmail.com>
Date: Fri, 7 Apr 2017 09:17:26 +0200
Message-ID: <CACtgX4kNXE7T5XKSKeH_zEcfUUmf2vXVASxYjaaK9cCn-3zb_g@mail.gmail.com>
Subject: =?UTF-8?Q?Peter_Pahol=C3=ADk?=
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

var data2 = `Subject: Re: Test Subject 2
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

var rfc5322exampleA11 = `From: John Doe <jdoe@machine.example>
Sender: Michael Jones <mjones@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

This is a message just to say hello.
So, "Hello".
`

var rfc5322exampleA12 = `From: "Joe Q. Public" <john.q.public@example.com>
To: Mary Smith <mary@x.test>, jdoe@example.org, Who? <one@y.test>
Cc: <boss@nil.test>, "Giant; \"Big\" Box" <sysservices@example.net>
Date: Tue, 1 Jul 2003 10:52:37 +0200
Message-ID: <5678.21-Nov-1997@example.com>

Hi everyone.
`

//todo: not yet implemented in net/mail
//once there is support for this, add it
var rfc5322exampleA13 = `From: Pete <pete@silly.example>
To: A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;
Cc: Undisclosed recipients:;
Date: Thu, 13 Feb 1969 23:32:54 -0330
Message-ID: <testabcd.1234@silly.example>

Testing.
`

//we skipped the first message bcause it's the same as A 1.1
var rfc5322exampleA2a = `From: Mary Smith <mary@example.net>
To: John Doe <jdoe@machine.example>
Reply-To: "Mary Smith: Personal Account" <smith@home.example>
Subject: Re: Saying Hello
Date: Fri, 21 Nov 1997 10:01:10 -0600
Message-ID: <3456@example.net>
In-Reply-To: <1234@local.machine.example>
References: <1234@local.machine.example>

This is a reply to your hello.
`

var rfc5322exampleA2b = `To: "Mary Smith: Personal Account" <smith@home.example>
From: John Doe <jdoe@machine.example>
Subject: Re: Saying Hello
Date: Fri, 21 Nov 1997 11:00:00 -0600
Message-ID: <abcd.1234@local.machine.test>
In-Reply-To: <3456@example.net>
References: <1234@local.machine.example> <3456@example.net>

This is a reply to your reply.
`

var rfc5322exampleA3 = `Resent-From: Mary Smith <mary@example.net>
Resent-To: Jane Brown <j-brown@other.example>
Resent-Date: Mon, 24 Nov 1997 14:22:01 -0800
Resent-Message-ID: <78910@example.net>
From: John Doe <jdoe@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

This is a message just to say hello.
So, "Hello".`

var rfc5322exampleA4 = `Received: from x.y.test
  by example.net
  via TCP
  with ESMTP
  id ABC12345
  for <mary@example.net>;  21 Nov 1997 10:05:43 -0600
Received: from node.example by x.y.test; 21 Nov 1997 10:01:22 -0600
From: John Doe <jdoe@node.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.node.example>

This is a message just to say hello.
So, "Hello".`

var imageContentExample = `From: John Doe <jdoe@machine.example>
Sender: Michael Jones <mjones@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>
Content-Type: image/jpeg;
	x-unix-mode=0644;
	name="image.gif"
Content-Transfer-Encoding: base64

R0lGODlhAQE7`
