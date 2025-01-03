package cfg

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
)

//go:embed default
var defaultDir embed.FS

type BlurCfg struct {
    Theme string `toml:"theme"`
}

func home() (homeDir string, err error) {
    homeDir, ok := os.LookupEnv("HOME")
    if !ok {
        err = fmt.Errorf("Could not find $HOME environment variable")
    }
    return
}

func LoadConfig() (cfg BlurCfg, err error) {
    homeDir, err := home()
    if err != nil {
        return
    }
    path := fmt.Sprintf("%s/.config/blur/config.toml", homeDir)
    if _, err = toml.DecodeFile(path, &cfg); err != nil {
        return
    }
    return
}

// Only run this is the filesystem is not in place
func InitSystem() (err error) {
    sub, err := fs.Sub(defaultDir, "default")
    if err != nil {
        return
    }
    homeDir, err := home()
    if err != nil {
        return
    }
    cfgPath := fmt.Sprintf("%s/.config/blur", homeDir)
    err = os.CopyFS(cfgPath, sub)
    return
}
