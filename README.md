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
optimize-images <folder> [-q QUALITY] [-r] [--delete-originals] [--format jpg|webp]
```

| Flag                 | Meaning                                              |
|----------------------|------------------------------------------------------|
| `-q QUALITY`         | Output quality 0-100 (default: 85)                   |
| `-r`                 | Recurse into subfolders (default: top-level only)    |
| `--delete-originals` | Remove the source file after a successful conversion |
| `--format jpg\|webp` | Output format (default: jpg)                         |

A source file already in the target format (e.g. a `.jpg`/`.jpeg` file when
`--format=jpg`) is left alone — it's not re-encoded or duplicated.

JPEG has no alpha channel, so `--format jpg` flattens transparent pixels onto
a white background before encoding. 

`--format webp` keeps transparency intact
(via `cwebp`), which is why an image meant to sit on any background (a logo,
say) belongs in webp, not jpg.

## Examples

```bash
optimize-images img
optimize-images img/screen-shots -q 90 -r --delete-originals
optimize-images img/infographics --format webp
```
