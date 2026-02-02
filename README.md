# Estelle - A Thumbnail Daemon

_**NOTE: This is experimental and still heavily under development. Most of the
features documented here is not implemeted yet.**_

## Description

Estelle is a daemon that generates and caches thumbnails of images
designed _for Linux embedded systems_.

In some systems, there are many programs that need thumbnails of images. However,
it's quite inefficient to generarte and keep multiple copies of a thumbnail of an
image.

Estelle solves this problem. It provides system-wide thumbnail pool for user
programs, which allows user to easily generate, keep and find (if already exists)
thumbnails.

## Requirements

Estelle relies on specific Linux capabilities to ensure high performance and correct cache validation.

* **Operating System**: Linux
  * Minimum Kernel version depends on the Go runtime policy (e.g. Linux 3.2+ for Go 1.25).
  * Windows and macOS are NOT supported.
* **Go Runtime**: 1.25 or later
* **File System**:
  * Recommended: ext4, XFS, Btrfs, F2FS (must support nanosecond resolution timestamps)
  * Limited Support: ext3, FAT32/exFAT
    * **Warning**: On file systems without nanosecond timestamp support, modifications made within the same second may be ignored by the cache system.
* **Library**: `libvips` development headers for building, and shared libraries for running.

## How to Install

Since Estelle depends on libvips, install libvips at first. You need
install libvips with package manager like `apt` or `yum`:

    apt install libvips

Or, you need to install it from source.

Estelle is implemented in Go. You need to install [Go tools](http://golang.org/doc/install).
Then, just get Estelle:

    go install github.com/Maki-Daisuke/estelle/cmd/estelled@latest

Or, you can clone the repository and build it:

    git clone https://github.com/Maki-Daisuke/estelle.git
    cd estelle/cmd/estelled
    go build -o path/to/estelled

That's it! Now you have a binary called `estelled`.

## Run Estelled

    ./estelled /path/to/allowed/images /another/path
    ./estelled . 

This command starts Estelle daemon. It starts listening TCP port specified by
`--addr` option.

You MUST specify one or more directories as positional arguments. Estelle will only allow access to images within these specified directories (and their subdirectories). If no directory is specified, it may fail to start or deny all requests (depending on implementation version).

### Command Line Options

You can configure the behavior of the daemon with the following command line options:

* `--addr=<ADDR>` | `-a <ADDR>`
  * Network address to listen.
  * Supports TCP (e.g. `:1186`, `127.0.0.1:1186`,  `[::1]:1186`) and UNIX Domain Socket (e.g. `unix:///var/run/estelled.sock`).
  * Default: `:1186`
* `--cache-dir=<PATH TO DIR>` | `-d <PATH TO DIR>`
  * Directory to cache thumbnails.
  * Default: `$HOME/.cache/estelled`
  * For system-wide configuration, `/var/cache/estelled` is recommended.
* `--cache-limit=<SIZE>` | `-l <SIZE>`
  * Maximum size of cache directory. Supports units like `KB`, `MB`, `GB`.
  * Default: `1GB`
* `--gc-high-ratio=<RATIO>`
  * The threshold ratio of cache usage to start Garbage Collection.
  * Value must be between 0.0 and 1.0.
  * Default: `0.90` (90%)
* `--gc-low-ratio=<RATIO>`
  * The target ratio of cache usage to stop Garbage Collection.
  * Value must be between 0.0 and 1.0.
  * Default: `0.75` (75%)

## How to Use

Estelle is a HTTP server, so that you can call it by just sending HTTP request.
For example:

    curl http://localhost:1186/get?source=/absolute/path/to/image/file&size=400x400

This will return a single line of string as the response body, that is the file
path of thumbnail you want.

### Commands

#### `/get`

* Method: GET / POST

`get` returns the absolute file path of thumbnail of the specified image.
If the thumbnail does not exist yet, Estelled generates it on the fly. That means,
it will block until the thumbnail is generated. If you do not want to block,
please use `/queue` instead.

An original image is specified by `source` patameter.

For example, if you want thumbnail of `/foo/bar/baz.jpg`, you can request like this:

    curl http://localhost:1186/get?source=/foo/bar/baz.jpg&size=400x300&overflow=fill

Here, `size` specifies thumbnail size and `overflow` specifies how to treat different aspect ratio.
See "Query Parameters" below for details.

#### `/queue`

* Method: GET / POST

Request to make thumbnail. Thumbnailing task is queued and the response will be
returned immediately. The thumbnailing task is executed in background in order.

If the thumbnailing task is successfully queued, `/queue` will return `202 Accepted`.
If the thumbnail already exists, it will return `200 OK`.

#### Query Parameters

* `source`
  * Path to image file
  * This parameter is required. If this is missing, Estelled returns `400 Bad Request`.
  * The path must be absolute path. If relative path is passed, it is treated as relative path from root directory.
  * **Security**: The path must be inside one of the allowed directories specified at startup. Otherwise `403 Forbidden` will be returned.
  * If the file specified by this parameter is not exists or not an image file, Estelled returns `404 Not Found`.
* `size`
  * Size of the generated thumbnail
  * Default: `85x85`
* `overflow`
  * When the aspect ratio of `size` differs from the one of the original file, this option specifies how to generate the thumbnail.
  * One of these:
    * `fill`: resizes the image with regarding `size` as maximum width and height, and fills background with white.
    * `fit`: resizes the image with regarding `size` as minimum width and height, and cut out extra edges as it fits the specified `size`.
    * `shrink`: resizes the image with regarding `size` as maximum width and height. The resulted thumbnail is smaller than `size`.
  * Default: `fill`
* `format`
  * Image format of the output thumbnail
  * One of: `jpg`, `png`, `webp`
  * Default: `jpg`

## Caching

Estelle caches generated thumbnails in a directory specified by `--cache-dir`, and manages the total size of the cache directory.

When the total size exceeds the limit specified by `--cache-limit`, Estelle automatically removes old thumbnails to free up space. This Garbage Collection (GC) uses **Random Sampling LRU (Approximated LRU)** strategy. This means that while it prioritizes removing least recently used files, it relies on random sampling to avoid performance overhead, so strict LRU order is not guaranteed.

## Term of Use

This software is distributed under the revised BSD License.

Copyright (c) 2014-2026, Daisuke (yet another) Maki All rights reserved.
