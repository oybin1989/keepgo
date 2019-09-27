package main
import (
		"fmt"
		"sync"
		"os"
		"bufio"
		"regexp"
		"syscall"
)

var configFilePath string = "/etc/keepgo.conf"

var magicString string = "oWvXDHYgGT"

var once sync.Once

type ConfigEntry map[string]string

type Configurator struct {
	configurationEntries []ConfigEntry
	configurationNum int
	configurationRegex string
}

var globalConfig Configurator

var configurationRegex string = "^\\s*(\\D\\S*)\\s*\\\"(\\S.*)\\\"\\s*\\\"(\\S.*)\\\"\\s*\\\"(\\S.*)\\\""

func (configurator * Configurator) load(configFilePath string) {
	f, err := os.Open(configFilePath)
	if err != nil {
		fmt.Printf("Failed to read the configuration file %s\n", configFilePath)
		return
	}
	defer f.Close()

	configurator.configurationRegex = configurationRegex
	configurator.configurationNum = 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if match, _ := regexp.MatchString(configurator.configurationRegex, line); match {
			regex := regexp.MustCompile(configurator.configurationRegex)
			res := regex.FindStringSubmatch(line)
			attributes := make(ConfigEntry,0)
			attributes["name"] = res[1]
			attributes["regex"] = res[2]
			attributes["running"] = res[3]
			attributes["restart"] = res[4]
			configurator.configurationEntries = append(configurator.configurationEntries, attributes)
			configurator.configurationNum = configurator.configurationNum + 1
		}
	}
}

func getGonfigurator() *Configurator {
	once.Do(func() {
		globalConfig = Configurator{}
		globalConfig.configurationEntries = make([]ConfigEntry, 0)
		globalConfig.load(configFilePath)
	})
	return &globalConfig
}

func main()  {
	// configurator := getGonfigurator()
	// configurator = getGonfigurator()
	if len(os.Args) > 1 && os.Args[len(os.Args) - 1] == magicString {
		//configurator := getGonfigurator()
		syscall.Setsid()
		for true {
		}

		// foreground process
	} else {
		var sysproc = &syscall.SysProcAttr{ Noctty:true }
		attr := os.ProcAttr{
			Dir: ".",
			Env: os.Environ(),
			Files: []*os.File{
				os.Stdout,
				os.Stderr,
			},
			Sys: sysproc,
		}
		args := []string{os.Args[0], magicString}
		process, err := os.StartProcess(os.Args[0], args, &attr)
		os.Stdout = os.NewFile(0, os.DevNull)
		if err == nil {
			// It is not clear from docs, but Realease actually detaches the process
			err = process.Release();
			if err != nil {
					fmt.Println(err.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
}


