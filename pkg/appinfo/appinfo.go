package appinfo

import (
	"fmt"
	"strings"
)

type Info struct {
	// Name is name of the application/service.
	Name string `toml:"name"`
	// Env is environment name of the application.
	// e.g. local|dev|uat|sit|prod
	Env string `toml:"env"`

	// Following values is set from above public var.

	// GitURL is URL of the git repository.
	GitURL string
	// GitCommitHash is commit hash for current build.
	GitCommitHash string
	// GitTag is the tagged version of the code.
	GitTag string
	// BuildOS OS which is used to build the binary.
	BuildOS string
	// BuildTime timestamp on build.
	BuildTime string
	// GoVersion is what version of Go this binary was built with.
	GoVersion string
}

func (i Info) NameWithEnv() string { return strings.ToLower(fmt.Sprintf("%s-%s", i.Env, i.Name)) }
