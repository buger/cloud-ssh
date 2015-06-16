package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Tag struct {
	Name, Value string
}

type Instances map[string][]Tag
type CloudInstances map[string]Instances

func getInstances(config Config) (clouds CloudInstances) {
	clouds = make(CloudInstances)

	var wg sync.WaitGroup
	var mux sync.RWMutex

	for name, cfg := range config {
		for k, v := range cfg {
			if k == "provider" {
				switch v {
				case "aws":
					wg.Add(1)
					go func(name string, cfg StrMap) {
						mux.Lock()
						clouds[name] = getEC2Instances(cfg)
						mux.Unlock()
						wg.Done()
					}(name, cfg)
				case "digital_ocean":
					wg.Add(1)
					go func(name string, cfg StrMap) {
						mux.Lock()
						clouds[name] = getDigitalOceanInstances(cfg)
						mux.Unlock()
						wg.Done()
					}(name, cfg)
				default:
					log.Println("Unknown provider: ", v)
				}
			}
		}
	}

	wg.Wait()

	return
}

type SortByTagValue []StrMap

func (a SortByTagValue) Len() int           { return len(a) }
func (a SortByTagValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByTagValue) Less(i, j int) bool { return a[i]["tag_value"] < a[j]["tag_value"] }

func getMatchedInstances(clouds CloudInstances, filter string) (matched []StrMap) {
	// Fuzzy matching, like SublimeText
	filter = strings.Join(strings.Split(filter, ""), ".*?")

	rHost := regexp.MustCompile(filter)

	for cloud, instances := range clouds {
		for addr, tags := range instances {
			for _, tag := range tags {
				if rHost.MatchString(cloud + tag.Value) {
					matched = append(matched, StrMap{
						"cloud":     cloud,
						"addr":      addr,
						"tag_name":  tag.Name,
						"tag_value": tag.Value,
						"instance_name": getInstanceName(tags),
					})

					break
				}
			}
		}
	}

	sort.Sort(SortByTagValue(matched))

	return
}

func formatMatchedInstance(inst StrMap, output string) string {
	c := strings.SplitAfter(output, "{")
	for i := 1; i < len(c); i++ {
		s := strings.SplitN(c[i], "}", 2)
		c[i] = getStringValue(inst, s[0])
		output = strings.Replace(output, "{" + s[0] + "}", c[i], -1)
	}
	return output
}

func getStringValue(inst StrMap, s string) string{
	if len(inst[s]) > 0 {
		return inst[s]
	}
	return "{" + s + "}"
}

func getInstanceName(tags []Tag) string {
	for _, tag := range tags {
		if tag.Name == "Name" { return tag.Value }
	}
	return "" 
}

func main() {
	config := readConfig()
	instances := getInstances(config)

	args := os.Args[1:len(os.Args)]

	user, hostname, arg_idx := getTargetHostname(args)

	match := getMatchedInstances(instances, hostname)

	var matched_instance map[string]string

	if len(match) == 0 {
		fmt.Println("Can't find cloud instance, trying to connect anyway")
	} else if len(match) == 1 {
		matched_instance = match[0]
	} else {
		for i, host := range match {
			fmt.Println(strconv.Itoa(i+1)+") ", formatMatchedInstance(host, config[host["cloud"]]["output_format"]))
		}
		fmt.Print("Choose instance: ")

		var i int
		_, err := fmt.Scanf("%d", &i)

		if err != nil || i > len(match)+1 {
			log.Fatal("Wrong index")
		}

		matched_instance = match[i-1]
	}

	if matched_instance != nil {
		hostname = matched_instance["addr"]
		default_user := config[matched_instance["cloud"]]["default_user"]

		if len(user) == 0 && len(default_user) > 0 {
			user = default_user
		}

		fmt.Println("Connecting to instance:")
		fmt.Println(formatMatchedInstance(matched_instance, config[matched_instance["cloud"]]["output_format"]))
	}

	if len(args) == 0 {
		args = append(args, joinHostname(user, hostname))
	} else {
		args[arg_idx] = joinHostname(user, hostname)
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()
}
