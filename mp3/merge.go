package mp3

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafov/m3u8"
	"github.com/imthaghost/scdl/decrypt"
	"github.com/imthaghost/scdl/joiner"
	"github.com/imthaghost/scdl/pool"
	"github.com/imthaghost/scdl/zhttp"
)

var (
	// ZHTTP client
	ZHTTP        *zhttp.Zhttp
	JOINER       *joiner.Joiner
	keyCache     map[string][]byte
	keyCacheLock sync.Mutex
)

func start(mpl *m3u8.MediaPlaylist) {
	p := pool.New(10, download)

	go func() {
		var count = int(mpl.Count())
		for i := 0; i < count; i++ {
			p.Push([]interface{}{i, mpl.Segments[i], mpl.Key})
		}
		p.CloseQueue()
	}()

	go p.Run()
}

func parseM3u8(m3u8Url string) (*m3u8.MediaPlaylist, error) {
	statusCode, data, err := ZHTTP.Get(m3u8Url)
	if err != nil {
		return nil, err
	}

	if statusCode/100 != 2 || len(data) == 0 {
		return nil, errors.New("download m3u8 file failed, http code: " + strconv.Itoa(statusCode))
	}

	playlist, listType, err := m3u8.Decode(*bytes.NewBuffer(data), true)
	if err != nil {
		return nil, err
	}

	if listType == m3u8.MEDIA {
		obj, _ := url.Parse(m3u8Url)
		mpl := playlist.(*m3u8.MediaPlaylist)

		if mpl.Key != nil && mpl.Key.URI != "" {
			uri, err := formatURI(obj, mpl.Key.URI)
			if err != nil {
				return nil, err
			}
			mpl.Key.URI = uri
		}

		count := int(mpl.Count())
		for i := 0; i < count; i++ {
			segment := mpl.Segments[i]

			uri, err := formatURI(obj, segment.URI)
			if err != nil {
				return nil, err
			}
			segment.URI = uri

			if segment.Key != nil && segment.Key.URI != "" {
				uri, err := formatURI(obj, segment.Key.URI)
				if err != nil {
					return nil, err
				}
				segment.Key.URI = uri
			}

			mpl.Segments[i] = segment
		}

		return mpl, nil
	}

	return nil, errors.New("Unsupport m3u8 type")
}

func getKey(url string) ([]byte, error) {
	keyCacheLock.Lock()
	defer keyCacheLock.Unlock()

	key := keyCache[url]
	if key != nil {
		return key, nil
	}

	statusCode, key, err := ZHTTP.Get(url)
	if err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return nil, errors.New("body is empty, http code: " + strconv.Itoa(statusCode))
	}

	keyCache[url] = key

	return key, nil
}

func download(in interface{}) {
	params := in.([]interface{})
	id := params[0].(int)
	segment := params[1].(*m3u8.MediaSegment)
	globalKey := params[2].(*m3u8.Key)

	statusCode, data, err := ZHTTP.Get(segment.URI)
	if err != nil {
		log.Fatalln("[-] Download failed:", err)
	}

	if len(data) == 0 {
		log.Fatalln("[-] Download failed: body is empty, http code:", statusCode)
	}

	var keyURL, ivStr string
	if segment.Key != nil && segment.Key.URI != "" {
		keyURL = segment.Key.URI
		ivStr = segment.Key.IV
	} else if globalKey != nil && globalKey.URI != "" {
		keyURL = globalKey.URI
		ivStr = globalKey.IV
	}

	if keyURL != "" {
		var key, iv []byte
		key, err = getKey(keyURL)
		if err != nil {
			log.Fatalln("[-] Download key failed:", keyURL, err)
		}

		if ivStr != "" {
			iv, err = hex.DecodeString(strings.TrimPrefix(ivStr, "0x"))
			if err != nil {
				log.Fatalln("[-] Decode iv failed:", err)
			}
		} else {
			iv = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(id)}
		}

		data, err = decrypt.Decrypt(data, key, iv)
		if err != nil {
			log.Fatalln("[-] Decrypt failed:", err)
		}
	}

	// log.Println("[+] Download succed:", segment.URI)

	JOINER.Join(id, data)
}

func formatURI(base *url.URL, u string) (string, error) {
	if strings.HasPrefix(u, "http") {
		return u, nil
	}

	obj, err := base.Parse(u)
	if err != nil {
		return "", err
	}

	return obj.String(), nil
}

func filename(u string) string {
	obj, _ := url.Parse(u)
	_, filename := filepath.Split(obj.Path)
	return filename
}

// Merge ...
func Merge(url string, songname string) {

	keyCache = map[string][]byte{}

	var err error
	ZHTTP, err = zhttp.New(time.Second*30, "")
	if err != nil {
		log.Fatalln("[-] Init failed:", err)
	}

	mpl, err := parseM3u8(url)
	if err != nil {
		log.Fatalln("[-] Parse m3u8 file failed:", err)
	} else {
		log.Println("[+] Parse m3u8 file succed")
	}

	outFile := songname + ".mp3"

	JOINER, err = joiner.New(outFile)
	if err != nil {
		log.Fatalln("[-] Open file failed:", err)
	} else {
		log.Println("[+] Will save to", JOINER.Name())
	}

	if mpl.Count() > 0 {
		log.Println("[+] Total", mpl.Count(), "files to download")

		start(mpl)

		err = JOINER.Run(int(mpl.Count()))
		if err != nil {
			log.Fatalln("[-] Write to file failed:", err)
		}
		log.Println("[+] Download succed, saved to", JOINER.Name())
	}
}
