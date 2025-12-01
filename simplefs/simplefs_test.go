package simplefs_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/darkweak/storages/core"
	"github.com/darkweak/storages/simplefs"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

func newDefaultStore(t *testing.T) core.Storer {
	t.Helper()
	s, err := simplefs.Factory(core.CacheProvider{}, zap.NewNop().Sugar(), 0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSetAndRetrieveValueFromStore(t *testing.T) {
	store := newDefaultStore(t)

	err := store.Set("key", []byte("123"), time.Duration(20)*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	got := store.Get("key")
	if !cmp.Equal([]byte("123"), got) {
		t.Error("stored and retrieved values don't match")
	}
}

func TestGetReturnsNilValueForNotExistingKey(t *testing.T) {
	store := newDefaultStore(t)

	got := store.Get("not-existing-key")
	if got != nil {
		t.Errorf("want nil, got %v", got)
	}
}

func TestRetrieveCachedDataFromStoreBeforeExpirationTime(t *testing.T) {
	store := newDefaultStore(t)

	err := store.Set("key", []byte("123"), time.Duration(20)*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	got := store.Get("key")
	if !cmp.Equal([]byte("123"), got) {
		t.Error("stored and retrieved values don't match")
	}
}

func TestSimplefs_SetRequestInCache_Negative_TTL(t *testing.T) {
	store := newDefaultStore(t)

	err := store.Set("key", []byte("123"), -1)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	err = store.Set("key", []byte("456"), time.Duration(20)*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	got := store.Get("key")
	if !cmp.Equal([]byte("456"), got) {
		t.Error("stored and retrieved values don't match")
	}
}

func TestSimplefs_DeleteRequestInCache(t *testing.T) {
	store := newDefaultStore(t)

	k := "key"
	store.Delete(k)
	time.Sleep(1 * time.Second)

	got := store.Get(k)
	if got != nil {
		t.Errorf("want nil value for not exising key, got %v", got)
	}
}

func TestInitializeDefaultStore(t *testing.T) {
	store := newDefaultStore(t)
	err := store.Init()
	if err != nil {
		t.Error(err)
	}
}

func TestSimplefs_EvictAfterXSeconds(t *testing.T) {
	store := newDefaultStore(t)
	err := store.Init()
	if err != nil {
		t.Fatal(t)
	}

	v := []byte("base value")
	for i := range 10 {
		k := fmt.Sprintf("Test_%d", i)
		err := store.SetMultiLevel(k, k, v, http.Header{}, "", time.Second, k)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	k0 := "Test_0"
	got := store.Get(k0)
	if got != nil {
		t.Errorf("want nil for key %s, got %v", k0, got)
	}

	k9 := "Test_9"
	got = store.Get(k9)
	if got == nil {
		t.Errorf("want %v for key %s, got nil", v, k9)
	}
	time.Sleep(3 * time.Second)
}
