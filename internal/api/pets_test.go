package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"genericagent-admin-go/internal/config"
)

func newPetsTestServer(t *testing.T) *Server {
	t.Helper()
	root := t.TempDir()
	return New(&config.Store{Root: root, Cfg: config.Default()}, nil, nil, fstest.MapFS{
		"ga-admin-pets/pets.json": &fstest.MapFile{Data: []byte(`{
			"runtime":"ga-admin-pets/v1",
			"description":"test pets",
			"pets":[
				{"id":"alpha","name":"Alpha","manifest":"ga-admin-pets/alpha/manifest.json"},
				{"id":"beta","name":"Beta","manifest":"ga-admin-pets/beta/manifest.json"}
			]
		}`)},
		"ga-admin-pets/alpha/manifest.json": &fstest.MapFile{Data: []byte(`{
			"display_name":"Alpha Manifest",
			"description":"alpha pet",
			"asset_base":"ga-admin-pets/alpha",
			"spritesheet":"sprite.png",
			"atlas":{"columns":4,"rows":9,"frame_width":128,"frame_height":128}
		}`)},
		"ga-admin-pets/beta/manifest.json": &fstest.MapFile{Data: []byte(`{
			"display_name":"Beta Manifest",
			"description":"beta pet",
			"asset_base":"ga-admin-pets/beta",
			"spritesheet":"sprite.png",
			"atlas":{"columns":4,"rows":9,"frame_width":128,"frame_height":128}
		}`)},
	})
}

func TestPetsActiveFallsBackToFirstCatalogPet(t *testing.T) {
	s := newPetsTestServer(t)
	if got := s.ActivePetID(); got != "alpha" {
		t.Fatalf("ActivePetID=%q want alpha", got)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/pets/active", nil)
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200 body=%s", rr.Code, rr.Body.String())
	}
	var got struct {
		ActivePetID string `json:"active_pet_id"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("response json: %v", err)
	}
	if got.ActivePetID != "alpha" {
		t.Fatalf("active_pet_id=%q want alpha", got.ActivePetID)
	}
}

func TestPetsActivePostPersistsAndSwitches(t *testing.T) {
	s := newPetsTestServer(t)
	var switched []string
	s.PetSwitch = func(id string) error {
		switched = append(switched, id)
		return nil
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/pets/active", strings.NewReader(`{"pet_id":"beta"}`))
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want 200 body=%s", rr.Code, rr.Body.String())
	}
	if got := s.ActivePetID(); got != "beta" {
		t.Fatalf("ActivePetID after post=%q want beta", got)
	}
	if len(switched) != 1 || switched[0] != "beta" {
		t.Fatalf("switched=%v want [beta]", switched)
	}

	data, err := os.ReadFile(filepath.Join(s.CfgStore.Root, "pets.local.json"))
	if err != nil {
		t.Fatalf("read pets.local.json: %v", err)
	}
	var st petsState
	if err := json.Unmarshal(data, &st); err != nil {
		t.Fatalf("state json: %v", err)
	}
	if st.ActivePetID != "beta" {
		t.Fatalf("persisted active=%q want beta", st.ActivePetID)
	}
}

func TestPetsActivePostRejectsUnknownPet(t *testing.T) {
	s := newPetsTestServer(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/pets/active", strings.NewReader(`{"pet_id":"missing"}`))
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404 body=%s", rr.Code, rr.Body.String())
	}
	if _, err := os.Stat(filepath.Join(s.CfgStore.Root, "pets.local.json")); !os.IsNotExist(err) {
		t.Fatalf("pets.local.json err=%v want not exist", err)
	}
}
