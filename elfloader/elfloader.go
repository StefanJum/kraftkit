// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2022, Unikraft GmbH and The KraftKit Authors.
// Licensed under the BSD-3-Clause License (the "License").
// You may not use this file except in compliance with the License.
package elfloader

import (
	"fmt"
	"context"
	"debug/elf"
	//"fmt"
	//"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"

	"kraftkit.sh/kconfig"
	"kraftkit.sh/unikraft/app"
	"kraftkit.sh/log"
	"kraftkit.sh/make"
	"kraftkit.sh/unikraft"
	"kraftkit.sh/unikraft/component"
	"kraftkit.sh/unikraft/core"
	"kraftkit.sh/unikraft/lib"
	"kraftkit.sh/unikraft/target"
	"kraftkit.sh/unikraft/template"
)

type ELFLoader struct {
	name          string
	version       string
	source        string
	path          string
	workingDir    string
	outDir        string
	filename      string
	binname       string
	kraftfiles    []string
	configuration kconfig.KeyValueMap
	unikraft      core.UnikraftConfig
	libraries     lib.Libraries
	targets       target.Targets
}

func (elfloader ELFLoader) Name() string {
	return elfloader.name
}

func (elfloader ELFLoader) Version() string {
	return elfloader.version
}

func (elfloader ELFLoader) Source() string {
	return elfloader.source
}

func (elfloader ELFLoader) WorkingDir() string {
	return elfloader.workingDir
}

func (elfloader ELFLoader) Path() string {
	return elfloader.workingDir
}

func (elfloader ELFLoader) Filename() string {
	return elfloader.filename
}

func (elfloader ELFLoader) Unikraft() core.Unikraft {
	return elfloader.unikraft
}

func (elfloader ELFLoader) OutDir() string {
	return elfloader.outDir
}

func (elfloader ELFLoader) Template() template.Template {
	return nil
}

func (elfloader ELFLoader) Libraries(ctx context.Context) (lib.Libraries, error) {
	uklibs, err := elfloader.unikraft.Libraries(ctx)
	if err != nil {
		return nil, err
	}

	libs := elfloader.libraries

	for _, uklib := range uklibs {
		libs[uklib.Name()] = uklib
	}

	return libs, nil
}

func (elfloader ELFLoader) Targets() target.Targets {
	return elfloader.targets
}

func (elfloader ELFLoader) Extensions() component.Extensions {
	return nil
}

func (elfloader ELFLoader) Kraftfiles() []string {
	return elfloader.kraftfiles
}

func (elfloader ELFLoader) MergeTemplate(context.Context, app.Application) (app.Application, error) {
	return nil, nil
}

func (elfloader ELFLoader) KConfigFile(tc target.Target) string {
	k := filepath.Join(elfloader.workingDir, kconfig.DotConfigFileName)

	if tc != nil {
		k += "." + filepath.Base(tc.Kernel())
	}

	return k
}

func (elfloader ELFLoader) IsConfigured(tc target.Target) bool {
	f, err := os.Stat(elfloader.KConfigFile(tc))
	return err == nil && !f.IsDir() && f.Size() > 0
}

func (elfloader ELFLoader) MakeArgs(tc target.Target) (*core.MakeArgs, error) {
	var libraries []string

	// TODO: This is a temporary solution to fix an ordering issue with regard to
	// syscall availability from a libc (which should be included first).  Long-term
	// solution is to determine the library order by generating a DAG via KConfig
	// parsing.
	unformattedLibraries := lib.Libraries{}
	for k, v := range elfloader.libraries {
		unformattedLibraries[k] = v
	}

	// All supported libCs right now
	if unformattedLibraries["musl"] != nil {
		libraries = append(libraries, unformattedLibraries["musl"].Path())
		delete(unformattedLibraries, "musl")
	} else if unformattedLibraries["newlib"] != nil {
		libraries = append(libraries, unformattedLibraries["newlib"].Path())
		delete(unformattedLibraries, "newlib")
		if unformattedLibraries["pthread-embedded"] != nil {
			libraries = append(libraries, unformattedLibraries["pthread-embedded"].Path())
			delete(unformattedLibraries, "pthread-embedded")
		}
	}

	for _, library := range unformattedLibraries {
		if !library.IsUnpacked() {
			return nil, fmt.Errorf("cannot determine library \"%s\" path without component source", library.Name())
		}

		libraries = append(libraries, library.Path())
	}

	// TODO: Platforms & architectures

	args := &core.MakeArgs{
		OutputDir:      elfloader.outDir,
		ApplicationDir: elfloader.workingDir,
		LibraryDirs:    strings.Join(libraries, core.MakeDelimeter),
		ConfigPath:     elfloader.KConfigFile(tc),
	}

	if tc != nil {
		args.Name = tc.Name()
	}

	return args, nil
}

