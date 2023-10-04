package parsemail

import (
	"errors"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"strings"
)

type (
	Charset string
)

var (
	ErrUnsupported = errors.New("unsupported charset provided")
)

const (
	Iso88591  Charset = "iso-8859-1"
	Iso88592  Charset = "iso-8859-2"
	Iso88593  Charset = "iso-8859-3"
	Iso88594  Charset = "iso-8859-4"
	Iso88595  Charset = "iso-8859-5"
	Iso88596  Charset = "iso-8859-6"
	Iso88597  Charset = "iso-8859-7"
	Iso88598  Charset = "iso-8859-8"
	Iso88599  Charset = "iso-8859-9"
	Iso885910 Charset = "iso-8859-10"
	Iso885913 Charset = "iso-8859-13"
	Iso885914 Charset = "iso-8859-14"
	Iso885915 Charset = "iso-8859-15"
	Iso885916 Charset = "iso-8859-16"
	Utf8      Charset = "utf-8"
)

func (c Charset) String() string {
	return string(c)
}

func charsetFromParams(params map[string]string) Charset {
	var (
		charset string
		ok      bool
	)
	if charset, ok = params["charset"]; !ok {
		return Utf8
	}
	return Charset(strings.ToLower(charset))
}

func charsetDecoder(c Charset) (*encoding.Decoder, error) {
	switch c {
	case Iso88591:
		return charmap.ISO8859_1.NewDecoder(), nil
	case Iso88592:
		return charmap.ISO8859_2.NewDecoder(), nil
	case Iso88593:
		return charmap.ISO8859_3.NewDecoder(), nil
	case Iso88594:
		return charmap.ISO8859_4.NewDecoder(), nil
	case Iso88595:
		return charmap.ISO8859_5.NewDecoder(), nil
	case Iso88596:
		return charmap.ISO8859_6.NewDecoder(), nil
	case Iso88597:
		return charmap.ISO8859_7.NewDecoder(), nil
	case Iso88598:
		return charmap.ISO8859_8.NewDecoder(), nil
	case Iso88599:
		return charmap.ISO8859_9.NewDecoder(), nil
	case Iso885910:
		return charmap.ISO8859_10.NewDecoder(), nil
	case Iso885913:
		return charmap.ISO8859_13.NewDecoder(), nil
	case Iso885914:
		return charmap.ISO8859_14.NewDecoder(), nil
	case Iso885915:
		return charmap.ISO8859_15.NewDecoder(), nil
	case Iso885916:
		return charmap.ISO8859_16.NewDecoder(), nil
	default:
		return nil, ErrUnsupported
	}
}
