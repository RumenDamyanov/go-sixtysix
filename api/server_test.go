package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.rumenx.com/sixtysix/api"
	"go.rumenx.com/sixtysix/engine"
	"go.rumenx.com/sixtysix/sixtysix"
	"go.rumenx.com/sixtysix/store"
)

func TestServer_Flow(t *testing.T) {
	mem := store.NewMemory()
	e := engine.New(mem)
	e.Register(sixtysix.Game{})

	srv := api.New(e)

	// list games
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games", nil)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte("sixtysix")) {
		t.Fatalf("games: %d %s", rr.Code, rr.Body.String())
	}

	// create session
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/sessions?game=sixtysix", nil)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rr.Code, rr.Body.String())
	}

	var sess map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &sess); err != nil {
		t.Fatalf("json: %v", err)
	}
	id := sess["id"].(string)

	// apply a simple no-payload action
	body := bytes.NewBufferString(`{"type":"closeStock"}`)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/sessions/"+id, body)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("apply: %d %s", rr.Code, rr.Body.String())
	}

	// get session
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/sessions/"+id, nil)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte("version")) {
		t.Fatalf("get: %d %s", rr.Code, rr.Body.String())
	}

	// delete session
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/sessions/"+id, nil)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent || rr.Body.Len() != 0 {
		t.Fatalf("delete: %d %s", rr.Code, rr.Body.String())
	}
}

func readAll(rc io.ReadCloser) []byte { b, _ := io.ReadAll(rc); return b }
