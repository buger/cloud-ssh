package main

import (
	"gopkg.in/amz.v2/aws"
	"gopkg.in/amz.v2/ec2"
	"log"
)

func getEC2Instances(config map[string]string) (instances Instances) {
	instances = make(Instances)

	if _, ok := config["access_key"]; !ok {
		log.Fatal("Missing access_key for ", config["name"], " AWS cloud")
	}

	if _, ok := config["secret_key"]; !ok {
		log.Fatal("Missing secret_key for ", config["name"], " AWS cloud")
	}

	if _, ok := config["region"]; !ok {
		config["region"] = "us-east-1"
	}

	if _, ok := config["output_format"]; !ok {
		config["output_format"] = "Cloud: {cloud} \tMatched by: {tag_name} = {tag_value} \tAddr: {addr}"
	}

	auth := aws.Auth{AccessKey: config["access_key"], SecretKey: config["secret_key"]}

	e := ec2.New(auth, aws.Regions[config["region"]])
	resp, err := e.Instances(nil, nil)

	if err != nil {
		log.Println(err)
		return
	}

	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {

			if inst.DNSName != "" {
				var tags []Tag

				for _, tag := range inst.Tags {
					tags = append(tags, Tag{tag.Key, tag.Value})
				}

				for _, sg := range inst.SecurityGroups {
					tags = append(tags, Tag{"Security group", sg.Name})
				}
				ci := config["connection_interface"]
				if ci == "private_ip" {
					instances[inst.PrivateIPAddress] = tags
				} else if ci == "public_ip" {
					instances[inst.IPAddress] = tags
				} else if ci == "private_dns" {
					instances[inst.PrivateDNSName] = tags
				} else {
					instances[inst.DNSName] = tags
				}
			}
		}
	}

	return
}
