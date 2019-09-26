package main
import (
		"fmt"
		"sync"
		"os"
		"bufio"
		"regexp"
)

var configFilePath string = "/etc/keepgo.conf"

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
			fmt.Println(res[1])
			attributes["regex"] = res[2]
			fmt.Println(res[2])
			attributes["running"] = res[3]
			fmt.Println(res[3])
			attributes["restart"] = res[4]
			fmt.Println(res[4])
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
	configurator := getGonfigurator()
	configurator = getGonfigurator()
	fmt.Println(*configurator)
}


