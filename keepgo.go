package main
import (
    "fmt"
    "sync"
    "os"
    "bufio"
    "regexp"
    "syscall"
    "time"
    "log"
		"log/syslog"
		"io/ioutil"
		"strconv"
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
	// daemon process
	if len(os.Args) > 1 && os.Args[len(os.Args) - 1] == magicString {
		syscall.Setsid()
		logwriter, e := syslog.New(syslog.LOG_NOTICE, "keepgo")
		if e == nil {
			log.SetOutput(logwriter)
		}
		log.Printf("keepgo starts at %s", time.Now().Format(time.UnixDate))

		configurator := getGonfigurator()
		for true {
			files, err := ioutil.ReadDir("/proc")
			if err != nil {
				log.Printf("%w", err)
			}
			for _, file := range files {
				if _, err := strconv.Atoi(file.Name()); err == nil {
					f, err := os.Open("/proc/" + file.Name() + "/cmdline")
					if err != nil {
						log.Printf("Failed to read the process cmd %w\n", err)
					}
					defer f.Close()
					commandLine, err := ioutil.ReadAll(f)
					if err != nil {
						log.Printf("Failed to read the process cmd %w\n", err)
					}
					commandLineStr := string(commandLine)
					if len(commandLineStr) > 0 {
						log.Print(commandLineStr)
					}
					f.Close()
				}
			}
			log.Printf("keepgo ticks at %s", time.Now().Format(time.UnixDate))
			log.Printf("%p", configurator)
			// procAttrProcess := new(os.ProcAttr)
			// procAttrProcess.Files = []*os.File{nil, nil, nil}
			// _, err = os.StartProcess("/usr/bin/less", []string{"usr/bin/less", "/etc/keepgo.conf"}, procAttrProcess)
			// if err != nil {
			// 	log.Printf("error %w", err)
			// }
			time.Sleep(3*time.Second)
		}


	// foreground process
	} else {
		var sysproc = &syscall.SysProcAttr{ Noctty:true }
		attr := os.ProcAttr{
			Dir: ".",
			Env: os.Environ(),
			Files: []*os.File{
				os.Stdin,// stdin
				os.Stdout,// stdout
				os.Stderr,// stderr
			},
			Sys: sysproc,
		}
		// no native fork support exists in go, use magic string to distinguish daemon process
		args := []string{os.Args[0], magicString}
		process, err := os.StartProcess(os.Args[0], args, &attr)
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