func (elfloader ELFLoader) Make(ctx context.Context, tc target.Target, mopts ...make.MakeOption) error {
	mopts = append(mopts,
		make.WithDirectory(elfloader.unikraft.Path()),
		make.WithNoPrintDirectory(true),
	)

	args, err := elfloader.MakeArgs(tc)
	if err != nil {
		return err
	}

	m, err := make.NewFromInterface(*args, mopts...)
	if err != nil {
		return err
	}

	// Unikraft currently requires each application to have a `Makefile.uk`
	// located within the working directory.  Create it if it does not exist:
	makefile_uk := filepath.Join(elfloader.WorkingDir(), unikraft.Makefile_uk)
	if _, err := os.Stat(makefile_uk); err != nil && os.IsNotExist(err) {
		if _, err := os.OpenFile(makefile_uk, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666); err != nil {
			return fmt.Errorf("could not create application %s: %v", makefile_uk, err)
		}
	}

	return m.Execute(ctx)
}

func (elfloader ELFLoader) SyncConfig(ctx context.Context, tc target.Target, mopts ...make.MakeOption) error {
	return elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget("syncconfig"),
		)...,
	)
}

func (elfloader ELFLoader) DefConfig(ctx context.Context, tc target.Target, extra kconfig.KeyValueMap, mopts ...make.MakeOption) error {
	values := kconfig.KeyValueMap{}
	values.OverrideBy(elfloader.KConfig())

	if tc != nil {
		values.OverrideBy(tc.KConfig())
	}

	if extra != nil {
		values.OverrideBy(extra)
	}

	for _, kv := range values {
		log.G(ctx).WithFields(logrus.Fields{
			kv.Key: kv.Value,
		}).Debugf("defconfig")
	}

	// Write the configuration to a temporary file
	tmpfile, err := os.CreateTemp("", elfloader.Name()+"-config*")
	if err != nil {
		return fmt.Errorf("could not create temporary defconfig file: %v", err)
	}

	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// Save and sync the file to the temporary file
	tmpfile.Write([]byte(values.String()))
	tmpfile.Sync()

	// TODO: This make dependency should be upstreamed into the Unikraft core as a
	// dependency of `make defconfig`
	if err := elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget(fmt.Sprintf("%s/Makefile", elfloader.outDir)),
			make.WithProgressFunc(nil),
		)...,
	); err != nil {
		return err
	}

	return elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget("defconfig"),
			make.WithVar("UK_DEFCONFIG", tmpfile.Name()),
		)...,
	)
}

func (elfloader ELFLoader) Configure(ctx context.Context, tc target.Target, mopts ...make.MakeOption) error {
	return elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget("configure"),
		)...,
	)
}

func (elfloader ELFLoader) Prepare(ctx context.Context, tc target.Target, mopts ...make.MakeOption) error {
	return elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget("prepare"),
		)...,
	)
}

func (elfloader ELFLoader) Clean(ctx context.Context, tc target.Target, mopts ...make.MakeOption) error {
	return elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget("clean"),
		)...,
	)
}

func (elfloader ELFLoader) Properclean(ctx context.Context, tc target.Target, mopts ...make.MakeOption) error {
	return elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget("properclean"),
		)...,
	)
}

func (elfloader ELFLoader) Fetch(ctx context.Context, tc target.Target, mopts ...make.MakeOption) error {
	return elfloader.Make(
		ctx,
		tc,
		append(mopts,
			make.WithTarget("fetch"),
		)...,
	)
}

func (elfloader ELFLoader) Set(context.Context, target.Target, ...make.MakeOption) error {
	return nil
}

func (elfloader ELFLoader) Unset(context.Context, target.Target, ...make.MakeOption) error {
	return nil
}

