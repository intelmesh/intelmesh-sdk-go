package testkit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	sdk "github.com/intelmesh/intelmesh-sdk-go"
	"github.com/intelmesh/intelmesh-sdk-go/provision"
	"github.com/intelmesh/intelmesh-sdk-go/testkit"
)

func respond(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	envelope := sdk.Envelope[json.RawMessage]{}
	raw, _ := json.Marshal(data)
	envelope.Data = raw
	_ = json.NewEncoder(w).Encode(envelope)
}

func TestHarness_CreatesAndDeletesEphemeralKey(t *testing.T) {
	t.Parallel()

	var created atomic.Int32
	var deleted atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/api-keys":
			created.Add(1)
			respond(w, map[string]any{
				"api_key": map[string]any{
					"id":          "key-1",
					"name":        "testkit",
					"permissions": []string{"events:write"},
					"enabled":     true,
					"created_at":  "2025-01-01T00:00:00Z",
				},
				"plain_key": "sk-test-ephemeral",
			})

		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/api-keys/"):
			deleted.Add(1)
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Use a sub-test so Cleanup runs before our assertions.
	t.Run("lifecycle", func(t *testing.T) {
		h := testkit.New(t, testkit.Config{
			BaseURL:  srv.URL,
			AdminKey: "admin-key",
		})

		if h.Client() == nil {
			t.Fatal("expected non-nil client")
		}

		if created.Load() != 1 {
			t.Errorf("expected 1 key creation, got %d", created.Load())
		}
	})

	// After sub-test cleanup, the key should have been deleted.
	if deleted.Load() != 1 {
		t.Errorf("expected 1 key deletion, got %d", deleted.Load())
	}
}

