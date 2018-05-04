# Parsemail - simple email parsing Go library

[![Build Status](https://circleci.com/gh/DusanKasan/parsemail.svg?style=shield&circle-token=:circle-token)](https://circleci.com/gh/DusanKasan/parsemail) [![Coverage Status](https://coveralls.io/repos/github/DusanKasan/Parsemail/badge.svg?branch=master)](https://coveralls.io/github/DusanKasan/Parsemail?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/DusanKasan/parsemail)](https://goreportcard.com/report/github.com/DusanKasan/parsemail)

This library allows for parsing an email message into a more convenient form than the `net/mail` provides. Where the `net/mail` just gives you a map of header fields and a `io.Reader` of its body, Parsemail allows access to all the standard header fields set in [RFC5322](https://tools.ietf.org/html/rfc5322), html/text body as well as attachements/embedded content as binary streams with metadata.

## Simple usage

You just parse a io.Reader that holds the email data. The returned Email struct contains all the standard email information/headers  as public fields.

```go
var reader io.Reader // this reads an email message
email, err := parsemail.Parse(reader) // returns Email struct and error
if err != nil {
    // handle error
}

fmt.Println(email.Subject)
fmt.Println(email.From)
fmt.Println(email.To)
fmt.Println(email.HTMLBody)
```

## Retrieving attachments

Attachments are a easily accessible as `Attachment` type, containing their mime type, filename and data stream.

```go
var reader io.Reader
email, err := parsemail.Parse(reader)
if err != nil {
    // handle error
}

for _, a := range(email.Attachments) {
    fmt.Println(a.Filename)
    fmt.Println(a.ContentType)
    //and read a.Data
}
```

## Retrieving embedded files

You can access embedded files in the same way you can access attachments. They contain the mime type, data stream and content id that is used to reference them through the email.

```go
var reader io.Reader
email, err := parsemail.Parse(reader)
if err != nil {
    // handle error
}

for _, a := range(email.EmbeddedFiles) {
    fmt.Println(a.CID)
    fmt.Println(a.ContentType)
    //and read a.Data
}
```