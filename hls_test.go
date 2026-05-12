package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestAbsoluteURL(t *testing.T) {
	base, _ := url.Parse("https://cdn.example.com/playlist/master.m3u8")
	tests := []struct {
		in, want string
	}{
		{"https://other.example.com/abs.ts", "https://other.example.com/abs.ts"},
		{"segment01.ts", "https://cdn.example.com/playlist/segment01.ts"},
		{"/keys/key.bin", "https://cdn.example.com/keys/key.bin"},
	}
	for _, tt := range tests {
		got, err := absoluteURL(base, tt.in)
		if err != nil {
			t.Fatalf("absoluteURL(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("absoluteURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestAESCBCDecryptRoundtrip(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := []byte("abcdefghijklmnop")
	plaintext := []byte("the quick brown fox jumps over the lazy dog!!!!")

	// PKCS#7-pad to block size.
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	pad := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := append([]byte{}, plaintext...)
	for i := 0; i < pad; i++ {
		padded = append(padded, byte(pad))
	}
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	got, err := aesCBCDecrypt(ciphertext, key, iv)
	if err != nil {
		t.Fatalf("aesCBCDecrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("decrypted = %q, want %q", got, plaintext)
	}
}

func TestHTTPGetBytes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			_, _ = w.Write([]byte("hello world"))
		case "/boom":
			http.Error(w, "nope", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	body, err := httpGetBytes(http.DefaultClient, srv.URL+"/ok")
	if err != nil {
		t.Fatalf("ok request: %v", err)
	}
	if string(body) != "hello world" {
		t.Errorf("body = %q, want %q", body, "hello world")
	}

	if _, err := httpGetBytes(http.DefaultClient, srv.URL+"/boom"); err == nil {
		t.Error("expected error from 500 response, got nil")
	}
}

func TestKeyCacheMemoizes(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte("0123456789abcdef"))
	}))
	defer srv.Close()

	kc := newKeyCache(http.DefaultClient)
	for i := 0; i < 5; i++ {
		got, err := kc.get(srv.URL + "/key")
		if err != nil {
			t.Fatalf("kc.get: %v", err)
		}
		if string(got) != "0123456789abcdef" {
			t.Errorf("key bytes = %q, want %q", got, "0123456789abcdef")
		}
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Errorf("server hits = %d, want 1 (cache should memoize)", got)
	}
}
