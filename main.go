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
	GoPath  string
	GoProxy string
)

type BinInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Mod     string `json:"mod"`
	Path    string `json:"path"`
}

type Command struct {
	Name       string
	FlagSet    *flag.FlagSet
	HandleFunc func() error
}

func main() {
	commands := map[string]*Command{
		"list":    setupListCommand(),
		"upgrade": setupUpgradeCommand(),
		"install": setupInstallCommand(),
	}

	printDefaults := func() {
		fmt.Printf("Usage: gbvm <command> [options]\n\n")
		fmt.Printf("A command line tool to manage Go binaries\n\n")
		for _, cmd := range commands {
			fmt.Printf("%s commands:\n", cmd.Name)
			cmd.FlagSet.PrintDefaults()
			fmt.Println("")
		}
	}

	if len(os.Args) < 2 {
		printDefaults()
		return
	}

	cmd, exists := commands[os.Args[1]]
	if !exists {
		printDefaults()
		return
	}

	if err := cmd.FlagSet.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return
	}

	if err := cmd.HandleFunc(); err != nil {
		fmt.Println(err)
	}
}

func init() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = "~/go"
	}
	goProxy := os.Getenv("GOPROXY")
	if goProxy == "" {
		if strings.Contains(goPath, ",") {
			proxies := strings.Split(goPath, ",")
			if len(proxies) > 0 {
				goProxy = proxies[0]
			}
		}
		if goProxy == "" {
			goProxy = "https://proxy.golang.org"
		}
	}
	GoPath = goPath
	GoProxy = goProxy
}

func setupListCommand() *Command {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	showVersion := fs.Bool("version", true, "show version")
	jsonMode := fs.Bool("json", false, "json mode")
	return NewCommand(fs, func() error {
		return handleList(*jsonMode, *showVersion)
	})
}

func setupUpgradeCommand() *Command {
	fs := flag.NewFlagSet("upgrade", flag.ExitOnError)
	all := fs.Bool("all", false, "upgrade all")
	bin := fs.String("bin", "", "bin name")
	skipDev := fs.Bool("skip-dev", false, "skip dev version")
	return NewCommand(fs, func() error {
		if *all {
			return upgradeAllBins(*skipDev)
		}
		return upgradeBin(*bin)
	})
}

func setupInstallCommand() *Command {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	backupPath := fs.String("backup", "", "backup json path")
	return NewCommand(fs, func() error {
		return handleInstall(*backupPath)
	})
}

func NewCommand(fs *flag.FlagSet, handleFunc func() error) *Command {
	cmd := &Command{
		Name:    fs.Name(),
		FlagSet: fs,
	}
	help := cmd.FlagSet.Bool("help", false, "show help")
	cmd.HandleFunc = func() error {
		if *help {
			fs.Usage()
			return nil
		}
		return handleFunc()
	}
	return cmd
}

func handleList(jsonMode, showVersion bool) error {
	versions, err := loadAllBinVersions()
	if err != nil {
		return err
	}
	if jsonMode {
		encoded, e := json.MarshalIndent(versions, "", "  ")
		if e != nil {
			return e
		}
		fmt.Println(string(encoded))
		return nil
	}
	for _, v := range versions {
		if showVersion {
			fmt.Printf("%s\t%s\n", v.Name, v.Version)
		} else {
			fmt.Println(v.Name)
		}
	}
	return nil
}

func handleInstall(backupPath string) error {
	file, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}
	var versions []*BinInfo
	err = json.Unmarshal(file, &versions)
	if err != nil {
		return err
	}
	for _, v := range versions {
		info, e := loadBinInfo(v.Path)
		if e != nil {
			fmt.Printf("failed to load %s: %v\n", v.Name, e)
			continue
		}
		if info.Version == v.Version {
			fmt.Printf("skip %s\n", v.Name)
			continue
		}
		e = installBinByVersion(v.Path, v.Version)
		if e != nil {
			fmt.Printf("failed to install %s: %v\n", v.Name, e)
		}
	}
	return nil
}

func loadBinInfo(binPath string) (*BinInfo, error) {
	file, err := os.Open(binPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := buildinfo.Read(file)
	if err != nil {
		return nil, err
	}
	// filename without path
	fineName := filepath.Base(binPath)
	return &BinInfo{
		Name:    fineName,
		Version: info.Main.Version,
		Mod:     info.Main.Path,
		Path:    info.Path,
	}, err
}

func loadAllBinVersions() ([]*BinInfo, error) {
	binPath := filepath.Join(GoPath, "bin")
	files, err := os.ReadDir(binPath)
	if err != nil {
		return nil, err
	}
	var result []*BinInfo
	for _, file := range files {
		fullPath := filepath.Join(binPath, file.Name())
		info, e := loadBinInfo(fullPath)
		if e != nil {
			continue
		}
		result = append(result, info)
	}
	return result, nil
}

func upgradeAllBins(skipDev bool) error {
	versions, err := loadAllBinVersions()
	if err != nil {
		return err
	}
	for _, v := range versions {
		if skipDev && v.Version == "devel" {
			continue
		}
		if e := tryUpgradeBin(v); e != nil {
			fmt.Printf("failed to upgrade %s: %v\n", v.Name, e)
		}
	}
	return nil
}

func upgradeBin(binName string) error {
	binPath := filepath.Join(GoPath, "bin", binName)
	info, err := loadBinInfo(binPath)
	if err != nil {
		return err
	}
	return tryUpgradeBin(info)
}

func tryUpgradeBin(bin *BinInfo) error {
	latestVersion, err := fetchLatestVersion(bin.Mod)
	if err != nil {
		return fmt.Errorf("failed to fetch latest version: %v", err)
	}
	if latestVersion != bin.Version {
		fmt.Printf("upgrading %s from %s to %s\n", bin.Name, bin.Version, latestVersion)
		return installBinByVersion(bin.Path, latestVersion)
	}
	return nil
}

func fetchLatestVersion(modName string) (string, error) {
	url := fmt.Sprintf("%s/%s/@latest", GoProxy, strings.ToLower(modName))
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

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

func installBinByVersion(cmdPath, version string) error {
	uri := fmt.Sprintf("%s@%s", cmdPath, version)
	return exec.Command("go", "install", uri).Run()
}
