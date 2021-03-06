Estelle - A Thumbnail Daemon
============================

_**NOTE: This is experimental and still heavily under development. Most of the
features documented here is not implemeted yet.**_

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

Since Estelle depends on ImageMagick, install ImageMagick at first. You can
install ImageMagick with package manager of your OS. For example, you can use
[Homebrew](http://brew.sh/) on OS X:

    brew install imagemagick

Estelle is implemented in Go. You need to install [Go tools](http://golang.org/doc/install).
Then, just get Estelle:

    go get github.com/Maki-Daisuke/estelle/cmd/estelled

Or, you can clone the repository and build it:

    git clone https://github.com/Maki-Daisuke/estelle.git
    cd estelle/cmd/estelled
    go build

That's it! Now you have a binary called `estelled`.


Run Estelled
------------

    ./estelled

This command starts Estelle daemon. It starts listening TCP port specified by
`-port` option and blocks your shell line until you hit Ctrl+C.

### Options

Command line options controls cache strategy, which id how Estelled purges old
thumbnails. There are two options available:

- `--port=<PORT>` | `-p <PORT>`
  - Port number that Estelled listens
  - Default: 1186
- `--cache-dir=<PATH TO DIR>` | `-d <PATH TO DIR>`
  - Directory to cache thumbnails
  - Default: ./estelled-cache
- `--expires=<MIN>` | `-E <MIN>`
  - Purge thumbnails that have not been accessed for `<MIN>` minutes. `<MIN>`
    smaller than or equals to zero means no expiration.
  - Default: 0
- `--limit=<SIZE>` | `-L <SIZE>`
  - Keep the size of cache-directory smaller than `<SIZE>` MB, by purging least
    recent used thumbnails. `<SIZE>` smaller than or equals to zero means no limit.
  - Default: 0

How to Use
----------

Estelle is a HTTP server, so that you can call it by just sending HTTP request.
For example:

    curl http://localhost:1186/path?source=<absolute-path-to-image-file>&size=400x400


This will return a single line of string as the response body, that is the file
path of thumbnail you want.

You can directly retrieve content of thumbnail as HTTP response by requesting
`/content`:

    curl http://localhost:1186/content?source=<absolute-path-to-image-file>&size=400x300

This will return response body in image/jpeg format containing thumbnail image.

Commands
--------

### `/path`, `/content`

- Method: GET

`path` returns the absolute file path of thumbnail of the specified image.
If the thumbnail does not exist yet, it will create it on the fly. That means,
it will block until the thumbnail is created. If you do not want to block,
please use `/status` and `/queue` instead.

An original image can be specified by either of `source` patameter, `id` parameter
or content body of HTTP request.
Priority is: `id` > `source` > request body.

For example, if you want thumbnail of `/foo/bar/baz.jpg`, you can request like this:

    curl http://localhost:1186/path?source=/foo/bar/baz.jpg&size=400x300&mode=fill

Here, `size` specifies thumbnail size and `mode` specifies how to treat different aspect ratio.

`/content` works the same way as `/path`, except it returns content of thumbnail
as response body, instead of file path. That means, clients that running on
different hosts can call this API.

### `/queue`

- Method: POST

Request to make thumbnail. Thumbnailing task is queued and the response will be
returned immediately. The thumbnailing task is executed in background in order.

If the thumbnailing task is successfully queued, `/queue` will return `202 Accepted`.
If the thumbnail already exists, it will return `200 OK`.

### `/status`

Returns status of thumbnail. This will return response immediately, that is, it
does not block until the thumbnail task is done.

The status is shown as status code as follows:

- 200: the requested thumbnail already exists
- 202: the requested thumbnail has been queued to thumbnail
- 404: the requested thumbnail does not exist

#### Query Parameters

- `source`
  - Path to image file
  - Required parameter
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

Estelle caches generated thumbnails in a directory specified by `--cache-dir`
command-line option. Estelle identifies a thumbnail corresponding to a passed image
with hash of the image (which would be SHA1, but implementation dependent). That is,
every time Estelle is asked to serve a thumbnail of an image, it calculates hash
value of the image, then find an appropriate thumbnail. If there is no thumbnail
cached for the request, it generates a thumbnail and returns file path to it.


Term of Use
-----------

This software is distributed under the revised BSD License.

Copyright (c) 2014-2015, Daisuke (yet another) Maki All rights reserved.
