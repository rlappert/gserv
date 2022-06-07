package gserv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vmihailenco/msgpack/v5"
	"go.oneofone.dev/otk"
)

// Common mime-types
const (
	MimeJSON       = "application/json"
	MimeEvent      = "text/event-stream"
	MimeMsgPack    = "application/msgpack"
	MimeXML        = "application/xml"
	MimeJavascript = "application/javascript"
	MimeHTML       = "text/html"
	MimePlain      = "text/plain"
	MimeBinary     = "application/octet-stream"
)

var (
	_ Codec = (*PlainTextCodec)(nil)
	_ Codec = (*JSONCodec)(nil)
	_ Codec = (*MsgpCodec)(nil)
	_ Codec = (*MixedCodec[JSONCodec, MsgpCodec])(nil)
)

type Encoder interface {
	Encode(v any) error
}

type Decoder interface {
	Decode(v any) error
}

type Codec interface {
	ContentType() string
	Decode(r io.Reader, body any) error
	Encode(w io.Writer, v any, err error) error
}

type PlainTextCodec struct{}

func (PlainTextCodec) ContentType() string { return "" }

func (PlainTextCodec) Decode(r io.Reader, out any) error {
	b, err := io.ReadAll(r)
	switch out := out.(type) {
	case *string:
		*out = string(b)
	case *[]byte:
		*out = b
	case io.Writer:
		out.Write(b)
	default:
		return fmt.Errorf("%T is not a valid type for PlainTextCodec", out)
	}
	return err
}

func (PlainTextCodec) Encode(w io.Writer, v any, err error) (err2 error) {
	if err != nil {
		err := getError(err)
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, err.Error(), err.Status())
		} else {
			w.Write(otk.UnsafeBytes(err.Error()))
		}
		return nil
	}

	switch v := v.(type) {
	case string:
		_, err2 = io.WriteString(w, v)
	case []byte:
		_, err2 = w.Write(v)
	case io.Reader:
		_, err2 = io.Copy(w, v)
	default:
		return fmt.Errorf("%T is not a valid type for PlainTextCodec", v)
	}
	return
}

type JSONCodec struct{}

func (JSONCodec) ContentType() string { return MimeJSON }

func (JSONCodec) Decode(r io.Reader, out any) error {
	return json.NewDecoder(r).Decode(&out)
}

func (JSONCodec) Encode(w io.Writer, v any, err error) error {
	enc := json.NewEncoder(w)
	if err != nil {
		err := getError(err)
		if rw, ok := w.(http.ResponseWriter); ok {
			rw.WriteHeader(err.Status())
		}
		return enc.Encode(err)
	}
	return enc.Encode(v)
}

type MsgpCodec struct{}

func (MsgpCodec) ContentType() string { return MimeJSON }

func (MsgpCodec) Decode(r io.Reader, out any) error {
	dec := GetMsgpDecoder(r)
	defer msgpack.PutDecoder(dec)

	return dec.Decode(&out)
}

func (c MsgpCodec) Encode(w io.Writer, v any, err error) error {
	enc := GetMsgpEncoder(w)
	defer msgpack.PutEncoder(enc)
	if err != nil {
		err := getError(err)
		if rw, ok := w.(http.ResponseWriter); ok {
			rw.WriteHeader(err.Status())
		}
		return enc.Encode(err)
	}
	return enc.Encode(v)
}

type MixedCodec[Dec, Enc Codec] struct {
	dec Dec
	enc Enc
}

func (m MixedCodec[Dec, Enc]) ContentType() string { return m.enc.ContentType() }

func (m MixedCodec[Dec, Enc]) Decode(r io.Reader, out any) error {
	return m.dec.Decode(r, out)
}

func (m MixedCodec[Dec, Enc]) Encode(w io.Writer, v any, err error) error {
	return m.enc.Encode(w, v, err)
}

func getError(err error) HTTPError {
	if err, ok := err.(HTTPError); ok {
		return err
	}
	return &Error{Code: http.StatusBadRequest, Message: err.Error()}
}

func GetMsgpEncoder(w io.Writer) *msgpack.Encoder {
	enc := msgpack.GetEncoder()
	enc.Reset(w)
	enc.SetCustomStructTag("json")
	enc.UseCompactFloats(true)
	enc.UseCompactFloats(true)
	return enc
}

func GetMsgpDecoder(r io.Reader) *msgpack.Decoder {
	dec := msgpack.GetDecoder()
	dec.Reset(r)
	dec.SetCustomStructTag("json")
	return dec
}
