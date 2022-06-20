package gserv

import (
	"net/http"
)

type Groupie interface {
	AddRoute(method, path string, handlers ...Handler) Route
}

func Get[CodecT Codec, Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return handleOutOnly[CodecT](g, http.MethodGet, path, handler)
}

func JSONGet[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Get[JSONCodec](g, path, handler)
}

func MsgpGet[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Get[MsgpCodec](g, path, handler)
}

func Delete[CodecT Codec, Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return handleOutOnly[CodecT](g, http.MethodDelete, path, handler)
}

func JSONDelete[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Delete[JSONCodec](g, path, handler)
}

func MsgpDelete[Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Delete[MsgpCodec](g, path, handler)
}

func Post[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return handleInOut[CodecT](g, http.MethodPost, path, handler)
}

func JSONPost[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Post[JSONCodec](g, path, handler)
}

func MsgpPost[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Post[MsgpCodec](g, path, handler)
}

func Put[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return handleInOut[CodecT](g, http.MethodPut, path, handler)
}

func JSONPut[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Put[JSONCodec](g, path, handler)
}

func MsgpPut[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Put[MsgpCodec](g, path, handler)
}

func Patch[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return handleInOut[CodecT](g, http.MethodPatch, path, handler)
}

func JSONPatch[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Patch[JSONCodec](g, path, handler)
}

func MsgpPatch[Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, path string, handler HandlerFn) Route {
	return Patch[MsgpCodec](g, path, handler)
}

func handleOutOnly[CodecT Codec, Resp any, HandlerFn func(ctx *Context) (resp Resp, err error)](g Groupie, method, path string, handler HandlerFn) Route {
	return g.AddRoute(method, path, func(ctx *Context) Response {
		resp, err := handler(ctx)
		if err != nil {
			return NewErrorResponse[CodecT](http.StatusBadRequest, getError(err))
		}
		return NewResponse[CodecT](resp)
	})
}

func handleInOut[CodecT Codec, Req, Resp any, HandlerFn func(ctx *Context, reqBody Req) (resp Resp, err error)](g Groupie, method, path string, handler HandlerFn) Route {
	var c CodecT
	return g.AddRoute(method, path, func(ctx *Context) Response {
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
