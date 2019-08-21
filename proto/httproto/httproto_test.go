package httproto_test

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/henrylee2cn/goutil/httpbody"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/proto/httproto"
)

type Home struct {
	tp.CallCtx
}

func (h *Home) Test(arg *map[string]string) (map[string]interface{}, *tp.Status) {
	tp.Infof("peer_id: %s", h.PeekMeta("peer_id"))
	return map[string]interface{}{
		"arg": *arg,
	}, nil
}

func (h *Home) TestError(arg *map[string]string) (map[string]interface{}, *tp.Status) {
	return nil, tp.NewStatus(1, "test error", "this is test:"+string(h.PeekMeta("peer_id")))
}

func TestHTTProto(t *testing.T) {
	// Server
	srv := tp.NewPeer(tp.PeerConfig{ListenPort: 9090})
	srv.RouteCall(new(Home))
	go srv.ListenAndServe(httproto.NewHTTProtoFunc(true))
	time.Sleep(1e9)

	cli := tp.NewPeer(tp.PeerConfig{})
	sess, stat := cli.Dial(":9090", httproto.NewHTTProtoFunc())
	if !stat.OK() {
		t.Fatal(stat)
	}
	var result interface{}
	var arg = map[string]string{
		"author": "henrylee2cn",
	}

	{
		testURL := "http://localhost:9090/home/test?peer_id=110"
		stat = sess.Call(
			testURL,
			arg,
			&result,
		).Status()
		if !stat.OK() {
			t.Fatal(stat)
		}
		t.Logf("teleport client response: %v", result)

		// HTTP Client
		contentType, body, _ := httpbody.NewJSONBody(arg)
		resp, err := http.Post(testURL, contentType, body)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Logf("http client response: %s", b)
	}

	{
		testErrURL := "http://localhost:9090/home/test_error?peer_id=110"
		result = nil
		stat = sess.Call(
			testErrURL,
			arg,
			&result,
		).Status()
		if stat.OK() {
			t.Fatal("test_error expect error")
		}
		t.Logf("teleport client response: %v, %v", stat, result)

		contentType, body, _ := httpbody.NewJSONBody(arg)
		resp, err := http.Post(testErrURL, contentType, body)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		t.Logf("http client response: %s", b)
	}
}