func (elfloader ELFLoader) Build(ctx context.Context, tc target.Target, opts ...app.BuildOption) error {
	bopts := &BuildOptions{}
	//appopts := &app.BuildOption(*bopts)
	for _, o := range opts {
		err := o(bopts)
		if err != nil {
			return fmt.Errorf("could not apply build option %v", err)
		}
	}

	if !elfloader.unikraft.IsUnpacked() {
		return fmt.Errorf("cannot build without Unikraft core component source. Please run `kraft pkg pull` and try again")
	}

	bopts.mopts = append(bopts.mopts, []make.MakeOption{
		make.WithProgressFunc(bopts.onProgress),
	}...)

	if !bopts.noPrepare {
		if err := elfloader.Prepare(
			ctx,
			tc,
			append(
				bopts.mopts,
				make.WithProgressFunc(nil),
			)...); err != nil {
			return err
		}
	}

	return elfloader.Make(ctx, tc, bopts.mopts...)
}

func (elfloader ELFLoader) LibraryNames() []string {
	var names []string
	for k := range elfloader.libraries {
		names = append(names, k)
	}

	sort.Strings(names)

	return names
}

func (elfloader ELFLoader) TargetNames() []string {
	var names []string
	for _, k := range elfloader.targets {
		names = append(names, k.Name())
	}

	sort.Strings(names)

	return names
}

func (elfloader ELFLoader) TargetByName(name string) (target.Target, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("no target name specified in lookup")
	}

	for _, k := range elfloader.targets {
		if k.Name() == name {
			return k, nil
		}
	}

	return nil, fmt.Errorf("unknown target: %s", name)
}

func (elfloader ELFLoader) Components() ([]component.Component, error) {
	components := []component.Component{
		elfloader.unikraft,
	}

	for _, library := range elfloader.libraries {
		components = append(components, library)
	}

	// TODO: Get unique components from each target.  A target will contain at
	// least two components: the architecture and the platform.  Both of these
	// components can stem from the Unikraft core (in the case of built-in
	// architectures and components).
	// for _, targ := range app.Targets {
	// 	components = append(components, targ)
	// }

	return components, nil
}

func (elfloader ELFLoader) WithTarget(targ target.Target) (app.Application, error) {
	ret := elfloader
	ret.targets = target.Targets{targ.(target.TargetConfig)}
	return ret, nil
}

func (elfloader ELFLoader) KConfig() kconfig.KeyValueMap {
	all := kconfig.KeyValueMap{}
	all.OverrideBy(elfloader.unikraft.KConfig())

	for _, library := range elfloader.libraries {
		all.OverrideBy(library.KConfig())
	}

	return all
}

func (elfloader ELFLoader) KConfigTree(env ...*kconfig.KeyValue) (*kconfig.KConfigFile, error) {
	config_uk := filepath.Join(elfloader.workingDir, unikraft.Config_uk)
	if _, err := os.Stat(config_uk); err != nil {
		return nil, fmt.Errorf("could not read component Config.uk: %v", err)
	}

	return kconfig.Parse(config_uk, elfloader.KConfig().Override(env...).Slice()...)
}

func (elfloader ELFLoader) PrintInfo(ctx context.Context) string {
	tree := treeprint.NewWithRoot(component.NameAndVersion(elfloader))

	uk := tree.AddBranch(component.NameAndVersion(elfloader.unikraft))
	uklibs, err := elfloader.unikraft.Libraries(ctx)
	if err == nil {
		for _, uklib := range uklibs {
			uk.AddNode(uklib.Name())
		}
	}

	if len(elfloader.libraries) > 0 {
		libraries := tree.AddBranch(fmt.Sprintf("libraries (%d)", len(elfloader.libraries)))
		for _, library := range elfloader.libraries {
			libraries.AddNode(component.NameAndVersion(library))
		}
	}

	if len(elfloader.targets) > 0 {
		targets := tree.AddBranch(fmt.Sprintf("targets (%d)", len(elfloader.targets)))
		for _, t := range elfloader.targets {
			branch := targets.AddBranch(component.NameAndVersion(t))
			branch.AddNode(fmt.Sprintf("architecture: %s", component.NameAndVersion(t.Architecture())))
			branch.AddNode(fmt.Sprintf("platform:     %s", component.NameAndVersion(t.Platform())))
		}
	}

	return tree.String()
}

func (elfloader ELFLoader) Type() unikraft.ComponentType {
	return unikraft.ComponentTypeApp
}

var _ app.Application = (*ELFLoader)(nil)

func New(bin string, eopts ...ELFLoaderOption) (app.Application, error) {
	f, err := os.Open(bin)

	if err != nil {
		return nil, err
	}

	_elf, err := elf.NewFile(f)

	if err != nil {
		return nil, err
	}

	fmt.Println(_elf.Machine.String())

	elfloader, err := NewELFLoaderFromOptions(eopts...)

	if err != nil {
		return nil, err
	}

	return elfloader, err
}
