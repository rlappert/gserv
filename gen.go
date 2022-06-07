package gserv

import (
	"net/http"
)

type Groupie interface {
	AddRoute(method, path string, handlers ...Handler)
}

func Get[CodecT Codec, Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	handleOutOnly[CodecT](g, http.MethodGet, path, handler)
}

func JSONGet[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Get[JSONCodec](g, path, handler)
}

func MsgpGet[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Get[MsgpCodec](g, path, handler)
}

func Delete[CodecT Codec, Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	handleOutOnly[CodecT](g, http.MethodDelete, path, handler)
}

func JSONDelete[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Delete[JSONCodec](g, path, handler)
}

func MsgpDelete[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Delete[MsgpCodec](g, path, handler)
}

func Post[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	handleInOut[CodecT](g, http.MethodPost, path, handler)
}

func JSONPost[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Post[JSONCodec](g, path, handler)
}

func MsgpPost[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Post[MsgpCodec](g, path, handler)
}

func Put[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	handleInOut[CodecT](g, http.MethodPut, path, handler)
}

func JSONPut[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Put[JSONCodec](g, path, handler)
}

func MsgpPut[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Put[MsgpCodec](g, path, handler)
}

func Patch[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	handleInOut[CodecT](g, http.MethodPatch, path, handler)
}

func JSONPatch[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Patch[JSONCodec](g, path, handler)
}

func MsgpPatch[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) {
	Patch[MsgpCodec](g, path, handler)
}

func handleOutOnly[CodecT Codec, Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, method, path string, handler HandlerFn) {
	g.AddRoute(method, path, func(ctx *Context) Response {
		resp, err := handler(ctx)
		if err != nil {
			return NewErrorResponse[CodecT](http.StatusBadRequest, getError(err))
		}
		return NewResponse[CodecT](resp)
	})
}

func handleInOut[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, method, path string, handler HandlerFn) {
	var c CodecT
	g.AddRoute(method, path, func(ctx *Context) Response {
		var body Req
		if err := c.Decode(ctx.Req.Body, &body); err != nil {
			return NewErrorResponse[CodecT](http.StatusBadRequest, getError(err))
		}

		ctx.SetContentType(c.ContentType())
		resp, err := handler(ctx, body)
		if err != nil {
			return NewErrorResponse[CodecT](http.StatusBadRequest, getError(err))
		}
		return NewResponse[CodecT](resp)
	})
}
