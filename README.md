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
  * Limited Support: Windows, macOS (and other non-Linux systems)
    * **Warning**: On non-Linux systems, GC relies on `mtime` (Modification Time) instead of `atime` (Access Time) because `atime` is not reliably available or updated. This means cache eviction might not perfectly follow LRU (Least Recently Used) strategy.
* **Go Runtime**: 1.25 or later
* **File System**:
  * Recommended: ext4, XFS, Btrfs, F2FS (must support nanosecond resolution timestamps)
  * Limited Support: ext3, FAT32/exFAT
    * **Warning**: On file systems without nanosecond timestamp support, modifications made within the same second may be ignored by the cache system.
* **Runtime Dependency**: `vipsthumbnail` command
  * usually part of `libvips-tools` or `libvips-utils` package.

## How to Install

Since Estelle uses `vipsthumbnail` command to generate thumbnails, you need to install it.
You can install it with package manager like `apt` or `yum`:

```bash
apt install libvips-tools
```

Or, you need to install it from source.
Estelle just executes `vipsthumbnail` command, so please make sure it is in your `$PATH`.

Estelle is implemented in Go. You need to install [Go tools](http://golang.org/doc/install).
Then, just get Estelle:

```bash
go install github.com/Maki-Daisuke/estelle/cmd/estelled@latest
```

Or, you can clone the repository and build it:

```bash
git clone https://github.com/Maki-Daisuke/estelle.git
cd estelle/cmd/estelled
go build -o path/to/estelled
```

That's it! Now you have a binary called `estelled`.

## Run Estelled

```bash
ESTELLE_ALLOWED_DIRS="/path/to/allowed/images:/another/path" ./estelled
```

This command starts Estelle daemon. It starts listening TCP port specified by
`ESTELLE_ADDR` environment variable.

You MUST specify at least one directory via `ESTELLE_ALLOWED_DIRS` environment variable. Estelle will only allow access to images within these specified directories (and their subdirectories). If no directory is specified, it will exit with error.

### Configuration (Environment Variables)

Estelle follows the [Twelve-Factor App](https://12factor.net/config) methodology and stores configuration in environment variables.

You can configure the behavior of the daemon with the following environment variables:

* `ESTELLE_ADDR`
  * Network address to listen.
  * Supports TCP (e.g. `:1186`, `127.0.0.1:1186`,  `[::1]:1186`) and UNIX Domain Socket (e.g. `unix:///var/run/estelled.sock`).
  * Default: `:1186`
* `ESTELLE_ALLOWED_DIRS`
  * List of directories to allow access, separated by OS-specific path list separator (e.g. `:` on Linux/Unix, `;` on Windows).
  * Example (Linux): `/var/images:/home/user/images`
  * **Required**.
* `ESTELLE_CACHE_DIR`
  * Directory to cache thumbnails.
  * Default: `$HOME/.cache/estelled` (on Linux/Mac) or `%USERPROFILE%\.cache\estelled` (on Windows)
  * For system-wide configuration, `/var/cache/estelled` is recommended.
* `ESTELLE_CACHE_LIMIT`
  * Maximum size of cache directory. Supports units like `KB`, `MB`, `GB`.
  * Default: `1GB`
* `ESTELLE_GC_HIGH_RATIO`
  * The threshold ratio of cache usage to start Garbage Collection.
  * Value must be between 0.0 and 1.0.
  * Default: `0.90` (90%)
* `ESTELLE_GC_LOW_RATIO`
  * The target ratio of cache usage to stop Garbage Collection.
  * Value must be between 0.0 and 1.0.
  * Default: `0.75` (75%)

## How to Use

Estelle is a HTTP server, so that you can call it by just sending HTTP request.
For example:

```bash
curl http://localhost:1186/get?source=/absolute/path/to/image/file&size=400x400
```

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

```bash
curl http://localhost:1186/get?source=/foo/bar/baz.jpg&size=400x300&mode=crop
```

Here, `size` specifies thumbnail size and `mode` specifies how to treat different aspect ratio.
See "Query Parameters" below for details.

#### `/queue`

* Method: GET / POST

Request to make thumbnail. Thumbnailing task is queued and the response will be
returned immediately. The thumbnailing task is executed in background in order.

If the thumbnailing task is successfully queued, `/queue` will return `202 Accepted`.
If the thumbnail already exists, it will return `200 OK` with the path to the thumbnail in the response body.

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
* `mode`
  * Specifies how to resize/crop the image to match the `size`.
  * One of these:
    * `crop`: (Default) Resizes the image to fill the specified `size` and crops excess. Smart crop (using `vipsthumbnail --smartcrop`) is applied to keep interesting parts.
    * `shrink`: Resizes the image to fit within the `size`. Aspect ratio is preserved. Result may be smaller than `size`.
    * `stretch`: Forces the image to exactly match `size` by ignoring aspect ratio.
  * Default: `crop`
* `format`
  * Image format of the output thumbnail
  * One of: `jpg`, `png`, `webp`
  * Default: `jpg`

## Caching

Estelle caches generated thumbnails in a directory specified by `ESTELLE_CACHE_DIR`, and manages the total size of the cache directory.

When the total size exceeds the limit specified by `ESTELLE_CACHE_LIMIT`, Estelle automatically removes old thumbnails to free up space. This Garbage Collection (GC) uses **Random Sampling LRU (Approximated LRU)** strategy. This means that while it prioritizes removing least recently used files, it relies on random sampling to avoid performance overhead, so strict LRU order is not guaranteed.

## Term of Use

This software is distributed under the revised BSD License.

Copyright (c) 2014-2026, Daisuke (yet another) Maki All rights reserved.
