// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command go-overlay prepares the `go build -overlay` file used for release
// builds and verifies its effect on the built executables.
//
// Go 1.26 added a 32 MiB package-level scratch buffer to the FIPS 140-3
// entropy source (crypto/internal/fips140/drbg.memory). On Windows it is part
// of the writable data section of the executable image, which costs 32 MiB of
// commit charge per process even when FIPS 140-3 mode is disabled. The overlay
// replaces that file with a copy that allocates the buffer on first use.
// See https://github.com/golang/go/issues/81956.
//
// Usage:
//
//	go run ./tools/go-overlay generate <dir>
//	go run ./tools/go-overlay verify <executable>...
//
// generate writes overlay.json and the patched source file to dir. Pass
// -overlay=<dir>/overlay.json to go build. verify fails if an executable still
// contains the package-level buffer.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"debug/pe"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// targetFile is the Go standard library file replaced by the overlay.
	targetFile = "src/crypto/internal/fips140/drbg/entropy_fips140.go"

	// targetSHA256 is the checksum of targetFile in the Go version the
	// overlay was written for. A Go update that changes the file must fail
	// the build, so the overlay is re-checked or dropped once Go fixes
	// golang/go#81956.
	targetSHA256 = "bd3c834a29e31c93e56d81da54d4a874561814a86550de1456c8d6d40a9c5fda"

	// bufferSymbol is the package-level buffer removed by the overlay.
	bufferSymbol = "crypto/internal/fips140/drbg.memory"
)

//go:embed entropy_fips140.go.overlay
var patchedSource []byte

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "go-overlay:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: go-overlay generate <dir> | verify <executable> [<executable> ...]")
	}

	switch args[0] {
	case "generate":
		if len(args) != 2 {
			return errors.New("usage: go-overlay generate <dir>")
		}

		return generate(context.Background(), args[1])
	case "verify":
		for _, path := range args[1:] {
			if err := verify(path); err != nil {
				return err
			}
		}

		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func generate(ctx context.Context, dir string) error {
	goroot, err := goEnv(ctx, "GOROOT")
	if err != nil {
		return err
	}

	target := filepath.Join(goroot, filepath.FromSlash(targetFile))

	original, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("read %s: %w", target, err)
	}

	sum := sha256.Sum256(original)
	if got := hex.EncodeToString(sum[:]); got != targetSHA256 {
		return fmt.Errorf("%s has checksum %s, expected %s: the Go toolchain changed this file, "+
			"so update tools/go-overlay/entropy_fips140.go.overlay from the new version "+
			"or remove the overlay if golang/go#81956 is fixed", target, got, targetSHA256)
	}

	// The go command rejects overlays for files in the module cache, which is
	// where toolchains downloaded through GOTOOLCHAIN are stored.
	gomodcache, err := goEnv(ctx, "GOMODCACHE")
	if err != nil {
		return err
	}

	if gomodcache != "" && isWithin(goroot, gomodcache) {
		return fmt.Errorf("GOROOT %s is inside GOMODCACHE %s, where go build does not allow overlays; "+
			"install the Go version from go.mod instead of using a downloaded toolchain", goroot, gomodcache)
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}

	//nolint:gosec // The directory is chosen by the caller of this build tool.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	patched := filepath.Join(dir, "entropy_fips140.go")
	//nolint:gosec // Build input, not a secret.
	if err := os.WriteFile(patched, patchedSource, 0o644); err != nil {
		return err
	}

	overlay, err := json.MarshalIndent(map[string]map[string]string{
		"Replace": {target: patched},
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "overlay.json"), append(overlay, '\n'), 0o644) //nolint:gosec // Build input, not a secret.
}

func verify(path string) error {
	f, err := pe.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	if len(f.Symbols) == 0 {
		return fmt.Errorf("%s has no symbol table, cannot verify the overlay", path)
	}

	for _, sym := range f.Symbols {
		if sym.Name == bufferSymbol {
			return fmt.Errorf("%s contains %s: the go-overlay was not applied", path, bufferSymbol)
		}
	}

	return nil
}

func goEnv(ctx context.Context, name string) (string, error) {
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", "env", name)
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go env %s: %w: %s", name, err, stderr.String())
	}

	return strings.TrimSpace(string(out)), nil
}

func isWithin(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)

	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
