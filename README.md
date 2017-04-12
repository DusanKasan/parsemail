# Parsemail - simple email parsing Go library

This library allows for parsing an email message into a more convenient form than the `net/mail` provides. Where the `net/mail` just gives you a map of header fields and a `io.Reader` of its body, Parsemail allows access to all the standard header fields set in [RFC5322](https://tools.ietf.org/html/rfc5322), html/text body as well as attachements/embedded as binary streams with metadata.

## Simple usage

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

## This library is WIP.

It is missing some tests, and needs more work. Use at your own discretion.

## TODO

- CI
- Readme with use cases
- More tests for 100% coverage
- quoted text?