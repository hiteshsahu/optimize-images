// Command optimize-images shrinks jpg/jpeg/png images in a folder by
// re-encoding them as jpg (pure Go, no external tools) or webp (shells out
// to cwebp, since Go's standard library has no WebP encoder).
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const usage = `Convert jpg/jpeg/png images in a folder to jpg or webp.

Usage:
  optimize-images <folder> [-q QUALITY] [-r] [--delete-originals] [--format jpg|webp]

Options:
  -q QUALITY           Output quality 0-100 (default: 85)
  -r                    Recurse into subfolders (default: top-level only)
  --delete-originals    Remove the source file after a successful conversion
  --format FORMAT       Output format: jpg (default, pure Go, no external
                        deps) or webp (shells out to cwebp; keeps transparency
                        that jpg would otherwise flatten away)

A source file already in the target format (e.g. a .jpg/.jpeg file when
--format=jpg) is left alone -- it's not re-encoded or duplicated.

Examples:
  optimize-images img
  optimize-images img/screen-shots -q 90 -r --delete-originals
  optimize-images img/infographics --format webp
`

type config struct {
	folder          string
	quality         int
	recursive       bool
	deleteOriginals bool
	format          string
}

var sourceExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true}

// jpgAliases are extensions that already count as "jpg" for skip purposes --
// re-encoding a .jpeg to .jpg would otherwise leave both a .jpeg and a
// same-content .jpg sitting side by side.
var jpgAliases = map[string]bool{".jpg": true, ".jpeg": true}

func parseArgs(args []string) (config, error) {
	cfg := config{quality: 85, format: "jpg"}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-q":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("-q requires a value")
			}
			q, err := strconv.Atoi(args[i])
			if err != nil || q < 0 || q > 100 {
				return cfg, fmt.Errorf("-q must be an integer 0-100, got %q", args[i])
			}
			cfg.quality = q
		case "-r":
			cfg.recursive = true
		case "--delete-originals":
			cfg.deleteOriginals = true
		case "--format":
			i++
			if i >= len(args) {
				return cfg, fmt.Errorf("--format requires a value")
			}
			if args[i] != "jpg" && args[i] != "webp" {
				return cfg, fmt.Errorf("--format must be jpg or webp, got %q", args[i])
			}
			cfg.format = args[i]
		case "-h", "--help":
			fmt.Print(usage)
			os.Exit(0)
		default:
			if cfg.folder != "" {
				return cfg, fmt.Errorf("unexpected argument: %s", args[i])
			}
			cfg.folder = args[i]
		}
	}
	if cfg.folder == "" {
		return cfg, fmt.Errorf("folder argument is required")
	}
	return cfg, nil
}

func findImages(root string, recursive bool) ([]string, error) {
	var files []string

	if recursive {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && sourceExts[strings.ToLower(filepath.Ext(path))] {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() && sourceExts[strings.ToLower(filepath.Ext(e.Name()))] {
				files = append(files, filepath.Join(root, e.Name()))
			}
		}
	}

	sort.Strings(files)
	return files, nil
}

func alreadyTargetFormat(path, format string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if format == "jpg" {
		return jpgAliases[ext]
	}
	return ext == ".webp"
}

// convertToJPEG flattens onto white before encoding -- JPEG has no alpha
// channel, and Go's jpeg encoder otherwise composites transparent pixels
// onto black, silently changing the image instead of matching the usual
// "flatten to white" convention.
func convertToJPEG(src, dst string, quality int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	flat := image.NewRGBA(bounds)
	draw.Draw(flat, bounds, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(flat, bounds, img, bounds.Min, draw.Over)

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, flat, &jpeg.Options{Quality: quality})
}

func convertToWebP(src, dst string, quality int) error {
	cmd := exec.Command("cwebp", "-quiet", "-q", strconv.Itoa(quality), src, "-o", dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cwebp: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func run(cfg config) error {
	info, err := os.Stat(cfg.folder)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("%q is not a directory", cfg.folder)
	}

	if cfg.format == "webp" {
		if _, err := exec.LookPath("cwebp"); err != nil {
			return fmt.Errorf("cwebp is not installed -- install it with 'brew install webp'")
		}
	}

	files, err := findImages(cfg.folder, cfg.recursive)
	if err != nil {
		return err
	}

	var totalBefore, totalAfter int64
	count := 0

	for _, src := range files {
		if alreadyTargetFormat(src, cfg.format) {
			fmt.Printf("Skipping (already %s): %s\n", cfg.format, src)
			continue
		}

		ext := filepath.Ext(src)
		dst := strings.TrimSuffix(src, ext) + "." + cfg.format

		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("Skipping (already exists): %s\n", dst)
			continue
		}

		beforeInfo, err := os.Stat(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", src, err)
			continue
		}

		var convErr error
		if cfg.format == "webp" {
			convErr = convertToWebP(src, dst, cfg.quality)
		} else {
			convErr = convertToJPEG(src, dst, cfg.quality)
		}
		if convErr != nil {
			fmt.Fprintf(os.Stderr, "Error converting %s: %v\n", src, convErr)
			continue
		}

		afterInfo, err := os.Stat(dst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", dst, err)
			continue
		}

		before, after := beforeInfo.Size(), afterInfo.Size()
		totalBefore += before
		totalAfter += after
		count++

		pct := 0.0
		if before > 0 {
			pct = (1 - float64(after)/float64(before)) * 100
		}
		fmt.Printf("%s: %d -> %d bytes (%.0f%% smaller)\n", src, before, after, pct)

		if cfg.deleteOriginals {
			if err := os.Remove(src); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting %s: %v\n", src, err)
			}
		}
	}

	if count == 0 {
		fmt.Printf("No convertible images found in %q.\n", cfg.folder)
		return nil
	}

	fmt.Println("---")
	fmt.Printf("Converted %d image(s): %d -> %d bytes total\n", count, totalBefore, totalAfter)
	return nil
}

func main() {
	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
