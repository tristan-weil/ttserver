{{ template "header.tpl" . }}
{{/********** Global Vars **********/ -}}
{{ $width := 70 -}}
{{/********** Main **********/ -}}
:: ABOUT THIS SERVER ::

{{ tablewriter (dict "data" (list (list "INTRODUCTION")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") }}
A long time ago, I found a website that listed on its venerable
finger service the GNU/Linux distributions and BSD systems that you
can download.
That misuse was so fun and ingenious that I wanted to imitate it.

{{ tablewriter (dict "data" (list (list "THE FINGER SPACE")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") }}
This `finger space' is part of my personal project of building an
infrastructure that:
- helps me (as a sys/admin) to train myself and test some tools
- is interesting and fun to manage

So it used a server I wrote in Go and is hosted on my infratructure.

The Web site "https://t18s.fr" is the main entry point for all my
personal projects and you'll find a detailed description of the
infrastructure.

This `finger space' won't be a carbon-copy of the Web site.
I will use it to provide some specific contents and use some features
that allow to generate some dynamic contents.

{{ tablewriter (dict "data" (list (list "THE OLD CODE")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") }}
Some time ago, I created 'pfinger' (sorry, the source code is not
available anymore).
It was in Perl and served this domain well for almost a year.

{{ tablewriter (dict "data" (list (list "THE NEW CODE")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") }}
But, as for now, I am currently working on a complete rewrite in Go:
"https://github.com/tristan-weil/ttserver"

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
"https://github.com/Masterminds/sprig/v3"
"https://github.com/caddyserver/certmagic"
"https://github.com/mmcdole/gofeed"
"https://github.com/jedib0t/go-pretty"
"https://github.com/patrickmn/go-cache"
"https://github.com/pires/go-proxyproto"
"https://github.com/prometheus/client_golang"
"https://github.com/robfig/cron/v3"

It should also be noted that everything also started from this code:
"https://github.com/mitchellh/go-finger"

{{ tablewriter (dict "data" (list (list "THE FINGER HANDLER")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") }}
The provided `finger' handler should allow to create basic pages.

Templates can use the following functions:

- all functions from: "https://github.com/Masterminds/sprig/v3"

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
"https://github.com/jedib0t/go-pretty"

{{ tablewriter (dict "data" (list (list "HOW TO CONNECT")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") }}
No need to write a client, 'finger' runs great.

Exemple:
$ finger @finger.t18s.fr

But if you want to use a SSL connection, the use of a third-party
tunneling program is needed.

{{ tablewriter (dict "data" (list (list "DEPLOYMENT")) "width" $width "text-alignment" "center" "box-separator" "~" "box-left" ")" "box-right" ")") }}
Binaries are available from here:
"https://github.com/tristan-weil/ttserver/releases"

You can also deploy the Docker image:
- docker pull ghcr.io/tristan-weil/ttserver:latest
- docker run -p 7070:7070 ghcr.io/tristan-weil/ttserver:latest
or
docker run -p 7070:7070 -v /your/config:/ttserver/ttserver.config:ro \
  -v /your/content/dir:/ttserver/public:ro \
  ghcr.io/tristan-weil/ttserver:latest

More Docker images here:
"github.com/users/tristan-weil/packages/container/package/ttserver"
{{ template "footer.tpl" . }}
