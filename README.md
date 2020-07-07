<p align="center">
    <img alt="gopher" src="docs/media/music.png"> <img alt="gopher" src="docs/media/pods.png"> 
</p>
<p align="center">
Scdl is the fastest SoundCloud music downloading CLI tool. Scdl utilizes a go routine pool ensuring multiple thread safe and fast downloads from SoundCloud within seconds. There are extended features such as search (no URL needed) recursively downloading all songs from a given playlist and more!
</p>
<br>
<p align="center">
   <a href="https://goreportcard.com/report/github.com/imthaghost/scdl"><img src="https://goreportcard.com/badge/github.com/imthaghost/scdl"></a>
   <a href="https://api.travis-ci.org/imthaghost/scdl.svg?branch=master">
    <img src="https://api.travis-ci.org/imthaghost/scdl.svg?branch=master"alt="build status">
   <a href="https://github.com/imthaghost/gitmoji-changelog">
    <img src="https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg"alt="awesome">
  </a>
</p>
<br>

![Download](/docs/media/v2.gif)

## Table of Contents

-   [Installation](#installation)
-   [Usage](#usage)
-   [Examples](#Examples)
-   [Todo](#Todo)
-   [License](#license)
-   [Contributors](#contributors)

## 🚀 Installation

### Brew

```bash
# tap
brew tap imthaghost/scdl
# install tool
brew install scdl
```

### Manual

```bash
# go get :)
go get https://github.com/imthaghost/scdl
# change to project directory using your GOPATH
cd $GOPATH/src/github.com/imthaghost/scdl
# build and install application
go install
```

### Binary

[Download Here](https://www.mediafire.com/file/ynkvkaoo4rvvv4v/scdl/file)

## Usage

  ``-h, --help`` -  Help screen and usage

  ``-s, --search`` - Option for searching for songs


## Examples

### Base Command
```bash 
# command + SounCloud URL
scdl https://soundcloud.com/polo-g/polo-g-feat-juice-wrld-flex
```

### Search
```bash 
# search flag
scdl lucid dreams --search
# or
scdl lucid dreams -s
```

## Todo

### Short term

-   [x] Cobra command line interface
-   [x] Download audio file from Soundcloud URL
-   [x] Goroutine pool for downloading m3u8 file
-   [x] Installation via Brew
-   [x] Mp3 file contains image cover
-   [x] Download a song through search functionality
-   [ ] 80-100% test coverage
-   [ ] Update tool for better performance
-   [ ] Proxy flag
-   [ ] Format flag
### Long term
-   [ ] Search results
-   [ ] Download all songs from a given playlist
-   [ ] Download all songs from a given album

## 📝 License

By contributing, you agree that your contributions will be licensed under its MIT License.

In short, when you submit code changes, your submissions are understood to be under the same [MIT License](http://choosealicense.com/licenses/mit/) that covers the project. Feel free to contact the maintainers if that's a concern.

## Contributors

Contributions are welcome! Please see [Contributing Guide](https://github.com/imthaghost/goclone/blob/master/docs/CONTRIBUTING.md) for more details.

<table>
  <tr>
    <td align="center"><a href="https://github.com/imthaghost"><img src="https://avatars3.githubusercontent.com/u/46610773?s=460&v=4" width="75px;" alt="Gary Frederick"/><br /><sub><b>Tha Ghost</b></sub></a><br /><a href="https://github.com/imthaghost/scdl/commits?author=imthaghost" title="Code">💻</a></td>
     <td align="center"><a href="https://github.com/Tylerholland12"><img src=https://avatars3.githubusercontent.com/u/29693747?s=460&u=fe7499a0450042c5cd0153c2f8945db89dd39e71&v=4" width="75px;" alt="Tyler H"/><br /><sub><b>Tyler H</b></sub></a><br /><a href="https://github.com/imthaghost/scdl/commits?author=Tylerholland12" title="Code">💻</a></td>
  </tr>