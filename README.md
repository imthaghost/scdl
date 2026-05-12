<p align="center">
    <img alt="gopher" src="docs/media/pods.png">
</p>
<p align="center">
A fast SoundCloud downloader written in Go. Give it a track or playlist URL, get back <code>.mp3</code> files with embedded cover art.
</p>
<p align="center">
   <a href="https://github.com/imthaghost/scdl/actions/workflows/ci.yml"><img src="https://github.com/imthaghost/scdl/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
   <a href="https://goreportcard.com/report/github.com/imthaghost/scdl"><img src="https://goreportcard.com/badge/github.com/imthaghost/scdl"></a>
</p>

![Download](/docs/media/v2.gif)

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [How it works](#how-it-works)
- [Roadmap](#roadmap)
- [License](#license)

## Installation

### Homebrew

```bash
brew tap imthaghost/scdl
brew install scdl
```

### Go

```bash
go install github.com/imthaghost/scdl@latest
```

### Pre-built binary

Grab the latest release from the [releases page](https://github.com/imthaghost/scdl/releases).

## Usage

```bash
scdl <url>
```

Tracks:

```bash
scdl https://soundcloud.com/polo-g/polo-g-feat-juice-wrld-flex
```

Playlists / sets (every track downloaded into a subdirectory named after the playlist):

```bash
scdl https://soundcloud.com/<user>/sets/<playlist>
```

The file is written as `<track title>.mp3` with the SoundCloud artwork embedded as an ID3v2 front-cover frame. scdl exits with a non-zero status if any track fails.

### Flags

| Flag | Description |
| --- | --- |
| `-o, --output <dir>` | Output directory (default: current directory). For playlists this is the parent of the playlist subdirectory. |
| `--token <token>` | SoundCloud OAuth token. Overrides `$SCDL_TOKEN`. |
| `--proxy <url>` | Proxy URL, e.g. `http://user:pass@host:3128`. When unset, `$HTTPS_PROXY` / `$HTTP_PROXY` / `$NO_PROXY` are honored. |

### Go+ / private tracks (authenticated downloads)

Supply a SoundCloud OAuth token to unlock 256 kbps Go+ transcodings and to download private tracks your account can access:

```bash
scdl --token "$YOUR_TOKEN" <url>
# or
export SCDL_TOKEN="your-token-here"
scdl <url>
```

To find your token: open soundcloud.com in your logged-in browser, open DevTools → Network, click any request to `api-v2.soundcloud.com`, and copy the value after `OAuth ` in the `Authorization` request header.

Private share links (`?secret_token=s-XXX` or `/s-XXX`) are supported with or without a token.

> **Note:** Major-label catalog tracks (RCA/Atlantic/etc.) are served by SoundCloud only via DRM-encrypted protocols (`cbc-encrypted-hls`, `ctr-encrypted-hls`). scdl downloads the plain HLS+MP3 path only, so those tracks will fail with a `404` from api-v2. Indie creator uploads and Go+ HQ transcodings work normally.

## How it works

1. Fetches the page (with `oauth_token` cookie when a token is configured) and parses its `__sc_hydration` JSON.
2. For playlist URLs (`/sets/`), expands stub track entries via api-v2 `/tracks?ids=…` (batched), then downloads each track sequentially.
3. For each track: scrapes a fresh `client_id` from SoundCloud's JS bundles, picks the best plain-HLS `audio/mpeg` transcoding (`hq` preferred over `sq`), and builds an authorized stream URL.
4. Downloads every segment in parallel, decrypts any AES-128 segments, and assembles them in order.
5. Writes the resulting MP3 and embeds the `og:image` cover art via [`bogem/id3v2`](https://github.com/bogem/id3v2).

## Roadmap

- [x] High-quality (256 kbps) downloads via SoundCloud Go+ auth ([#13](https://github.com/imthaghost/scdl/issues/13))
- [x] Private share-link downloads (`?secret_token=` / `/s-XXX`) ([#10](https://github.com/imthaghost/scdl/issues/10))
- [x] Playlist / set URLs (`/sets/...`)
- [x] Non-zero exit code on download failure
- [x] `--output` / `-o` flag for output path
- [x] Proxy support (`--proxy` flag, `$HTTPS_PROXY` env)
- [ ] Algorithmic `/discover/sets/...` URLs (currently load via XHR — would need api-v2 playlist resolution)
- [ ] DRM-protected (major-label) tracks (Widevine/Fairplay; intentionally not supported)

## License

[MIT](https://choosealicense.com/licenses/mit/) — see [LICENSE](LICENSE).
