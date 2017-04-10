# Parsemail - simple email parsing Go library

Simple usage:

```go
var reader io.Reader // this reads an email message
email, err := parsemail.Parse(reader) // returns Email struct and error
if err != nil {
    // handle error
}

fmt.Println(email.Subject()) // parsed from header on demand
fmt.Println(email.Cc()) // parsed from header on demand
fmt.Println(email.HTMLBody) // this is not a method, parsed beforehand
```

## This library is WIP.

It is missing some tests, and needs more work. Use at your own discretion.



## TODO

- CI
- Readme with use cases
- More tests for 100% coverage
- email address type => getEmail, getName
- quoted§ text?