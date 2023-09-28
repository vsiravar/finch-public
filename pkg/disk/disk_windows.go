// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build windows
// +build windows

package disk

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"github.com/runfinch/finch/pkg/winutil"
)

// EnsureUserDataDisk checks the current disk configuration and fixes it if needed.
// on Windows, this is a no-op.
func (m *userDataDiskManager) EnsureUserDataDisk() error {
	diskPath := m.finch.UserDataDiskPath(m.rootDir)
	disksDir := filepath.Dir(diskPath)

	m.logger.Debugf("diskPath: %s", diskPath)
	m.logger.Debugf("disksDir: %s", disksDir)

	if _, err := m.fs.Stat(disksDir); errors.Is(err, fs.ErrNotExist) {
		if err := m.fs.MkdirAll(disksDir, 0o755); err != nil {
			return fmt.Errorf("could not create persistent disk directory: %w", err)
		}
	}

	if _, err := m.fs.Stat(diskPath); errors.Is(err, fs.ErrNotExist) {
		if err := m.createDisk(diskPath); err != nil {
			return fmt.Errorf("could not create persistent disk: %w", err)
		}
	}

	if err := m.attachDisk(diskPath); err != nil {
		return fmt.Errorf("could not attach persistent disk: %w", err)
	}

	return nil
}

// DetachUserDataDisk unmounts the disk in wsl.
func (m *userDataDiskManager) DetachUserDataDisk() error {
	cmd := m.ecc.Create(
		"wsl.exe",
		"--unmount",
		`\\?\`+m.finch.UserDataDiskPath(m.rootDir),
	)

	m.logger.Debugf("running attach cmd: %v", cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to attach disk: %w, command output: %s", err, out)
	}

	return nil
}

func (m *userDataDiskManager) diskExists(diskPath string) (bool, error) {
	disksDir := filepath.Dir(diskPath)
	_, err := m.fs.Stat(filepath.Dir(diskPath))
	if errors.Is(err, fs.ErrNotExist) {
		if err := m.fs.MkdirAll(disksDir, 0o755); err != nil {
			return false, fmt.Errorf("could not create persistent disk directory: %w", err)
		}
	}

	return true, nil
}

//go:embed createDiskAdmin.TEMPLATE.ps1
var createDiskTmpl string

type createDiskOpts struct {
	Path         string
	Size         int64
	TempFilePath string
}

func (m *userDataDiskManager) createDisk(diskPath string) error {
	size, err := sizeInMB()
	if err != nil {
		return fmt.Errorf("failed to get disk size: %w", err)
	}

	m.logger.Infof("creating disk at path: %s", diskPath)

	tempOut, _ := afero.TempFile(m.fs, "", "finchCreateDiskOutput*")
	dpgoPath := filepath.Join(string(m.finch), "bin", "dpgo.exe")

	_ = tempOut.Close()
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := winutil.RunElevated(
		dpgoPath,
		filepath.Dir(execPath),
		[]string{
			"--json",
			"disk",
			"create",
			"--path",
			diskPath,
			"--size",
			fmt.Sprint(size),
			">",
			tempOut.Name(),
			"2>&1",
		},
	); err != nil {
		return fmt.Errorf("failed to run command: %s", err)
	}

	tempOutContents, _ := afero.ReadFile(m.fs, tempOut.Name())
	m.logger.Infof("createDisk out: %s", string(tempOutContents))

	_ = m.fs.Remove(tempOut.Name())

	if err != nil {
		return fmt.Errorf("failed to create disk: %w, command output: %s", err, tempOutContents)
	}

	return nil
}

func (m *userDataDiskManager) attachDisk(diskPath string) error {
	m.logger.Infof("attaching disk at path: %s", diskPath)

	cmd := m.ecc.Create(
		"wsl.exe",
		"--mount",
		"--bare",
		"--vhd",
		diskPath,
	)

	m.logger.Debugf("running attach cmd: %v", cmd)

	out, err := cmd.CombinedOutput()
	outDecoded, _ := FromUTF16leToString(bytes.NewBuffer(out))
	if err != nil {
		return fmt.Errorf("failed to attach disk: %w, command output: %s", err, outDecoded)
	}

	return nil
}

func sizeInMB() (int64, error) {
	sizeB, err := diskSize()
	if err != nil {
		return 0, err
	}

	return sizeB / 1048576, nil
}

// FromUTF16le returns an io.Reader for UTF16le data.
// Windows uses little endian by default, use unicode.UseBOM policy to retrieve BOM from the text,
// and unicode.LittleEndian as a fallback.
func FromUTF16le(r io.Reader) io.Reader {
	o := transform.NewReader(r, unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder())
	return o
}

// FromUTF16leToString reads from Unicode 16 LE encoded data from an io.Reader and returns a string.
func FromUTF16leToString(r io.Reader) (string, error) {
	out, err := io.ReadAll(FromUTF16le(r))
	if err != nil {
		return "", err
	}

	return string(out), nil
}
