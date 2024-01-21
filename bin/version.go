package bin

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/cobra"

	"github.com/usememos/memos/server/version"
)

type VersionInfo struct {
	Version       string `json:"version"`
	OSVersion     string `json:"os_version"`
	OSKernel      string `json:"os_kernel"`
	OSType        string `json:"os_type"`
	OSArch        string `json:"os_arch"`
	GoVersion     string `json:"go_version"`
	BinaryModTime string `json:"binary_mod_time"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Memos version information",
	Long: `
Show Memos version number, kernel, OS, architecture and Go version.
For example:
    $ memos version
    Memos v0.18.0
    - os/version: ubuntu 22.04 (64 bit)
    - os/kernel: 6.2.0-37-generic
    - os/type: linux
    - os/arch: amd64
    - go/version: go1.21.6
	- bin/modtime: 2024-01-21 15:04:05`,
	Run: func(cmd *cobra.Command, args []string) {
		versionInfo := GetVersionInfo()
		fmt.Printf("Memos v%s\n", versionInfo.Version)
		fmt.Printf("- os/version: %s\n", versionInfo.OSVersion)
		fmt.Printf("- os/kernel: %s\n", versionInfo.OSKernel)
		fmt.Printf("- os/type: %s\n", versionInfo.OSType)
		fmt.Printf("- os/arch: %s\n", versionInfo.OSArch)
		fmt.Printf("- go/version: %s\n", versionInfo.GoVersion)
		fmt.Printf("- bin/modtime: %s\n", versionInfo.BinaryModTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// GetVersionInfo returns version information.
func GetVersionInfo() VersionInfo {
	osVersion, osKernel := GetOSVersion()
	return VersionInfo{
		Version:       version.Version,
		OSVersion:     osVersion,
		OSKernel:      osKernel,
		OSType:        runtime.GOOS,
		OSArch:        runtime.GOARCH,
		GoVersion:     runtime.Version(),
		BinaryModTime: GetBinaryModTime(),
	}
}

// GetBinaryModTime returns main binary file modification time.
func GetBinaryModTime() string {
	binary, err := os.Executable()
	if err != nil {
		return ""
	}

	stat, err := os.Stat(binary)
	if err != nil {
		return ""
	}

	return stat.ModTime().Local().Format("2006-01-02 15:04:05")
}

// GetOSVersion returns OS version, kernel and bitness.
func GetOSVersion() (osVersion, osKernel string) {
	if platform, _, version, err := host.PlatformInformation(); err == nil && platform != "" {
		osVersion = platform
		if runtime.GOOS != "windows" && version != "" {
			osVersion += " " + version
		}
	}
	if osVersion == "" {
		osVersion = "unknown"
	}

	if version, err := host.KernelVersion(); err == nil {
		osKernel = version
	}
	if osKernel == "" {
		osKernel = "unknown"
	}

	if arch, err := host.KernelArch(); err == nil {
		if strings.HasSuffix(arch, "64") {
			osVersion += " (64 bit)"
		}
		osKernel += " (" + arch + ")"
	}

	return osVersion, osKernel
}
