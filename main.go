package main

import (
	"debug/buildinfo"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	GoPath  = os.Getenv("GOPATH")
	GoProxy = os.Getenv("GOPROXY")
)

func init() {
	if GoPath == "" {
		GoPath = "~/go"
	}
	if GoProxy == "" {
		if strings.Contains(GoPath, ",") {
			proxies := strings.Split(GoPath, ",")
			if len(proxies) > 0 {
				GoProxy = proxies[0]
			}
		}
		if GoProxy == "" {
			GoProxy = "https://proxy.golang.org"
		}
	}
}

type BinInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Mod     string `json:"mod"`
	Path    string `json:"path"`
}

func main() {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listShowVersion := listCmd.Bool("version", true, "show version")
	listJsonMode := listCmd.Bool("json", false, "json mode")

	upgradeCmd := flag.NewFlagSet("upgrade", flag.ExitOnError)
	upgradeAll := upgradeCmd.Bool("all", false, "upgrade all")
	binName := upgradeCmd.String("bin", "", "bin name")
	skipDev := upgradeCmd.Bool("skip-dev", false, "skip dev version")

	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	backupJsonPath := installCmd.String("backup", "", "backup json path")

	if len(os.Args) < 2 {
		fmt.Println("usage: go run main.go [list|upgrade]")
		return
	}
	switch os.Args[1] {
	case "list":
		err := listCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			return
		}
		err = runListSubCommand(err, *listJsonMode, *listShowVersion)
		if err != nil {
			fmt.Println(err)
			return
		}
	case "upgrade":
		err := upgradeCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			return
		}
		if *upgradeAll {
			err = upgradeAllBinVersion(*skipDev)
		} else {
			err = upgradeBinVersion(*binName)
		}
		if err != nil {
			fmt.Println(err)
			return
		}
	case "install":
		err := installCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			return
		}
		err = runInstallBackupCommand(*backupJsonPath)
		if err != nil {
			fmt.Println(err)
			return
		}
	default:
		fmt.Println("usage: go run main.go [list|upgrade]")
	}
}

func runListSubCommand(err error, listJsonMode bool, listShowVersion bool) error {
	version, err := loadAllBinVersion()
	if err != nil {
		return err
	}
	if listJsonMode {
		data, e := json.MarshalIndent(version, "", "  ")
		if e != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	for _, v := range version {
		if listShowVersion {
			fmt.Printf("%s\t%s\n", v.Name, v.Version)
		} else {
			fmt.Println(v.Name)
		}
	}
	return nil
}

func runInstallBackupCommand(backupJsonPath string) error {
	file, err := os.ReadFile(backupJsonPath)
	if err != nil {
		return err
	}
	var version []BinInfo
	err = json.Unmarshal(file, &version)
	if err != nil {
		return err
	}
	for _, v := range version {
		if info, e := loadBinInfo(v.Path); e == nil && info.Main.Version == v.Version {
			fmt.Printf("skip %s\n", v.Name)
		}
		e := installBinByVersion(v.Path, v.Version)
		if e != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func loadBinInfo(binPath string) (*buildinfo.BuildInfo, error) {
	file, err := os.Open(binPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return buildinfo.Read(file)
}

func loadAllBinVersion() ([]BinInfo, error) {
	binPaths := filepath.Join(GoPath, "bin")
	files, err := os.ReadDir(binPaths)
	if err != nil {
		return nil, err
	}
	var result []BinInfo
	for _, file := range files {
		binPath := filepath.Join(binPaths, file.Name())
		bi, e := loadBinInfo(binPath)
		if e != nil {
			continue
		}
		result = append(result, BinInfo{
			Name:    file.Name(),
			Version: bi.Main.Version,
			Mod:     bi.Main.Path,
			Path:    bi.Path,
		})
	}
	return result, nil
}

func upgradeAllBinVersion(skipDev bool) error {
	version, err := loadAllBinVersion()
	if err != nil {
		return err
	}
	for _, v := range version {
		if skipDev && v.Version == "devel" {
			continue
		}
		latestVersion, e := fetchLatestVersion(v.Mod)
		if e != nil {
			fmt.Printf("get latest version of %s failed: %s\n", v.Name, e)
			continue
		}
		if latestVersion != v.Version {
			fmt.Printf("upgrading %s from %s to %s\n", v.Name, v.Version, latestVersion)
			err = installBinByVersion(v.Path, latestVersion)
			if err != nil {
				fmt.Printf("upgrade %s failed: %s\n", v.Name, err)
			}
		}
	}
	return nil
}

func upgradeBinVersion(binName string) error {
	binPath := filepath.Join(GoPath, "bin", binName)
	bi, err := loadBinInfo(binPath)
	if err != nil {
		return err
	}
	latestVersion, err := fetchLatestVersion(bi.Main.Path)
	if err != nil {
		return err
	}
	if latestVersion != bi.Main.Version {
		fmt.Printf("upgrading %s from %s to %s\n", bi.Main.Path, bi.Main.Version, latestVersion)
		err = installBinByVersion(bi.Path, latestVersion)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func fetchLatestVersion(modName string) (string, error) {
	url := fmt.Sprintf("%s/%s/@latest", GoProxy, strings.ToLower(modName))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var versionInfo struct {
		Version string `json:"Version"`
	}
	err = json.Unmarshal(body, &versionInfo)
	if err != nil {
		return "", err
	}
	return versionInfo.Version, nil
}

func installBinByVersion(cmdPath string, version string) error {
	uri := fmt.Sprintf("%s@%s", cmdPath, version)
	cmd := exec.Command("go", "install", uri)
	return cmd.Run()
}
