package gserv

import (
	"bytes"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/gorilla/securecookie"
)

func TestSecureCookie(t *testing.T) {
	srv := newServerAndWait(t, "")
	defer srv.Shutdown(0)

	srv.Use(SecureCookie(bytes.Repeat([]byte("1"), 32), securecookie.GenerateRandomKey(32)))

	JSONGet(&srv.Group, "/", func(ctx *Context) (any, error) {
		ctx.SetCookie("cooookie", M{"stuff": "and things"}, "", false, time.Hour)
		return nil, nil
	})

	JSONGet(&srv.Group, "/cookie", func(ctx *Context) (M, error) {
		var m M
		ctx.GetCookieValue("cooookie", &m)
		return m, nil
	})

	addr := srv.Addrs()[0]

	var cli http.Client
	cli.Jar, _ = cookiejar.New(nil)

	resp, err := cli.Get("http://" + addr + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	cs := resp.Cookies()

	if len(cs) != 1 {
		t.Fatal("couldn't find the cookie :(")
	}

	resp, err = cli.Get("http://" + addr + "/cookie")
	if err != nil {
		t.Fatal(err)
	}

	var respValue M
	if _, err = ReadJSONResponse(resp.Body, &respValue); err != nil {
		t.Fatal(err)
	}

	if respValue["stuff"] != "and things" {
		t.Fatalf("unexpected response: %#+v", respValue)
	}
}
