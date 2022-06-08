package gserv

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"go.oneofone.dev/oerrs"
	"go.oneofone.dev/otk"
)

// Common responses
var (
	RespMethodNotAllowed Response = NewJSONErrorResponse(http.StatusMethodNotAllowed).Cached()
	RespNotFound         Response = NewJSONErrorResponse(http.StatusNotFound).Cached()
	RespForbidden        Response = NewJSONErrorResponse(http.StatusForbidden).Cached()
	RespBadRequest       Response = NewJSONErrorResponse(http.StatusBadRequest).Cached()
	RespOK               Response = NewJSONResponse("OK").Cached()
	RespEmpty            Response = &CachedResponse{code: http.StatusNoContent}
	RespPlainOK          Response = &CachedResponse{code: http.StatusOK, body: []byte("OK\n")}
	RespRedirectRoot     Response = Redirect("/", false)

	// Break can be returned from a handler to break a handler chain.
	// It doesn't write anything to the connection.
	// if you reassign this, a wild animal will devour your face.
	Break Response = &CachedResponse{code: -1}
)

// Response represents a generic return type for http responses.
type Response interface {
	Status() int
	WriteToCtx(ctx *Context) error
}

func NewCachedResponse(v Response, fn func(v any) ([]byte, error)) Response {
	b, err := fn(v)
	if err != nil { // marshal function can never fail unless it's a system error
		log.Panicf("marshal error: %v", err)
	}

	return &CachedResponse{code: v.Status(), body: b}
}

type CachedResponse struct {
	ct   string
	body []byte
	code int
}

func (r *CachedResponse) Status() int { return r.code }
func (r *CachedResponse) WriteToCtx(ctx *Context) error {
	if r.ct != "" {
		ctx.SetContentType(r.ct)
	}
	if r.code != 0 {
		ctx.WriteHeader(r.code)
	}
	_, err := ctx.Write(r.body)
	return err
}

// ReadJSONResponse reads a response from an io.ReadCloser and closes the body.
// dataValue is the data type you're expecting, for example:
//
//	r, err := ReadJSONResponse(res.Body, &map[string]*Stats{})
func ReadJSONResponse(rc io.ReadCloser, dataValue any) (r *JSONResponse, err error) {
	defer rc.Close()

	r = &JSONResponse{
		Data: dataValue,
	}

	if err = json.NewDecoder(rc).Decode(r); err != nil {
		return
	}

	if r.Success {
		return
	}

	var me MultiError
	for _, v := range r.Errors {
		me.Push(&v)
	}

	if err = me.Err(); err == nil {
		err = oerrs.String(http.StatusText(r.Code))
	}

	return
}

func JSONRequest(method, url string, reqData, respData any) (err error) {
	return otk.Request(method, "", url, reqData, func(r *http.Response) error {
		_, err := ReadJSONResponse(r.Body, respData)
		return err
	})
}

// Redirect returns a redirect Response.
// if perm is false it uses http.StatusFound (302), otherwise http.StatusMovedPermanently (302)
func Redirect(url string, perm bool) Response {
	code := http.StatusFound
	if perm {
		code = http.StatusMovedPermanently
	}
	return RedirectWithCode(url, code)
}

// RedirectWithCode returns a redirect Response with the specified status code.
func RedirectWithCode(url string, code int) Response {
	return redirResp{url, code}
}

type redirResp struct {
	url  string
	code int
}

func (r redirResp) Status() int { return r.code }
func (r redirResp) WriteToCtx(ctx *Context) error {
	if r.url == "" {
		return ErrInvalidURL
	}
	http.Redirect(ctx, ctx.Req, r.url, r.code)
	return nil
}

// File returns a file response.
// example: return File("plain/html", "index.html")
func File(contentType, fp string) Response {
	return fileResp{contentType, fp}
}

type fileResp struct {
	ct string
	fp string
}

func (f fileResp) Status() int { return 0 }
func (f fileResp) WriteToCtx(ctx *Context) error {
	if f.ct != "" {
		ctx.SetContentType(f.ct)
	}
	return ctx.File(f.fp)
}

// PlainResponse returns SimpleResponse(200, contentType, val).
func PlainResponse(contentType string, val any) Response {
	return SimpleResponse(http.StatusOK, contentType, val)
}

// SimpleResponse is a QoL wrapper to return a response with the specified code and content-type.
// val can be: nil, []byte, string, io.Writer, anything else will be written with fmt.Printf("%v").
func SimpleResponse(code int, contentType string, val any) Response {
	return &simpleResp{
		ct:   contentType,
		v:    val,
		code: code,
	}
}

type simpleResp struct {
	v    any
	ct   string
	code int
}

func (r *simpleResp) Status() int { return r.code }
func (r simpleResp) WriteToCtx(ctx *Context) error {
	if r.ct != "" {
		ctx.SetContentType(r.ct)
	}

	if r.code > 0 {
		ctx.WriteHeader(r.code)
	}

	var err error
	switch v := r.v.(type) {
	case nil:
	case []byte:
		_, err = ctx.Write(v)
	case string:
		_, err = ctx.Write(otk.UnsafeBytes(v))
	case io.Reader:
		_, err = io.Copy(ctx, v)
	default:
		_, err = fmt.Fprintf(ctx, "%v", r.v)
	}
	return err
}
