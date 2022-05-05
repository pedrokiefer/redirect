# HTTP service discovery

![Mx](docs/logo.svg)

[![GitHub release](https://img.shields.io/github/release/reddec/redirect.svg)](https://github.com/reddec/redirect/releases)
[![license](https://img.shields.io/github/license/reddec/redirect.svg)](https://github.com/reddec/redirect)
[![](https://godoc.org/github.com/reddec/redirect?status.svg)](http://godoc.org/github.com/reddec/redirect)
[![Snap Status](https://build.snapcraft.io/badge/reddec/redirect.svg)](https://build.snapcraft.io/user/reddec/redirect)

Simple, very fast, in a single binary, with web UI HTTP redirect service

# Install

[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/redirect-to)

* [snapcraft: redirect-to](https://snapcraft.io/redirect-to)

* Precompilled binaries: [release page](https://github.com/reddec/redirect/releases)

* From source (required Go toolchain):

```
go get -v -u github.com/reddec/redirect/...
```

**snapcraft use `redirect-to` package**

# WEB panel

![web interface](http://reddec.github.io/images/redirect_ui.png)

# Usage

## Docker

Translate this ports as you wish:

* `10100` - main redirect server
* `10101` - UI web panel

Use exposed volume `/etc/redirect` to persist data
## CLI

### -bind

Redirect address (default "0.0.0.0:10100"). You can do any HTTP operation
to address http://your-server:10100/your/cool/service/name and it will be redirected to specified address

### -config

File to save configuration (default "./redir.json").
It will be loaded at startup (if exists) and saved after each modification operation over API

### -ui

Directory of static UI files. If not defined - embedded used

### -ui-addr string

Address for UI (default "127.0.0.1:10101")

* `/` - Will be served as static directory from specified directory
* `/ui/` - UI interface
* `/api/`  - API handlers

# Actions on redirect server

* `GET/POST/PUT/DELETE` - returns redirection with 302 Found status
* `HEAD` - returns only real service location in `Location` header with 200 OK status

# API

### GET

Get list of services or detailed information of one service if service name provided.
Adds into response headers (`X-Redir-Port`) port of main server

* Endpoint (all):  `http://ui-addr/api/`
* Endpoint (one):  `http://ui-addr/api/your/cool/service/name`

### POST

Add or update one service. If service already exists, hits will saved.
Expects this form fields:

* `service` - service name
* `template` - content of template

Each template must be valid expression of [Go template engine](https://golang.org/pkg/text/template/)
with [http request](https://golang.org/pkg/net/http/#Request) as environment.

#### Simple example

* `service` = test
* `template` = http://google.com

All requests to http://test will be redirected to http://google.com

#### Complex example

* `service` = mdn
* `template`

    https://developer.mozilla.org/ru/docs/Web/JavaScript/Reference/Global_Objects/{{.URL.Query.Get "q"}}

All requests to http://mdn will be redirected to
https://developer.mozilla.org/ru/docs/Web/JavaScript/Reference/Global_Objects/QUERY

E.x.:

`http://mdn/?q=encodeuricomponent` maps to
`https://developer.mozilla.org/ru/docs/Web/JavaScript/Reference/Global_Objects/encodeuricomponent`

* Endpoint:  `http://ui-addr/api/`

### DELETE

Remove service if it exists

* Endpoint:  `http://ui-addr/api/your/cool/service/name`
