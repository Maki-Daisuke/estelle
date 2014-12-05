Estelle - A Thumbnail Daemon
============================

Description
-----------

Estelle is a daemon that generates and caches thumbnails of images.

In some systems, there are many programs that need thumbnails of images. However,
it's quite inefficient to generarte and keep multiple copies of a thumbnail of an
image.

Estelle solves this problem. It provides system-wide thumbnail pool for user
programs, which allows user to easily generate, keep and find (if already exists)
thumbnails.


How to Install
--------------

Estelle is implemented in Go. You need to install Go tools at first.
Then, clone the repository and build it:

    git clone https://github.com/Maki-Daisuke/Estelle.git
    cd Estelle
    go build

That's it!


Run Estelled
------------

    ./estelled

This command starts Estelle daemon. It starts listening TCP port specified by
`-port` option and blocks your shell line until you hit Ctrl+C.

### Options

Command line options controls cache strategy, which id how Estelled purges old
thumbnails. There are two options available:

- `-port=<PORT>` | `-p <PORT>`
  - Port number that Estelled listens
  - Default: 1186
- `-cache-dir=<PATH TO DIR>` | `-d <PATH TO DIR>`
  - Directory to cache thumbnails
  - Default: ./estelled-cache
- `-expires=<MIN>` | `-E <MIN>`
  - Purge thumbnails that have not been accessed for `<MIN>` minutes.
- `-limit=<SIZE>` | `-L <SIZE>`
  - Keep the size of cache-directory smaller than `<SIZE>` MB, by purging least
    recent used thumbnails.

Get Thumbnail
-------------

Estelle is a HTTP server, so that you can call it by just sending HTTP request. For example:

    curl http://localhost:<port-number>/<absolute-path-to-image-file>?size=400x400

This will return a single line of string as the response body, that is the file
path of thumbnail you want.

### Query Parameters

- `size`
  - Size of the generated thumbnail
  - Default: `85x85`
- `overflow`
  - When the aspect ratio of `size` differs from the one of the original file, this option specifies how to handle
  - One of these:
    - `fill`: resizes the image with regarding `size` as maximum width and height, and fills background with white.
    - `fit`: resizes the image with regarding `size` as minimum width and height, and cut out extra edges as it fits the specified `size`.
    - `shrink`: resizes the image with regarding `size` as maximum width and height. The resulted thumbnail is smaller than `size`.
  - Default: `fill`
- `format`
  - Image format of the output thumbnail
  - One of: `jpg`, `png`, `webp`
  - Default: `jpg`

Caching
-------

Estelle caches generated thumbnails in a directory specified by `path-to-cache-dir`
command-line parameter. Estelle identifies a thumbnail corresponding to a passed image
with hash of the image (which would be SHA1, but implementation dependent). That is,
every time Estelle is asked to serve a thumbnail of an image, it calculates hash
value of the image, then find an appropriate thumbnail. If there is no thumbnail
cached for the request, it generates a thumbnail and returns file path to it.


Term of Use
-----------

This software is distributed under the revised BSD License.

Copyright (c) 2014, Daisuke (yet another) Maki All rights reserved.
