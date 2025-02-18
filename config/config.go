// SPDX-License-Identifier: BSD-3-Clause
//
// Authors: Alexander Jung <alex@unikraft.io>
//
// Copyright (c) 2022, Unikraft GmbH.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the names of its
//    contributors may be used to endorse or promote products derived from
//    this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package config

// AuthConfig represents a very abstract representation of authentication used
// by some service.  Most APIs and services which can be authenticated have the
// defined four parameters found within AuthConfig.
type AuthConfig struct {
	User      string `json:"user"       yaml:"user"       env:"KRAFTKIT_AUTH_%s_USER"`
	Token     string `json:"token"      yaml:"token"      env:"KRAFTKIT_AUTH_%s_TOKEN"`
	Endpoint  string `json:"endpoint"   yaml:"endpoint"   env:"KRAFTKIT_AUTH_%s_ENDPOINT"`
	VerifySSL bool   `json:"verify_ssl" yaml:"verify_ssl" env:"KRAFTKIT_AUTH_%s_VERIFY_SSL"`
}

type Config struct {
	NoPrompt       bool   `json:"no_prompt"        yaml:"no_prompt"                  env:"KRAFTKIT_NO_PROMPT"`
	NoParallel     bool   `json:"no_parallel"      yaml:"no_parallel"                env:"KRAFTKIT_NO_PARALLEL"`
	Emojis         bool   `json:"no_emojis"        yaml:"no_emojis"                  env:"KRAFTKIT_NO_EMOJIS"`
	Editor         string `json:"editor"           yaml:"editor,omitempty"           env:"KRAFTKIT_EDITOR"`
	Browser        string `json:"browser"          yaml:"browser,omitempty"          env:"KRAFTKIT_BROWSER"`
	GitProtocol    string `json:"git_protocol"     yaml:"git_protocol"               env:"KRAFTKIT_GIT_PROTOCOL"`
	Pager          string `json:"pager"            yaml:"pager,omitempty"            env:"KRAFTKIT_PAGER"`
	HTTPUnixSocket string `json:"http_unix_socket" yaml:"http_unix_socket,omitempty" env:"KRAFTKIT_HTTP_UNIX_SOCKET"`
	RuntimeDir     string `json:"runtime_dir"      yaml:"runtime_dir"                env:"KRAFTKIT_RUNTIME_DIR"`
	DefaultPlat    string `json:"default_plat"     yaml:"default_plat"               env:"KRAFTKIT_DEFAULT_PLAT"`
	DefaultArch    string `json:"default_arch"     yaml:"default_arch"               env:"KRAFTKIT_DEFAULT_ARCH"`
	EventsPidFile  string `json:"events_pidfile"   yaml:"events_pidfile"             env:"KRAFTKIT_EVENTS_PIDFILE"`

	Paths struct {
		Plugins   string `json:"plugins"   yaml:"plugins,omitempty"   env:"KRAFTKIT_PATHS_PLUGINS"`
		Config    string `json:"config"    yaml:"config,omitempty"    env:"KRAFTKIT_PATHS_CONFIG"`
		Manifests string `json:"manifests" yaml:"manifests,omitempty" env:"KRAFTKIT_PATHS_MANIFESTS"`
		Sources   string `json:"sources"   yaml:"sources,omitempty"   env:"KRAFTKIT_PATHS_SOURCES"`
	} `json:"paths" yaml:"paths,omitempty"`

	Log struct {
		Level      string `json:"level"      yaml:"level"      env:"KRAFTKIT_LOG_LEVEL"`
		Timestamps bool   `json:"timestamps" yaml:"timestamps" env:"KRAFTKIT_LOG_TIMESTAMPS"`
		Type       string `json:"type"       yaml:"type"       env:"KRAFTKIT_LOG_TYPE"`
	} `json:"log" yaml:"log"`

	Unikraft struct {
		Mirrors   []string `json:"mirrors"   yaml:"mirrors"   env:"KRAFTKIT_UNIKRAFT_MIRRORS"`
		Manifests []string `json:"manifests" yaml:"manifests" env:"KRAFTKIT_UNIKRAFT_MANIFESTS"`
	} `json:"unikraft" yaml:"unikraft"`

	Auth map[string]AuthConfig `json:"auth" yaml:"auth,omitempty"`

	Aliases map[string]map[string]string `json:"aliases" yaml:"aliases"`
}

type ConfigDetail struct {
	Key           string
	Description   string
	AllowedValues []string
}

// Descriptions of each configuration parameter as well as valid values
var configDetails = []ConfigDetail{
	{
		Key:         "no_prompt",
		Description: "toggle interactive prompting in the terminal",
	},
	{
		Key:         "editor",
		Description: "the text editor program to use for authoring text",
	},
	{
		Key:         "browser",
		Description: "the web browser to use for opening URLs",
	},
	{
		Key:         "git_protocol",
		Description: "the protocol to use for git clone and push operations",
		AllowedValues: []string{
			"https",
			"ssh",
		},
	},
	{
		Key:         "pager",
		Description: "the terminal pager program to send standard output to",
	},
	{
		Key:         "log.level",
		Description: "Set the logging verbosity",
		AllowedValues: []string{
			"fatal",
			"error",
			"warn",
			"info",
			"debug",
			"trace",
		},
	},
	{
		Key:         "log.type",
		Description: "Set the logging verbosity",
		AllowedValues: []string{
			"quiet",
			"basic",
			"fancy",
			"json",
		},
	},
	{
		Key:         "log.timestamps",
		Description: "Show timestamps with log output",
	},
}

func ConfigDetails() []ConfigDetail {
	return configDetails
}
