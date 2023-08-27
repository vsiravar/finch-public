// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package config

import (
	"fmt"

	"github.com/lima-vm/lima/pkg/limayaml"
	"github.com/xorcare/pointer"
)

func (lca *limaConfigApplier) applyInit(limaCfg *limayaml.LimaYAML) (*limayaml.LimaYAML, error) {
	// Check if system supports wsl2
	if !SupportsWSL2() {
		return nil, fmt.Errorf("wsl2 is not supported by your system")
	}
	if *lca.cfg.VMType == "wsl2" {
		limaCfg.MountType = pointer.String("wsl2")
		limaCfg.VMType = lca.cfg.VMType
	} else {
		return nil, fmt.Errorf("unsupported vm type \"%s\" for windows", *lca.cfg.VMType)
	}
	return limaCfg, nil
}
