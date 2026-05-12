package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/grafov/m3u8"
	"golang.org/x/sync/errgroup"
)

const (
	hlsHTTPTimeout = 30 * time.Second
	hlsConcurrency = 100
)

var (
	red   = color.New(color.FgRed).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
)

// downloadHLS fetches every segment of the HLS playlist at playlistURL,
// decrypts them if AES-128 keys are present, and writes the assembled bytes
// to outPath. Uses the Soundcloud client's Transport (proxy-aware) but with
// a per-request timeout suitable for short segment fetches.
//
// If any segment fails (fetch, decrypt, or context cancellation), in-flight
// downloads are cancelled and the output file is not written, so we never
// produce a silently-corrupt mp3.
func (s *Soundcloud) downloadHLS(ctx context.Context, playlistURL, outPath string) error {
	client := *s.Client
	client.Timeout = hlsHTTPTimeout
	return downloadHLSWith(ctx, &client, playlistURL, outPath)
}

// downloadHLSWith is the actual implementation, parameterized over the
// http.Client so tests can swap in a mock.
func downloadHLSWith(ctx context.Context, client *http.Client, playlistURL, outPath string) error {
	mpl, err := fetchMediaPlaylist(ctx, client, playlistURL)
	if err != nil {
		return fmt.Errorf("parse m3u8: %w", err)
	}
	fmt.Printf("%s Parsed m3u8 playlist\n", green("[+]"))
	fmt.Printf("%s Will save to %s\n", green("[+]"), outPath)
	fmt.Printf("%s Total %d segments to download\n", green("[+]"), mpl.Count())

	count := int(mpl.Count())
	segments := make([][]byte, count)
	keys := newKeyCache(client)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(hlsConcurrency)
	for i := 0; i < count; i++ {
		idx := i
		g.Go(func() error {
			seg := mpl.Segments[idx]
			data, err := fetchSegment(gctx, client, seg)
			if err != nil {
				return fmt.Errorf("segment %d fetch: %w", idx, err)
			}
			data, err = maybeDecrypt(gctx, data, idx, seg, mpl.Key, keys)
			if err != nil {
				return fmt.Errorf("segment %d decrypt: %w", idx, err)
			}
			segments[idx] = data
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	out, err := writeSegments(outPath, segments)
	if err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	fmt.Printf("%s Download succeeded, saved to %s\n", green("[+]"), out)
	return nil
}

// fetchMediaPlaylist downloads and decodes the m3u8 at u, resolving any
// relative key and segment URIs against the playlist URL.
func fetchMediaPlaylist(ctx context.Context, client *http.Client, u string) (*m3u8.MediaPlaylist, error) {
	body, err := httpGetBytes(ctx, client, u)
	if err != nil {
		return nil, err
	}
	playlist, listType, err := m3u8.Decode(*bytes.NewBuffer(body), true)
	if err != nil {
		return nil, err
	}
	if listType != m3u8.MEDIA {
		return nil, errors.New("unsupported m3u8 type")
	}

	base, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	mpl := playlist.(*m3u8.MediaPlaylist)

	if mpl.Key != nil && mpl.Key.URI != "" {
		if mpl.Key.URI, err = absoluteURL(base, mpl.Key.URI); err != nil {
			return nil, err
		}
	}
	for i := 0; i < int(mpl.Count()); i++ {
		seg := mpl.Segments[i]
		if seg.URI, err = absoluteURL(base, seg.URI); err != nil {
			return nil, err
		}
		if seg.Key != nil && seg.Key.URI != "" {
			if seg.Key.URI, err = absoluteURL(base, seg.Key.URI); err != nil {
				return nil, err
			}
		}
		mpl.Segments[i] = seg
	}
	return mpl, nil
}

func fetchSegment(ctx context.Context, client *http.Client, seg *m3u8.MediaSegment) ([]byte, error) {
	data, err := httpGetBytes(ctx, client, seg.URI)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("empty segment body")
	}
	return data, nil
}

// maybeDecrypt applies AES-CBC decryption when the segment (or playlist) has a
// key URI. iv defaults to the segment index when not declared.
func maybeDecrypt(ctx context.Context, data []byte, idx int, seg *m3u8.MediaSegment, globalKey *m3u8.Key, keys *keyCache) ([]byte, error) {
	keyURL, ivStr := "", ""
	switch {
	case seg.Key != nil && seg.Key.URI != "":
		keyURL, ivStr = seg.Key.URI, seg.Key.IV
	case globalKey != nil && globalKey.URI != "":
		keyURL, ivStr = globalKey.URI, globalKey.IV
	default:
		return data, nil
	}

	key, err := keys.get(ctx, keyURL)
	if err != nil {
		return nil, fmt.Errorf("fetch key %s: %w", keyURL, err)
	}

	var iv []byte
	if ivStr != "" {
		iv, err = hex.DecodeString(strings.TrimPrefix(ivStr, "0x"))
		if err != nil {
			return nil, fmt.Errorf("decode iv: %w", err)
		}
	} else {
		iv = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(idx)}
	}
	return aesCBCDecrypt(data, key, iv)
}

func writeSegments(path string, segments [][]byte) (string, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	for _, seg := range segments {
		if _, err := f.Write(seg); err != nil {
			return "", err
		}
	}
	return path, nil
}

// absoluteURL resolves a possibly-relative URI against a base URL.
func absoluteURL(base *url.URL, u string) (string, error) {
	if strings.HasPrefix(u, "http") {
		return u, nil
	}
	parsed, err := base.Parse(u)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func httpGetBytes(ctx context.Context, client *http.Client, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, errors.New("http status " + strconv.Itoa(resp.StatusCode))
	}
	return body, nil
}

// keyCache memoizes AES key bytes by URL so we only fetch each key once.
type keyCache struct {
	client *http.Client
	mu     sync.Mutex
	cache  map[string][]byte
}

func newKeyCache(client *http.Client) *keyCache {
	return &keyCache{client: client, cache: map[string][]byte{}}
}

func (k *keyCache) get(ctx context.Context, u string) ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if v, ok := k.cache[u]; ok {
		return v, nil
	}
	key, err := httpGetBytes(ctx, k.client, u)
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, errors.New("empty key body")
	}
	k.cache[u] = key
	return key, nil
}

// aesCBCDecrypt decrypts AES-128-CBC ciphertext in place and strips
// PKCS#7-style trailing padding indicated by the final byte.
func aesCBCDecrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(data, data)
	n := len(data)
	if n == 0 {
		return data, nil
	}
	pad := int(data[n-1])
	if pad > n {
		return data, nil
	}
	return data[:n-pad], nil
}
