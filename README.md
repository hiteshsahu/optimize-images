# Optimize Images
[![🖼️ CI](https://github.com/hiteshsahu/optimize-images/actions/workflows/ci.yml/badge.svg)](https://github.com/hiteshsahu/optimize-images/actions/workflows/ci.yml)
[![🚀 Release](https://github.com/hiteshsahu/optimize-images/actions/workflows/release.yml/badge.svg)](https://github.com/hiteshsahu/optimize-images/actions/workflows/release.yml)

A small CLI that shrinks jpg/jpeg/png images in a folder by re-encoding them
as **jpg** (pure Go, no external tools) or **webp** (shells out to `cwebp`,
since Go's standard library has no WebP encoder).

### Why this tool?

It is a tedious task to optimize screenshots &  images generated from OpenAI's DALL·E or Midjourney, and this tool is meant to automate that process.

This tool is not meant to be a replacement for professional image optimization tools like ImageMagick or Photoshop, but rather a quick and easy way to reduce the size of images in bulk.

## Install

```bash
go build -o optimize-images .

# or, once pushed to GitHub:
go install github.com/HiteshSahu/optimize-images@latest
```

## Usage

```
optimize-images <folder> [-q QUALITY] [-r] [--replace] [--format jpg|webp]
```

| Flag                 | Meaning                                              | Default              |
|----------------------|------------------------------------------------------|----------------------|
| `--format jpg\|webp` | Output format                                        | `jpg`                |
| `-q <QUALITY>`       | Output quality 0-100                                 | `85`                 |
| `-r`                 | Recurse into subfolders                              | off (top-level only) |
| `--replace`          | Remove the source file after a successful conversion | off                  |

A source file already in the target format (e.g. a `.jpg`/`.jpeg` file when
`--format=jpg`) is left alone — it's not re-encoded or duplicated.

JPEG has no alpha channel, so `--format jpg` flattens transparent pixels onto
a white background before encoding. 

`--format webp` keeps transparency intact
(via `cwebp`), which is why an image meant to sit on any background (a logo,
say) belongs in webp, not jpg.

## Examples


```bash
# Recursively optimize all images in the `img` folder, outputting to `jpg` at quality 85
optimize-images img -r

# Optimize all images at root level in the `img` folder, outputting to `jpg` at quality 90
optimize-images img -q 90 

# Recursively optimize& replace all images in `img` folder, with jpg quality 90
optimize-images img/screen-shots -q 90 -r --replace

# Optimize all images in the `img` folder, outputting to `webp` keping original images
optimize-images img/infographics --format webp
```