func TestEventAssertion_ExpectDecision(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/api-keys":
			respond(w, map[string]any{
				"api_key": map[string]any{
					"id": "key-1", "name": "testkit",
					"permissions": []string{}, "enabled": true,
					"created_at": "2025-01-01T00:00:00Z",
				},
				"plain_key": "sk-test",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/events":
			respond(w, map[string]any{
				"event_id": "evt-1",
				"decision": map[string]any{
					"action":   "block",
					"severity": "critical",
				},
				"transient_score": 42,
				"duration_ms":     5,
			})

		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Run("assert", func(t *testing.T) {
		h := testkit.New(t, testkit.Config{
			BaseURL:  srv.URL,
			AdminKey: "admin-key",
		})

		h.Send("transaction.pix", map[string]any{"amount": 100}).
			ExpectDecision("block", "critical").
			ExpectScore(42).
			Then()
	})
}

func TestEventAssertion_ExpectNoDecision(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/api-keys":
			respond(w, map[string]any{
				"api_key": map[string]any{
					"id": "key-1", "name": "testkit",
					"permissions": []string{}, "enabled": true,
					"created_at": "2025-01-01T00:00:00Z",
				},
				"plain_key": "sk-test",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/events":
			respond(w, map[string]any{
				"event_id":        "evt-2",
				"transient_score": 0,
				"duration_ms":     3,
			})

		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Run("assert", func(t *testing.T) {
		h := testkit.New(t, testkit.Config{
			BaseURL:  srv.URL,
			AdminKey: "admin-key",
		})

		h.Send("login.success", map[string]any{"user": "test"}).
			ExpectNoDecision().
			ExpectScore(0).
			Then()
	})
}

func TestHarness_SendSimulate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/api-keys":
			respond(w, map[string]any{
				"api_key": map[string]any{
					"id": "key-1", "name": "testkit",
					"permissions": []string{}, "enabled": true,
					"created_at": "2025-01-01T00:00:00Z",
				},
				"plain_key": "sk-test",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/events/simulate":
			respond(w, map[string]any{
				"event_id":        "evt-sim",
				"transient_score": 10,
				"duration_ms":     2,
			})

		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Run("assert", func(t *testing.T) {
		h := testkit.New(t, testkit.Config{
			BaseURL:  srv.URL,
			AdminKey: "admin-key",
		})

		h.SendSimulate("transaction.pix", map[string]any{"amount": 200}).
			ExpectNoDecision().
			ExpectScore(10).
			Then()
	})
}

func TestHarness_SendIngestOnly(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/api-keys":
			respond(w, map[string]any{
				"api_key": map[string]any{
					"id": "key-1", "name": "testkit",
					"permissions": []string{}, "enabled": true,
					"created_at": "2025-01-01T00:00:00Z",
				},
				"plain_key": "sk-test",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/events":
			// IngestOnly adds ?evaluate=false query param
			if r.URL.Query().Get("evaluate") == "false" {
				respond(w, map[string]any{
					"event_id":        "evt-io",
					"transient_score": 0,
					"duration_ms":     1,
				})
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}

		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Run("assert", func(t *testing.T) {
		h := testkit.New(t, testkit.Config{
			BaseURL:  srv.URL,
			AdminKey: "admin-key",
		})

		h.SendIngestOnly("event.raw", map[string]any{"data": "x"}).
			ExpectNoDecision().
			ExpectScore(0).
			Then()
	})
}

func TestHarness_ProvisionAndVerifyList(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	listItems := make(map[string][]string)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/api-keys":
			respond(w, map[string]any{
				"api_key": map[string]any{
					"id": "key-1", "name": "testkit",
					"permissions": []string{}, "enabled": true,
					"created_at": "2025-01-01T00:00:00Z",
				},
				"plain_key": "sk-test",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/phases":
			respond(w, map[string]any{
				"id": "phase-1", "name": "screening", "position": 1,
				"created_at": "2025-01-01T00:00:00Z",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/lists":
			respond(w, map[string]any{
				"id": "list-1", "name": "blocklist",
				"created_at": "2025-01-01T00:00:00Z",
				"updated_at": "2025-01-01T00:00:00Z",
			})

		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/rules":
			respond(w, map[string]any{
				"id": "rule-1", "name": "r", "phase_id": "phase-1",
				"priority": 1, "expression": "true",
				"actions": map[string]any{}, "enabled": true,
				"created_at": "2025-01-01T00:00:00Z",
				"updated_at": "2025-01-01T00:00:00Z",
			})

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/lists/") &&
			strings.HasSuffix(r.URL.Path, "/items"):
			listID := strings.TrimPrefix(r.URL.Path, "/api/v1/lists/")
			listID = strings.TrimSuffix(listID, "/items")
			mu.Lock()
			items := listItems[listID]
			mu.Unlock()

			apiItems := make([]map[string]any, 0, len(items))
			for i, v := range items {
				apiItems = append(apiItems, map[string]any{
					"id": "item-" + v, "list_id": listID, "value": v,
					"created_at": "2025-01-01T00:00:00Z",
				})
				_ = i
			}
			respond(w, map[string]any{
				"items": apiItems,
				"count": len(apiItems),
			})

		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Run("verify", func(t *testing.T) {
		h := testkit.New(t, testkit.Config{
			BaseURL:  srv.URL,
			AdminKey: "admin-key",
		})

		p := provision.New(h.Client()).
			Phase("screening", 1).
			List("blocklist").
			Rule("block-rule").InPhase("screening").Priority(1).
			When("true").Decide("block", "critical").Halt().Done()

		h.Provision(p)

		// Simulate the list having an item.
		mu.Lock()
		listItems["list-1"] = []string{"c-002"}
		mu.Unlock()

		h.VerifyListContains("blocklist", "c-002")
		h.VerifyListNotContains("blocklist", "c-999")
	})
}

func TestHarness_WaitForProjectors(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/api-keys":
			respond(w, map[string]any{
				"api_key": map[string]any{
					"id": "key-1", "name": "testkit",
					"permissions": []string{}, "enabled": true,
					"created_at": "2025-01-01T00:00:00Z",
				},
				"plain_key": "sk-test",
			})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	t.Run("wait", func(t *testing.T) {
		h := testkit.New(t, testkit.Config{
			BaseURL:  srv.URL,
			AdminKey: "admin-key",
		})

		// WaitForProjectors should return the harness for chaining.
		result := h.WaitForProjectors()
		if result != h {
			t.Error("WaitForProjectors should return the harness")
		}
	})
}
