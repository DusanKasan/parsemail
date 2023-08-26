package parsemail

import (
	"encoding/base64"
	"fmt"
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
			mailData:    data1,
			contentType: `multipart/mixed; boundary=f403045f1dcc043a44054c8e6bbf`,
			content:     "",
			subject:     "Peter Paholík",
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
					filename:    "Peter Paholík 1 4 2017 2017-04-07.json",
					contentType: "application/json",
					data:        "[1, 2, 3]",
				},
			},
		},
		7: {
			mailData:    data2,
			contentType: `multipart/alternative; boundary="------------C70C0458A558E585ACB75FB4"`,
			content:     "",
			subject:     "Re: Test Subject 2",
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
			messageID:   "1234@local.machine.example",
			date:        parseDate("Fri, 21 Nov 1997 09:55:06 -0600"),
			contentType: `image/jpeg; x-unix-mode=0644; name="image.gif"`,
			content:     `GIF89a;`,
		},
		9: {
			contentType: `multipart/mixed; boundary="0000000000007e2bb40587e36196"`,
			mailData:    textPlainInMultipart,
			subject:     "Re: kern/54143 (virtualbox)",
			from: []mail.Address{
				{
					Name:    "Rares",
					Address: "rares@example.com",
				},
			},
			to: []mail.Address{
				{
					Name:    "",
					Address: "bugs@example.com",
				},
			},
			date:     parseDate("Fri, 02 May 2019 11:25:35 +0300"),
			textBody: `plain text part`,
		},
		10: {
			contentType: `multipart/mixed; boundary="0000000000007e2bb40587e36196"`,
			mailData:    textHTMLInMultipart,
			subject:     "Re: kern/54143 (virtualbox)",
			from: []mail.Address{
				{
					Name:    "Rares",
					Address: "rares@example.com",
				},
			},
			to: []mail.Address{
				{
					Name:    "",
					Address: "bugs@example.com",
				},
			},
			date:     parseDate("Fri, 02 May 2019 11:25:35 +0300"),
			textBody: ``,
			htmlBody: "<div dir=\"ltr\"><div>html text part</div><div><br></div><div><br><br></div></div>",
		},
		11: {
			mailData: rfc5322exampleA12WithTimezone,
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
		12: {
			contentType: "multipart/mixed; boundary=f403045f1dcc043a44054c8e6bbf",
			mailData:    attachment7bit,
			subject:     "Peter Foobar",
			from: []mail.Address{
				{
					Name:    "Peter Foobar",
					Address: "peter.foobar@gmail.com",
				},
			},
			to: []mail.Address{
				{
					Name:    "",
					Address: "dusan@kasan.sk",
				},
			},
			messageID: "CACtgX4kNXE7T5XKSKeH_zEcfUUmf2vXVASxYjaaK9cCn-3zb_g@mail.gmail.com",
			date:      parseDate("Tue, 02 Apr 2019 11:12:26 +0000"),
			htmlBody:  "<div dir=\"ltr\"><br></div>",
			attachments: []attachmentData{
				{
					filename:    "unencoded.csv",
					contentType: "application/csv",
					data:        fmt.Sprintf("\n"+`"%s", "%s", "%s", "%s", "%s"`+"\n"+`"%s", "%s", "%s", "%s", "%s"`+"\n", "Some", "Data", "In", "Csv", "Format", "Foo", "Bar", "Baz", "Bum", "Poo"),
				},
			},
		},
		13: {
			contentType: "multipart/related; boundary=\"000000000000ab2e2205a26de587\"",
			mailData:    multipartRelatedExample,
			subject:     "Saying Hello",
			from: []mail.Address{
				{
					Name:    "John Doe",
					Address: "jdoe@machine.example",
				},
			},
			sender: mail.Address{
				Name:    "Michael Jones",
				Address: "mjones@machine.example",
			},
			to: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@example.net",
				},
			},
			messageID: "1234@local.machine.example",
			date:      parseDate("Fri, 21 Nov 1997 09:55:06 -0600"),
			htmlBody:  "<div dir=\"ltr\"><div>Time for the egg.</div><div><br></div><div><br><br></div></div>",
			textBody:  "Time for the egg.",
		},
		14: {
			mailData:    data3,
			contentType: `multipart/mixed; boundary=f403045f1dcc043a44054c8e6bbf`,
			content:     "",
			subject:     "Peter Paholík",
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
					filename:    "Peter Paholík 1 4 2017 2017-04-07.json",
					contentType: "application/json",
					data:        "[1, 2, 3]",
				},
			},
		},
		15: {
			contentType: "text/plain; charset=utf-8",
			mailData:    rfc2045exampleA,
			subject:     "Lead from Allstate LeadVantage",
			from: []mail.Address{
				{
					Address: "LVsupport@allstateleadvantage.com",
				},
			},
			to: []mail.Address{
				{
					Address: "test@email.com",
				},
			},
			replyTo: []mail.Address{
				{
					Address: "no-reply@allstateleadvantage.com",
				},
			},
			messageID: "0100017fcf817777-481efc68-4a9a-4c11-ba2c-40ff0357e7b1-000000@email.amazonses.com",
			date:      parseDate("Mon, 28 Mar 2022 07:50:42 +0000"),
			textBody:  rfc2045exampleAtext,
		},
		16: {
			contentType: `text/html; charset="utf-8"`,
			mailData:    rfc2045exampleB,
			subject:     "New Business Property/Casualty Lead Received (#245200111)",
			from: []mail.Address{
				{
					Name:    "AllWebLeads",
					Address: "no-reply@allwebleads.com",
				},
			},
			to: []mail.Address{
				{
					Address: "sample@example.com",
				},
			},
			replyTo: []mail.Address{
				{
					Address: "no-reply@allwebleads.com",
				},
			},
			messageID: "1187856165.40703531648591546580.JavaMail.app@rapp51.atlis1",
			date:      parseDate("Tue, 29 Mar 2022 22:05:46 +0000"),
			htmlBody:  rfc2045exampleBhtml,
		},
		17: {
			contentType: "multipart/related; boundary=\"000000000000ab2e2205a26de587\"",
			mailData:    multipartRelatedExampleQuoted,
			subject:     "Saying Hello",
			from: []mail.Address{
				{
					Name:    "John Doe",
					Address: "jdoe@machine.example",
				},
			},
			sender: mail.Address{
				Name:    "Michael Jones",
				Address: "mjones@machine.example",
			},
			to: []mail.Address{
				{
					Name:    "Mary Smith",
					Address: "mary@example.net",
				},
			},
			messageID: "1234@local.machine.example",
			date:      parseDate("Fri, 21 Nov 1997 09:55:06 -0600"),
			htmlBody:  rfc2045exampleBhtml,
			textBody:  "Time for the egg. Should we hardboil the egg or fry it. We can scramble it or poach it.",
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

					if ra.Filename == ad.filename && ra.ContentType == ad.contentType {
						found = true
						attachs = append(attachs[:i], attachs[i+1:]...)
					}

					if string(b) != ad.data {
						t.Errorf("[Test Case %v] Bad data for attachment: \nEXPECTED:\n%s\nHAVE:\n%s", index, ad.data, string(b))
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
	data        string
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
Content-Type: application/json;
	name="=?UTF-8?Q?Peter_Paholi=CC=81k_1?=
	=?UTF-8?Q?_4_2017_2017=2D04=2D07=2Ejson?="
Content-Disposition: attachment;
	filename="=?UTF-8?Q?Peter_Paholi=CC=81k_1?=
	=?UTF-8?Q?_4_2017_2017=2D04=2D07=2Ejson?="
Content-Transfer-Encoding: base64
X-Attachment-Id: f_j17i0f0d0

WzEsIDIsIDNd
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

var data3 = `From: =?UTF-8?Q?Peter_Pahol=C3=ADk?= <peter.paholik@gmail.com>
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
Content-Type: application/json;
	name="=?UTF-8?Q?Peter_Paholi=CC=81k_1?=
	=?UTF-8?Q?_4_2017_2017=2D04=2D07=2Ejson?="
Content-Disposition: attachment;
	filename="=?UTF-8?Q?Peter_Paholi=CC=81k_1?=
	=?UTF-8?Q?_4_2017_2017=2D04=2D07=2Ejson?="
Content-Transfer-Encoding: BASE64
X-Attachment-Id: f_j17i0f0d0

WzEsIDIsIDNd
--f403045f1dcc043a44054c8e6bbf--
`

var textPlainInMultipart = `From: Rares <rares@example.com>
Date: Thu, 2 May 2019 11:25:35 +0300
Subject: Re: kern/54143 (virtualbox)
To: bugs@example.com
Content-Type: multipart/mixed; boundary="0000000000007e2bb40587e36196"

--0000000000007e2bb40587e36196
Content-Type: text/plain; charset="UTF-8"

plain text part
--0000000000007e2bb40587e36196--
`

var textHTMLInMultipart = `From: Rares <rares@example.com>
Date: Thu, 2 May 2019 11:25:35 +0300
Subject: Re: kern/54143 (virtualbox)
To: bugs@example.com
Content-Type: multipart/mixed; boundary="0000000000007e2bb40587e36196"

--0000000000007e2bb40587e36196
Content-Type: text/html; charset="UTF-8"

<div dir="ltr"><div>html text part</div><div><br></div><div><br><br></div></div>

--0000000000007e2bb40587e36196--
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

var rfc5322exampleA12WithTimezone = `From: "Joe Q. Public" <john.q.public@example.com>
To: Mary Smith <mary@x.test>, jdoe@example.org, Who? <one@y.test>
Cc: <boss@nil.test>, "Giant; \"Big\" Box" <sysservices@example.net>
Date: Tue, 1 Jul 2003 10:52:37 +0200 (GMT)
Message-ID: <5678.21-Nov-1997@example.com>

Hi everyone.
`

// todo: not yet implemented in net/mail
// once there is support for this, add it
var rfc5322exampleA13 = `From: Pete <pete@silly.example>
To: A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;
Cc: Undisclosed recipients:;
Date: Thu, 13 Feb 1969 23:32:54 -0330
Message-ID: <testabcd.1234@silly.example>

Testing.
`

// we skipped the first message bcause it's the same as A 1.1
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

var multipartRelatedExample = `MIME-Version: 1.0
From: John Doe <jdoe@machine.example>
Sender: Michael Jones <mjones@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>
Subject: ooops
To: test@example.rocks
Content-Type: multipart/related; boundary="000000000000ab2e2205a26de587"

--000000000000ab2e2205a26de587
Content-Type: multipart/alternative; boundary="000000000000ab2e1f05a26de586"

--000000000000ab2e1f05a26de586
Content-Type: text/plain; charset="UTF-8"

Time for the egg.

--000000000000ab2e1f05a26de586
Content-Type: text/html; charset="UTF-8"

<div dir="ltr"><div>Time for the egg.</div><div><br></div><div><br><br></div></div>

--000000000000ab2e1f05a26de586--


--000000000000ab2e2205a26de587--
`
var attachment7bit = `From: =?UTF-8?Q?Peter_Foobar?= <peter.foobar@gmail.com>
Date: Tue, 2 Apr 2019 11:12:26 +0000
Message-ID: <CACtgX4kNXE7T5XKSKeH_zEcfUUmf2vXVASxYjaaK9cCn-3zb_g@mail.gmail.com>
Subject: =?UTF-8?Q?Peter_Foobar?=
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
Content-Type: application/csv; 
	name="unencoded.csv"
Content-Transfer-Encoding: 7bit
Content-Disposition: attachment; 
	filename="unencoded.csv"


"Some", "Data", "In", "Csv", "Format"
"Foo", "Bar", "Baz", "Bum", "Poo"

--f403045f1dcc043a44054c8e6bbf--
`

var rfc2045exampleA = `From 0100017fcf817777-481efc68-4a9a-4c11-ba2c-40ff0357e7b1-000000@amazonses.com  Mon Mar 28 07:50:43 2022
Return-Path: <0100017fcf817777-481efc68-4a9a-4c11-ba2c-40ff0357e7b1-000000@amazonses.com>
X-Original-To: test@email.com
Delivered-To: leads@reciever.com
Message-ID: <0100017fcf817777-481efc68-4a9a-4c11-ba2c-40ff0357e7b1-000000@email.amazonses.com>
Date: Mon, 28 Mar 2022 07:50:42 +0000
Subject: Lead from Allstate LeadVantage
From: LVsupport@allstateleadvantage.com
Reply-To: no-reply@allstateleadvantage.com
To: test@email.com
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: quoted-printable


You just received a lead! Please check your lead management system, or u=
se the contact information
below.  Please do not respond to this email ad=
dress, as it is not active. You may also view your leads
in Allstate Lead=
Vantage. Please call Allstate LeadVantage Support at 855-317-4233 or sign u=
p here:
https://allstateleadvantage.com/#/orders/list

Lead Informati=
on:
Unique ID: 138296007
Vertical: Auto Insurance
Alliance URL: https=
://agencygateway.allstate.com/ALLIANCE/launch?AgentNumber=3DA0c3858&ST=3DNV=
&FunctionType=3DAF&SourceOfLaunchPoint=3D01&ControlNumber=3D198220870336180=

Contact Information:
First Name: Brenda
Last Name: Qualls
Phone Nu=
mber: (702) 485-1038
Email Address: brendaqualls29@yahoo.com
Street Add=
ress: 3236 Brayton Mist Dr
City: North Las Vegas
State: NV
Zip: 89081=

Are You A Homeowner: Yes
Best Time To Contact:=20
Vendor:
Vendor Nam=
e: Inside Response
Order Information:
Name: Custom Order 1
Policy Det=
ails:
Self Credit Rating: Good (620 - 719)
Currently Insured: Yes
Cur=
rent Insurance Company: State Farm County
Insured Since: 03/28/2020
Pol=
icy Start: 03/28/2020
Policy Expiration: 05/28/2022
Desired Coverage Ty=
pe: standard
Desired Collision Deductible: 1000
Desired Comprehensive D=
eductible: 1000
Driver 1:
Gender: female
Marital Status: married
Ed=
ucation Level: ged
Occupation: other
Date of Birth: 01/29/1981
Age Li=
censed: 19
Has Valid License: Yes
Has DUI: No
Requires SR-22: No
Re=
lation to applicant: self
Years Employed: 2
Years at Residence: 2
Has=
 Tickets / Accidents: No
Vehicle 1:
Type: 2006 LEXUS SC 430 2WD CONVERT=
IBLE - 4.3L V8  FI  DOHC 32V  F
Vin: JTHFN48Y060000000
Leased: No
Pri=
mary Use: Pleasure Use Only
Commute Days: 5
Daily Mileage: 5
Annual M=
ileage: 15000
Has Alarm: Yes
Garage: nocover
`

var rfc2045exampleAtext string = `
You just received a lead! Please check your lead management system, or use the contact information
below.  Please do not respond to this email address, as it is not active. You may also view your leads
in Allstate LeadVantage. Please call Allstate LeadVantage Support at 855-317-4233 or sign up here:
https://allstateleadvantage.com/#/orders/list

Lead Information:
Unique ID: 138296007
Vertical: Auto Insurance
Alliance URL: https://agencygateway.allstate.com/ALLIANCE/launch?AgentNumber=A0c3858&ST=NV&FunctionType=AF&SourceOfLaunchPoint=01&ControlNumber=198220870336180
Contact Information:
First Name: Brenda
Last Name: Qualls
Phone Number: (702) 485-1038
Email Address: brendaqualls29@yahoo.com
Street Address: 3236 Brayton Mist Dr
City: North Las Vegas
State: NV
Zip: 89081
Are You A Homeowner: Yes
Best Time To Contact: 
Vendor:
Vendor Name: Inside Response
Order Information:
Name: Custom Order 1
Policy Details:
Self Credit Rating: Good (620 - 719)
Currently Insured: Yes
Current Insurance Company: State Farm County
Insured Since: 03/28/2020
Policy Start: 03/28/2020
Policy Expiration: 05/28/2022
Desired Coverage Type: standard
Desired Collision Deductible: 1000
Desired Comprehensive Deductible: 1000
Driver 1:
Gender: female
Marital Status: married
Education Level: ged
Occupation: other
Date of Birth: 01/29/1981
Age Licensed: 19
Has Valid License: Yes
Has DUI: No
Requires SR-22: No
Relation to applicant: self
Years Employed: 2
Years at Residence: 2
Has Tickets / Accidents: No
Vehicle 1:
Type: 2006 LEXUS SC 430 2WD CONVERTIBLE - 4.3L V8  FI  DOHC 32V  F
Vin: JTHFN48Y060000000
Leased: No
Primary Use: Pleasure Use Only
Commute Days: 5
Daily Mileage: 5
Annual Mileage: 15000
Has Alarm: Yes
Garage: nocover`

var rfc2045exampleB string = `From v-biheobc_begnlldjf_icanamoe_icanamoe_a-1@bounce.allweb.mkt3103.com  Tue Mar 29 22:05:46 2022
Return-Path: <v-biheobc_begnlldjf_icanamoe_icanamoe_a-1@bounce.allweb.mkt3103.com>
X-Original-To: sample@example.com
Delivered-To: leads@reciever.com
Received: by mail2792.allweb.mkt3188.com id h8e1bk2r7ao5 for <sample@example.com>; Tue, 29 Mar 2022 22:05:46 +0000 (envelope-from <v-biheobc_begnlldjf_icanamoe_icanamoe_a-1@bounce.allweb.mkt3103.com>)
Date: Tue, 29 Mar 2022 22:05:46 +0000 (GMT)
From: AllWebLeads <no-reply@allwebleads.com>
Reply-To: no-reply@allwebleads.com
To: sample@example.com
Message-ID: <1187856165.40703531648591546580.JavaMail.app@rapp51.atlis1>
Subject: New Business Property/Casualty Lead Received (#245200111)
Content-Type: text/html; charset="utf-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">
=09<div>Time for the egg.</div>
=09<div><br/></div>
=09<div><br/><br></div>
=09<div>Should we hardboil the egg or fry it. We can scramble it or poach i=
t.</div>
</div>`

var rfc2045exampleBhtml string = `<div dir="ltr">
	<div>Time for the egg.</div>
	<div><br/></div>
	<div><br/><br></div>
	<div>Should we hardboil the egg or fry it. We can scramble it or poach it.</div>
</div>`
var multipartRelatedExampleQuoted = `MIME-Version: 1.0
From: John Doe <jdoe@machine.example>
Sender: Michael Jones <mjones@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>
Subject: ooops
To: test@example.rocks
Content-Type: multipart/related; boundary="000000000000ab2e2205a26de587"

--000000000000ab2e2205a26de587
Content-Type: multipart/alternative; boundary="000000000000ab2e1f05a26de586"

--000000000000ab2e1f05a26de586
Content-Type: text/plain; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

Time for the egg. Should we hardboil the egg or fry it. We can scramble it =
or poach it.

--000000000000ab2e1f05a26de586
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

<div dir=3D"ltr">
=09<div>Time for the egg.</div>
=09<div><br/></div>
=09<div><br/><br></div>
=09<div>Should we hardboil the egg or fry it. We can scramble it or poach i=
t.</div>
</div>

--000000000000ab2e1f05a26de586--


--000000000000ab2e2205a26de587--
`
