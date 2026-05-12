<p align="center">
    <img alt="gopher" src="docs/media/pods.png">
</p>
<p align="center">
A fast SoundCloud track downloader written in Go. Give it a track URL, get back an <code>.mp3</code> with embedded cover art.
</p>
<p align="center">
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
scdl <track-url>
```

Example:

```bash
scdl https://soundcloud.com/polo-g/polo-g-feat-juice-wrld-flex
```

The file is written to the current directory as `<track title>.mp3` with the SoundCloud artwork embedded as an ID3v2 front-cover frame.

### Go+ / private tracks (authenticated downloads)

Supply a SoundCloud OAuth token to unlock 256 kbps Go+ transcodings and to download private tracks your account can access. Either pass it on the flag:

```bash
scdl --token "$YOUR_TOKEN" <track-url>
```

…or set it once in your shell:

```bash
export SCDL_TOKEN="your-token-here"
scdl <track-url>
```

To find your token: open soundcloud.com in your logged-in browser, open DevTools → Network, click any request to `api-v2.soundcloud.com`, and copy the value after `OAuth ` in the `Authorization` request header.

Private share links (`?secret_token=s-XXX` or `/s-XXX`) are supported with or without a token.

## How it works

1. Fetches the track page and parses its `__sc_hydration` JSON.
2. Scrapes a fresh `client_id` from SoundCloud's JS bundles.
3. Builds an authenticated HLS playlist URL using the track's `track_authorization` token.
4. Downloads every segment in parallel, decrypts any AES-128 segments, and assembles them in order.
5. Writes the resulting MP3 and embeds the `og:image` cover art via [`bogem/id3v2`](https://github.com/bogem/id3v2).

## Roadmap

- [x] High-quality (256 kbps) downloads via SoundCloud Go+ auth ([#13](https://github.com/imthaghost/scdl/issues/13))
- [x] Private share-link downloads (`?secret_token=` / `/s-XXX`) ([#10](https://github.com/imthaghost/scdl/issues/10))
- [ ] Playlist / set URLs (`/sets/...`)
- [ ] Non-zero exit code on download failure
- [ ] `--output` / `-o` flag for output path
- [ ] Proxy support

## License

[MIT](https://choosealicense.com/licenses/mit/) — see [LICENSE](LICENSE).
