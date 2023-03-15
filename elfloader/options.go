// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package elfloader

import (
	"fmt"
	"os"
	"path/filepath"

	"kraftkit.sh/kconfig"
	"kraftkit.sh/unikraft"
	"kraftkit.sh/unikraft/core"
	"kraftkit.sh/unikraft/lib"
	"kraftkit.sh/unikraft/target"
	"kraftkit.sh/unikraft/app"
)

type ELFLoaderOption func (*ELFLoader) error

// NewApplicationFromOptions accepts a series of options and returns a rendered
// *ApplicationConfig structure
func NewELFLoaderFromOptions(aopts ...ELFLoaderOption) (app.Application, error) {
	var err error
	ac := &ELFLoader{
		configuration: kconfig.KeyValueMap{},
	}

	for _, o := range aopts {
		if err := o(ac); err != nil {
			return nil, fmt.Errorf("could not apply option: %v", err)
		}
	}

	if ac.name != "" {
		ac.configuration.Set(unikraft.UK_NAME, ac.name)
	}

	if ac.outDir == "" {
		if ac.workingDir == "" {
			ac.workingDir, err = os.Getwd()
			if err != nil {
				return nil, err
			}
		}

		ac.outDir = filepath.Join(ac.workingDir, unikraft.BuildDir)
	}

	if len(ac.unikraft.Source()) > 0 {
		if p, err := os.Stat(ac.unikraft.Source()); err == nil && p.IsDir() {
			ac.configuration.Set(unikraft.UK_BASE, ac.unikraft.Source())
		}
	}

	return ac, nil
}

// WithName sets the application component name
func WithName(name string) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.name = name
		return nil
	}
}

// WithVersion sets the application version
func WithVersion(version string) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.version = version
		return nil
	}
}

// WithWorkingDir sets the application's working directory
func WithWorkingDir(workingDir string) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.workingDir = workingDir
		return nil
	}
}

// WithSource sets the library's source which indicates where it was retrieved
// and in component context and not the origin.
func WithSource(source string) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.source = source
		return nil
	}
}

// WithFilename sets the application's file name
func WithFilename(filename string) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.filename = filename
		return nil
	}
}

// WithOutDir sets the application's output directory
func WithOutDir(outDir string) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.outDir = outDir
		return nil
	}
}

// WithUnikraft sets the application's core
func WithUnikraft(unikraft core.UnikraftConfig) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.unikraft = unikraft
		return nil
	}
}

// WithLibraries sets the application's library list
func WithLibraries(libraries lib.Libraries) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.libraries = libraries
		return nil
	}
}

// WithTargets sets the application's target list
func WithTargets(targets target.Targets) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.targets = targets
		return nil
	}
}

// WithKraftfiles sets the application's kraft yaml files
func WithKraftfiles(kraftfiles []string) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		ac.kraftfiles = kraftfiles
		return nil
	}
}

// WithConfiguration sets the application's kconfig list
func WithConfiguration(config ...*kconfig.KeyValue) ELFLoaderOption {
	return func(ac *ELFLoader) error {
		if ac.configuration == nil {
			ac.configuration = kconfig.KeyValueMap{}
		}

		ac.configuration.Override(config...)
		return nil
	}
}
