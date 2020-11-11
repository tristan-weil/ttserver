{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{/********** Main **********/ -}}
:: ABOUT THIS SERVER ::

{{ ginfo (tablewriter (dict "data" (list (list "INTRODUCTION")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
`Gopher' is a direct competitor of HTTP. It failed to evolve and is
stuck with a basic set of features. But it does provide links.
Thus it is a really light alternative to what became the WWW
(websites + browsers).

And there are still a hundred (or a thousand) of users.

{{ ginfo (tablewriter (dict "data" (list (list "THE GOPHER SPACE")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
This `Gopher space' is part of my personal project of building an
infrastructure that:
- helps me (as a sys/admin) to train myself and test some tools
- is interesting and fun to manage

So it used a server I wrote in Go and is hosted on my infratructure.

The Web site
{{ gurl "https://t18s.fr" "https://t18s.fr" -}}
is the main entry point for all my personal projects and you'll find
a detailed description of the infrastructure.

This `Gopher space' won't be a carbon-copy of the Web site.
I will use it to provide some specific contents and use some features
that allow to generate some dynamic contents.

{{ ginfo (tablewriter (dict "data" (list (list "THE OLD CODE")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
Some time ago, I created 'pygopher' (sorry, the source code is not
available anymore).
It served this domain well for almost a year.

The code used the impressive
{{ gurl "https://github.com/michael-lazar/flask-gopher" "Flask-Gopher library" -}}
maintained by Michael Lazar to handle the Gopher protocol.It is itself
based on the well-known Python Web framework:
{{ gurl "http://flask.pocoo.org/" "Flask" }}

{{ ginfo (tablewriter (dict "data" (list (list "THE NEW CODE")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
But, as for now, I am currently working on a complete rewrite in Go:
{{ gurl "https://github.com/tristan-weil/ttserver" "github.com/tristan-weil/ttserver" }}
`ttserver' is a more generic TCP server with more features and
it allows to handle different protocols (like Gopher or finger).

The goal of this project is not to create a fast or efficient
TCP server. It is a playground to develop my skills in Go and have
fun with some protocols.

Features:
- Route:
  - a route:
    - points to a template (with .tpl extension)
    - or points to a file (served as binary)
  - it can be configured in the configuration file or not:
    - if not: the template (.tpl) or the file must exists
    - if configured: a custom template or file can be pointed
  - a route can be a regex:
    - the name of the template or file can use the regex's
      captured groups

- Template:
  - the main method to display data is templating (but bin files can
    be served)
  - a lots of functions can be used inside templates
  - custom functions can be added as part of a custom handler
  - a header or a footer template can be loaded

- Handler:
  - a connection handler implement the read, parse, templating and
    write operations
  - it is executed for each new connection
  - it serves a route

- Misc:
  - an internal cron can execute 'routes'
  - a caching system (global but can be configured by 'route')

- Implementation:
  - graceful shutdowns / reloads
  - smart reloads (configurations and/or listeners modifications are
    loaded with no interruptions if possible)
  - 2 handlers included:
    - a basic Finger and Gopher implementations
  - default implementations of the read, parse, template and write
    operations allow to:
    - fetch data from external sources (JSON and Feeds) with a
      routines pool
    - use an internal cache
    - handle errors (missing route and server errors fallback to
      another route)

- Protocols:
  - ACME protocol (Let's Encrypt)
  - PROXY protocol (v1/v2)
  - metrics are exposed with the Prometheus exposition format on
    a HTTP /metrics compatible endpoint

The main modules used by this project:
{{ gurl "https://github.com/Masterminds/sprig/v3" "github.com/Masterminds/sprig/v3" -}}
{{ gurl "https://github.com/caddyserver/certmagic" "github.com/caddyserver/certmagic" -}}
{{ gurl "https://github.com/mmcdole/gofeed" "github.com/mmcdole/gofeed" -}}
{{ gurl "https://github.com/jedib0t/go-pretty" "github.com/jedib0t/go-pretty" -}}
{{ gurl "https://github.com/patrickmn/go-cache" "github.com/patrickmn/go-cache" -}}
{{ gurl "https://github.com/pires/go-proxyproto" "github.com/pires/go-proxyproto" -}}
{{ gurl "https://github.com/prometheus/client_golang" "github.com/prometheus/client_golang" -}}
{{ gurl "https://github.com/robfig/cron/v3" "github.com/robfig/cron/v3" }}

It should also be noted that everything also started from this code:
{{ gurl "https://github.com/mitchellh/go-finger" "github.com/mitchellh/go-finger" }}

The `Gopher' handler I implemented use a lot of the good ideas from
this Python module:
{{ gurl "https://github.com/michael-lazar/flask-gopher" "github.com/michael-lazar/flask-gopher" }}

{{ ginfo (tablewriter (dict "data" (list (list "THE GOPHER HANDLER")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
The provided `Gopher' handler should allow to create basic `Gopher'
pages. Not all types can be used for the moment.

Templates can use the following functions:

- all functions from:
{{ gurl "https://github.com/Masterminds/sprig/v3" "github.com/Masterminds/sprig/v3" }}

- common handlers' functions:
  - normalize: ASCIIfy
  - floating: allow to have one text left-aligned and another righ-aligned
  - rpad: right-align with a custom character
  - lpad: left-align with a custom character
  - center: center
  - rjust: right-align
  - ljust: left-align
  - underline: underline with a custom character
  - bytesize: convert an int to a human-readable bytes size
  - duration: convert seconds to human-readable duration
  - build_version: the version of the current instance of `ttserver'
  - build_date: the date of the build of the current instance of *
  `ttserver'
  - bitcoincoreServices: convert to bitcoin services names
  - tablewriter: construct complex tables, see:
{{ gurl "https://github.com/jedib0t/go-pretty" "github.com/jedib0t/go-pretty" }}

- this `Gopher' handler special functions:
  - gurl_for: find an internal link
  - gmenu: internal link (DIR tag)
  - ginfo: info
  - gerror: error
  - gtitle: title
  - gurl: external link (HTML tag)

{{ ginfo (tablewriter (dict "data" (list (list "HOW TO CONNECT")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
No need to write a client, 'lynx' runs great.

Exemple:
$ lynx gopher://gopher.t18s.fr

But if you want to use a SSL connection, the use of a third-party
tunneling program is needed.

{{ ginfo (tablewriter (dict "data" (list (list "DEPLOYMENT")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")")) -}}
Binaries are available from here:
{{ gurl "https://github.com/tristan-weil/ttserver/releases" "github.com/tristan-weil/ttserver/releases" }}

You can also deploy the Docker image:
- docker pull ghcr.io/tristan-weil/ttserver:latest
- docker run -p 7070:7070 ghcr.io/tristan-weil/ttserver:latest
or
docker run -p 7070:7070 -v /your/config:/ttserver/ttserver.config:ro \
  -v /your/content/dir:/ttserver/public:ro \
  ghcr.io/tristan-weil/ttserver:latest

{{ gurl "https://github.com/users/tristan-weil/packages/container/package/ttserver" "More Docker images here" }}
{{ template "footer.tpl" . }}
