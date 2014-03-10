package main

import (
	"fmt"
	"github.com/buger/cloud-ssh/provider"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type CloudInstances map[string]provider.Instances
type StrMap map[string]string
type Config map[string]StrMap

func splitHostname(str string) (user string, hostname string) {
	if arr := strings.Split(str, "@"); len(arr) > 1 {
		return arr[0], arr[1]
	} else {
		return "", str
	}
}

func joinHostname(user string, hostname string) string {
	if user != "" {
		return user + "@" + hostname
	} else {
		return hostname
	}
}

func getTargetHostname(args []string) (user string, hostname string, arg_idx int) {
	for idx, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			if idx == 0 {
				hostname = arg
				arg_idx = idx
				break
			} else {
				if !strings.HasPrefix(args[idx-1], "-") {
					hostname = arg
					arg_idx = idx
					break
				}
			}
		}
	}

	user, hostname = splitHostname(hostname)

	return
}

func getInstances(config Config) (clouds CloudInstances) {
	clouds = make(CloudInstances)

	for name, cfg := range config {
		for k, v := range cfg {
			cfg["name"] = name

			if k == "provider" {
				switch v {
				case "aws":
					clouds[name] = provider.GetEC2Instances(cfg)
				default:
					log.Println("Unknown provider: ", v)
				}
			}
		}
	}

	return
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func readConfig() (config Config) {
	config = make(Config)

	prefferedPaths := []string{
		"./cloud-ssh.yaml",
		userHomeDir() + "/.ssh/cloud-ssh.yaml",
		"/etc/cloud-ssh.yaml",
	}

	var content []byte

	for _, path := range prefferedPaths {
		if _, err := os.Stat(path); err == nil {
			fmt.Println("Found config:", path)
			content, err = ioutil.ReadFile(path)

			if err != nil {
				log.Fatal("Error while reading config: ", err)
			}
		}
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		config["default"] = make(StrMap)
		config["default"]["access_key"] = os.Getenv("AWS_ACCESS_KEY_ID")
		config["default"]["secret_key"] = os.Getenv("AWS_SECRET_ACCESS_KEY")
		config["default"]["region"] = os.Getenv("AWS_REGION")
		config["default"]["provider"] = "aws"
	}

	if len(content) == 0 {
		if len(config) == 0 {
			fmt.Println("Can't find any configuration or ENV variables. Check http://github.com/buger/cloud-ssh for documentation.")
		}
		return
	} else if err := yaml.Unmarshal(content, &config); err != nil {
		log.Fatal(err)
	}

	return
}

func getMatchedInstances(clouds CloudInstances, filter string) (matched []StrMap) {

	// Fuzzy matching (like SublimeText)
	filter = strings.Join(strings.Split(filter, ""), ".*?")

	rHost := regexp.MustCompile(filter)

	for cloud, instances := range clouds {
		for addr, tags := range instances {
			for _, tag := range tags {
				if rHost.MatchString(tag.Value) {
					matched = append(matched, StrMap{
						"cloud":     cloud,
						"addr":      addr,
						"tag_name":  tag.Name,
						"tag_value": tag.Value,
					})

					break
				}
			}
		}
	}

	return
}

func formatMatchedInstance(inst StrMap) string {
	return "Cloud: " + inst["cloud"] + "\tMatched by: " + inst["tag_name"] + "=" + inst["tag_value"] + "\tAddr: " + inst["addr"]
}

func main() {
	config := readConfig()
	instances := getInstances(config)

	args := os.Args[1:len(os.Args)]

	user, hostname, arg_idx := getTargetHostname(args)

	match := getMatchedInstances(instances, hostname)

	if len(match) == 0 {
		fmt.Println("Can't find cloud instance, trying to connect anyway")
	} else if len(match) == 1 {
		hostname = match[0]["addr"]
		fmt.Println("Found clound instance:")
		fmt.Println(formatMatchedInstance(match[0]))
	} else {
		fmt.Println("Found multiple instances:")
		for i, host := range match {
			fmt.Println(strconv.Itoa(i+1)+") ", formatMatchedInstance(host))
		}
		fmt.Print("Choose instance: ")

		var i int
		_, err := fmt.Scanf("%d", &i)

		if err != nil || i > len(match)+1 {
			log.Fatal("Wrong index")
		}

		hostname = match[i-1]["addr"]
	}

	args[arg_idx] = joinHostname(user, hostname)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}
