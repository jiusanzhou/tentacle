package client

import (
	"fmt"
	"github.com/jiusanzhou/tentacle/log"
	"github.com/jiusanzhou/tentacle/version"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	//"os"
	//"os/user"
	//"path"
	"github.com/jiusanzhou/tentacle/util"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	cachePath = ".cache"
)

type Configuration struct {
	ServerAddr string `yaml:"server_addr,omitempty"`
	AuthToken  string `yaml:"auth_token,omitempty"`
	PoolSize   int    `yaml:"-"`
	LogTo      string `yaml:"-"`
	Path       string `yaml:"-"`
	DialInfo   string `yaml:"dial_info,omitempty"`
}

func LoadConfiguration(opts *Options) (config *Configuration, err error) {

	configPath := opts.config
	if configPath == "" {
		configPath = defaultPath()
	}

	log.Info("Reading configuration file %s", configPath)
	configBuf, err := ioutil.ReadFile(configPath)
	if err != nil {

		// failure to read a configuration file is only a fatal error if
		// the user specified one explicitly
		if opts.config != "" {
			err = fmt.Errorf("Failed to read configuration file %s: %v", configPath, err)
			if os.IsNotExist(err) {
				f, e := os.Create(opts.config)
				if e == nil {
					f.Close()
				}
			}

			return
		}
	}

	// deserialize/parse the config
	config = new(Configuration)
	if err = yaml.Unmarshal(configBuf, &config); err != nil {
		err = fmt.Errorf("Error parsing configuration file %s: %v", configPath, err)
		return
	}

	// try to parse the old format for backwards compatibility
	matched := false
	content := strings.TrimSpace(string(configBuf))
	if matched, err = regexp.MatchString("^[0-9a-zA-Z_\\-!]+$", content); err != nil {
		return
	} else if matched {
		config = &Configuration{AuthToken: content}
	}

	// set configuration defaults
	if config.ServerAddr == "" {
		config.ServerAddr = defaultServerAddr
	}

	if config.ServerAddr, err = normalizeAddress(config.ServerAddr, "server_addr"); err != nil {
		return
	}

	// override configuration with command-line options
	config.LogTo = opts.logto
	config.Path = configPath

	if config.PoolSize < 0 {
		config.PoolSize = 0
	}

	if opts.authtoken != "" {
		config.AuthToken = opts.authtoken
	}

	switch opts.command {
	// start tunnels
	case "start":
	// case "info":
	case "redial":
		switch runtime.GOOS {
		case "windows":
			if len(opts.args) == 0 {
				opts.args = append(opts.args, strings.Split(config.DialInfo, " ")...)
			}

			if len(opts.args) != 3 {
				fmt.Println("Redial connection to Internet, you must offer `addressname account password`")
				os.Exit(1)
			}
			o, err := util.DoCommand(fmt.Sprintf("rasdial %s /DISCONNECT", opts.args[0]))
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(util.B2s(o))
			}
			time.Sleep(3 * time.Second)
			o, err = util.DoCommand(fmt.Sprintf("rasdial %s %s %s", opts.args[0], opts.args[1], opts.args[2]))
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(util.B2s(o))
			}
		default:
			fmt.Println("Don't support any other OS except windows yet.")
		}
		os.Exit(0)

	default:
		err = fmt.Errorf("Unknown command: %s", opts.command)
		return
	}

	return
}

func defaultPath() string {
	//user, err := user.Current()
	//
	//// user.Current() does not work on linux when cross compiling because
	//// it requires CGO; use os.Getenv("HOME") hack until we compile natively
	//homeDir := os.Getenv("HOME")
	//if err != nil {
	//	log.Warn("Failed to get user's home directory: %s. Using $HOME: %s", err.Error(), homeDir)
	//} else {
	//	homeDir = user.HomeDir
	//}
	//
	//return path.Join(homeDir, "."+version.Name)
	return version.Name + ".yaml"
}

func normalizeAddress(addr string, propName string) (string, error) {
	// normalize port to address
	if _, err := strconv.Atoi(addr); err == nil {
		addr = ":" + addr
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("Invalid address %s '%s': %s", propName, addr, err.Error())
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return fmt.Sprintf("%s:%s", host, port), nil
}

func validateProtocol(proto, propName string) (err error) {
	switch proto {
	case "http", "https", "http+https", "tcp":
	default:
		err = fmt.Errorf("Invalid protocol for %s: %s", propName, proto)
	}

	return
}

func SaveAuthToken(configPath, authtoken string) (err error) {
	// empty configuration by default for the case that we can't read it
	c := new(Configuration)

	// read the configuration
	oldConfigBytes, err := ioutil.ReadFile(configPath)

	if err == nil {
		// unmarshal if we successfully read the configuration file
		if err = yaml.Unmarshal(oldConfigBytes, c); err != nil {
			return
		}
	}

	// no need to save, the authtoken is already the correct value
	if c.AuthToken == authtoken {
		return
	}

	// update auth token
	c.AuthToken = authtoken

	// rewrite configuration
	newConfigBytes, err := yaml.Marshal(c)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(configPath, newConfigBytes, 0600)
	return
}

func GetCachedId() string {
	b, _ := ioutil.ReadFile(cachePath)
	return util.B2s(b)
}

func SaveCacheId(id string) {
	ioutil.WriteFile(cachePath, util.S2b(id), 0600)
}
