// Package buildinfo provides a collection of helpers to make introspecting a
// build as a human or a machine easier.
//
// To use this library, you'll need to update your CI builds to include the
// relevant information. For example, if you have a shell script building your
// project, it could be updated like such:
//
//    BUILD_TIME=$(date +%s)
//    BUILD_HASH=$(git rev-parse --short HEAD)
//    RELEASE_VERSION=${RELEASE_VERSION:-dev}
//    BUILD_URL=${BUILD_URL:-}
//
//    LDFLAGS=""
//    LDFLAGS="$LDFLAGS -X github.com/daaku/buildinfo.buildTimeUnix=$BUILD_TIME"
//    LDFLAGS="$LDFLAGS -X github.com/daaku/buildinfo.buildHash=$BUILD_HASH"
//    LDFLAGS="$LDFLAGS -X github.com/daaku/buildinfo.releaseVersion=$RELEASE_VERSION"
//    LDFLAGS="$LDFLAGS -X github.com/daaku/buildinfo.buildURL=$BUILD_URL"
//
//    go build \
//      -trimpath \
//      -ldflags "$LDFLAGS" \
//      -o myapp \
//      github.com/me/myapp
package buildinfo

import (
	"bytes"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"text/tabwriter"
	"time"
)

var (
	startupTime = time.Now()

	buildTimeUnix  = "0"
	buildHash      = "dev"
	buildURL       = ""
	releaseVersion = "dev"

	buildTime time.Time

	buildInfo  []byte
	moduleInfo string
)

func init() {
	buildTimeUnixI, err := strconv.ParseInt(buildTimeUnix, 0, 0)
	if err != nil {
		panic(err)
	}

	buildTime = time.Unix(buildTimeUnixI, 0)

	info := bytes.Buffer{}
	fmt.Fprintf(&info, "Build Hash:\t%s\n", buildHash)
	fmt.Fprintf(&info, "Release Version:\t%s\n", releaseVersion)
	fmt.Fprintf(&info, "Go Version:\t%s\n", runtime.Version())
	if buildURL != "" {
		fmt.Fprintf(&info, "Build URL:\t%s\n", buildURL)
	}
	buildInfo = info.Bytes()

	if bi, ok := debug.ReadBuildInfo(); ok {
		info := bytes.Buffer{}
		fmt.Fprint(&info, "Modules:\n")
		tw := tabwriter.NewWriter(&info, 0, 0, 1, ' ', 0)
		for _, m := range bi.Deps {
			fmt.Fprintf(tw, "%s\t%s\n", m.Path, m.Version)
		}
		tw.Flush()
		moduleInfo = info.String()
	}
}

// ReleaseVersion returns the release version of this built binary. It may
// return "dev" if a build version isn't avaiable.
func ReleaseVersion() string {
	return releaseVersion
}

// BuildHash returns the release hash of this built binary. It may
// return "dev" if a build hash isn't avaiable.
func BuildHash() string {
	return buildHash
}

// BuildTime returns the time at which this binary was built.
func BuildTime() time.Time {
	return buildTime
}

// BuildURL returns the URL for the CI build. It may be blank.
func BuildURL() string {
	return buildURL
}

// StartupTime returns the time at which this binary was executed.
func StartupTime() time.Time {
	return startupTime
}

// BasicInfo returns a pretty-print version of various useful pieces of build
// information.
func BasicInfo() []byte {
	var b bytes.Buffer
	tw := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	if buildTimeUnix != "0" {
		fmt.Fprintf(tw, "Build Time:\t%v (%v ago)\n", buildTime,
			time.Since(buildTime).Truncate(time.Second))
	}
	uptime := time.Since(startupTime).Truncate(time.Second)
	if uptime != 0 {
		fmt.Fprintf(tw, "Server Uptime:\t%v\n", uptime)
	}
	_, _ = tw.Write(buildInfo)
	_ = tw.Flush()
	return b.Bytes()
}

// ModuleInfo provides a pretty table with the modules and corresponding
// versions.
func ModuleInfo() string {
	return moduleInfo
}

// FullInfo provide a combined pretty printed information containing build info
// as well as module info.
func FullInfo() []byte {
	bi := BasicInfo()
	bi = append(bi, '\n')
	bi = append(bi, moduleInfo...)
	return bi
}
