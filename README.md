# ttserver

`ttserver` is a simple TCP server aiming to handle simple text protocols.

[![Build Status](https://github.com/tristan-weil/ttserver/workflows/Build/badge.svg)](https://github.com/tristan-weil/ttserver/actions?query=workflow%3ABuild)
[![Linters Status](https://github.com/tristan-weil/ttserver/workflows/Lint%20Code%20Base/badge.svg)](https://github.com/tristan-weil/ttserver/actions?query=workflow%3A%22Lint+Code+Base%22)

## SYNOPSIS

```shell
ttserver -config <path> [-help]
```

## OPTIONS

```shell
-config <path>
```
The path of the configuration file.

```shell
-help
```
Display the help message.

## DESCRIPTION

`ttserver` is a simple TCP server allowing to:
- serve content over custom connection handlers:
  - `finger` and `gopher` protocols' handlers included
  - a basic page renderer supporting:
    - Go templates
    - caching
    - fetching remotes contents
- define generic routes with regex
- fetch remote contents (json, html, feeds) and display them in pages
- run cron tasks to create/update pages
- server content over TLS with the support of the ACME protocol ([Let's Encrypt](https://letsencrypt.org/))
- gather stats about the server on a
[Prometheus compatible endpoint](https://prometheus.io/docs/instrumenting/exposition_formats/), /metrics
- handle [PROXY protocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt)

## CONFIGURATION

The configuration file is a JSON formatted file.

### Logging (log)

The `log` object is used to configure the logging system:

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `level` | info | error, warn, info, debug, trace, panic | The log level | |
| `fomat` | text | text, json | The log format | |
| `to` | stderr | discard, stderr, stdout, syslog | The destination of the log message | |
| `syslog` | | | The address of the syslog daemon | |

Example:
```json
  "log": {
    "level": "trace",
    "format": "text",
    "to": "stderr"
  },
```

### Prometheus (prometheus)

The `prometheus` object is used to configure the Prometheus endpoint:

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `address` | 127.0.0.1:9777 | any valid IP:port address | The address used for the endpoint | |
| `endpoint` | /metrics | any valid path | The path for the endpoint | |
| `summaryMaxAge` | default [prometheus](https://github.com/prometheus/client_golang) library value (see prometheus.DefMaxAge) | any valid float64 | The duration, in seconds, for which observations stay relevant. | |
| `chanSize` | 1024 | any valid int | The size of the channel processing fired events | |
| `auth` | | basic | The authentication method to access the endpoint | |

Example:
```json
  "prometheus": {
    "address": "127.0.0.1:9777",
    "endpoint": "/metrics",
    "summaryMaxAge": 86400,
    "chanSize": 1024
  },
```

#### Authentication (prometheus.auth)

##### Basic Auth (prometheus.auth.basic)

The `prometheus.auth.basic` object is used to configure the Prometheus authentication access to the endpoint
with `Basic Auth`:

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `username` | | any valid string | The username used to authenticate | |
| `password` | | any valid string | The password used to authenticate | |

Example:
```json
  "prometheus": {
    ...
    "auth": {
      "basic": {
        "username": "username",
        "password": "password"
      }
    },
  },
```

### Space (space)

The `space` object is used to configure a space (in other words: what will be seved and how):

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `cache` | see below | see below | The cache manager | |
| `listener` | see below | see below | The TCP listener | |
| `handler` | | finger, gopher | | X |
| `basedir` | current workdir | any valid path | The path where the contents are stored | |
| `routes` | see below | see below | The configuration of the routes | |

Example:
```json
  "space": {
    "cache": {
        ...
    },
    "listener": {
        ...
    },
    "handler": {
        ...
    },
    "basedir": "examples/gopher",
    "routes": {
        ...
    }
```

#### Cache (space.cache)

The `space.cache` object is used to configure the cache.

The cache is only relevant if you use a handler with page renderer functions that use it.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `expiration` | 300 | -1, 0, any valid int | The default TTL, in seconds, of objects in the cache (-1 disables the cache / 0 means unlimited) | |
| `memory` | see below | see below | A cache manager using an in-memory store | |

##### Memory (space.cache.memory)

The `space.cache.memory` object is used to configure the in-memory cache store.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `cleanup` | 350 | any valid int | The delay, in seconds, to run the GC of the stored objects | |

#### Listener (space.listener)

The `space.listener` object is used to configure the listener.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `address` | 127.0.0.1:7575 | any valid IPv4:port address | The IPv4 and port to bind to | |
| `domains` | ["localhost"] | any valid list of domains | A list of domains (currently only used by the ACME feature) | |
| `tls` | | acme | The TLS certificates manager | |

Example:
```json
  "space": {
    ...
    "listener": {
      "address": "127.0.0.1:7070",
      "domains": ["gopher-test.t18s.fr"]
    }
  }
```

##### TLS (space.listener.tls)

###### ACME (space.listener.tls.acme)

The `space.listener.tls.acme` object is used to configure the ACME client.

Currently, only the DNS challenge is supported with a small subset of providers.

To configure a provider, the credentials need to be provided using environment variables:
- Gandi: GANDI_API_TOKEN
- Cloudflare: CLOUDFLARE_API_TOKEN
- Hetzner: HETZNER_API_TOKEN
- DigitalOcean: DO_API_TOKEN

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `dnsprovider` | | gandi, cloudflare, hetzner, digitalocean | The DNS provider used for the DNS challenge | X |
| `email` | | any valid email | The email used to register against the ACME server | X |
| `ca` | <https://acme-v02.api.letsencrypt.org/directory> | any valid ACME url | The url used to request a certificate | |
| `testca` | <https://acme-staging-v02.api.letsencrypt.org/directory> | any valid ACME url | The url used to test the access | |
| `storage` | | file | The storage for the ACME working files | |

Example:
```json
  "space": {
    ...
    "listener": {
      ...
      "tls": {
        "acme": {
          "dnsprovider": "gandi",
          "email": "xxxx@xxxx",
          "ca": "https://acme-staging-v02.api.letsencrypt.org/directory",
          "testca": "https://acme-staging-v02.api.letsencrypt.org/directory",
          "storage": {
            "file": {
              "path": ".certmagic"
            }
          }
        }
      }
    }
  }
```

####### ACME storage (space.listener.tls.acme.storage)

The `space.listener.tls.acme.storage` object is used to configure the ACME storage.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `file` | see below | see below | The file storage | |

####### ACME file storage (space.listener.tls.acme.storage.file)

The `space.listener.tls.acme.storage.file` object is used to configure the ACME file storage.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `path` | .certmagic | any valid path | The folder containing the ACME working files | |

#### Handler (space.handler)

The `space.handler` object is used to configure the handler.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `name` | | any valid name | The name of a registred handler | X |
| `parameters` | | a map of custom parameters | Some handlers could require extra parameters | |

Example:
```json
  "space": {
    ...
    "handler": {
      "name": "finger",
    }
  }
```

##### Handler: Finger (space.handler)

The `space.handler` object used by the `Finger` handler allows custom parameters.

The Finger protocol don't have know how to identify the requested domain (like the `Host` header in HTTP).
But the address could be used in a template

So here is the decision tree for the domain: `response_domain` > `SNI domain (if enabled)` > `space.listener.domains[0]` > `space.listener.address (address part)`

And for the port: `response_rt` > `space.listener.address (port part)`

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `response_domain` | see above | any valid domain or IP | The domain used in the response | |
| `response_port` | | any valid domain or IP | The port used in the response | |

Example:
```json
  "space": {
    ...
    "handler": {
      "name": "finger",
      "parameters": {
        "response_domain": "finger-test.t18s.fr",
        "response_port": "7979"
      }
    }
  }
```

##### Handler: Gopher (space.handler)

The `space.handler` object used by the `Gopher` handler allows custom parameters.

The Gopher protocol don't have know how to identify the requested domain (like the `Host` header in HTTP).
But :
- the response needs to contain the address and port of the server
- the address could be used in a template

So here is the decision tree for the domain: `response_domain` > `SNI domain (if enabled)` > `space.listener.domains[0]` > `space.listener.address (address part)`

And for the port: `response_rt` > `space.listener.address (port part)`

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `response_domain` | see above | any valid domain or IP | The domain used in the response | |
| `response_port` | | any valid domain or IP | The port used in the response | |

Example:
```json
  "space": {
    ...
    "handler": {
      "name": "gopher",
      "parameters": {
        "response_domain": "gopher-test.t18s.fr",
        "response_port": "7070"
      }
    }
  }
```

#### Routes (space.routes)

The `space.routes` object is used to configure a map<name, route object> of routes.

The **name** can be:
- a name of a page pointing to an existing template file (without the .tpl extension)
- a name of a page where the template file is defined in the **route** object
- a regex that can matches multiples name:
  - the template file is defined in the **route** object
  - the regex capturing groups can be used to point to a template file

##### Route (space.routes.\<name>)

The `space.route.<name>` object is used to configure a route.

A route can render the content of a template or a file.
By default, the name of the route is used to find a template and then a file.

A route can be automatically executed by a cron process that will fake a connection.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `template` | the name of the route | any valid file in **basedir** (accepts regex's capturing group) | A template file (without the .tpl extension) to render for this route | |
| `file` | the name of the route |  any valid file in **basedir** (accepts regex's capturing group) | A file (raw contents are returned) | |
| `fetch` | | see below | a map of content to fetch when the page is rendered | |
| `cron` | | any valid cron format (+ the seconds at first position) | A cron render the page | |
| `cache` | | any valid file in **basedir**  | Custom parameters for the caching of this page | |

Example:
```json
  "space": {
    ...
    "routes": {
      "~^blog/(\\d{4})/(\\d{2})$": {
        "template": "blog/$1-$2.tpl"
      },
    }
  }
```

###### Fetching (space.routes.\<name>.fetch)

The `space.route.<name>.fetch` object is used to configure a map<name, fetch object> of fetches.

###### Fetching (space.routes.\<name>.fetch.\<name>)

The `space.route.<name>.fetch.<name>` object is used to configure a fetch.

A fetch allows to get content from external JSON or Feed (atom, rss) pages and inject them into a page.

| Option | Default value | Allowed values | Description | Mandatory |
| ------ | ------------- | -------------- | ----------- | --------- |
| `uri` | | any valid URI | A target URI to fetch | |
| `type` | | html, json, feed, prometheus | The type of the target URI | |

Example:
```json
  "space": {
    ...
    "routes": {
      "aggregator/unix": {
        "template": "aggregator.tpl",
        "fetch": {
          "1_slashdot": {
            "type": "feed",
            "uri": "http://rss.slashdot.org/Slashdot/slashdotMain"
          },
          "2_undeadly": {
            "type": "feed",
            "uri": "https://undeadly.org/cgi?action=rss&items=15"
          },
          "3_undeadly_errata": {
            "type": "feed",
            "uri": "https://undeadly.org/errata/errata.rss"
          }
        },
        "cron": "0 30 */5 * * *"
      },
    }
  }
```

## EXAMPLES

See [examples](examples) for some configurations.

## TODO

Common:
- make the linters happy
- tests

Listener:
- allow to listen on multiple addresses
- add IPv6 support

TLS:
- allow to customize TLS ciphers
- add more certs' handlers

ACME:
- handle more dns providers
- add more challenges
- better handling of the dns providers (don't import them all?)

## LICENSE

See [LICENSE.md](LICENSE.md)
