package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

func getUserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func readAwsProfile(awsProfileName string, awsCredsFilePath string) *ini.Section {
	iniFile, err := ini.Load(awsCredsFilePath)
	if err != nil {
		log.Fatalf("Fail to read AWS credentials file: %v \n", err)
	}
	if section, err := iniFile.GetSection(awsProfileName); err == nil {
		return section
	}

	return nil
}

func readYamlConfig(configFile string) *awsMfa {
	var output awsMfa
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Read AWS mfa config file err: %v \n", err)
	}
	err = yaml.Unmarshal(data, &output)
	if err != nil {
		log.Fatalf("Unmarshal AWS mfa config err: %v \n", err)
	}
	return &output
}

func structToMap(input interface{}) *map[string]interface{} {
	var out map[string]interface{}
	in, _ := json.Marshal(input)
	json.Unmarshal(in, &out)

	return &out
}

func printBanner(ver string, rev string) {
	fmt.Printf("aws-temporary-creds %s (rev-%s)\n", ver, rev)
}

func printUsage() {
	fmt.Println(
		`
aws-temporary-creds is a command line to get your temporary AWS session credentials,
and export the temporary credentials into your new shell session.

Usage:
    AWS_PROFILE=<Your AWS profile> ./aws-temporary-creds

Examples:
    AWS_PROFILE=605-dev ./aws-temporary-creds

    or ./aws-temporary-session using your default AWS profile

`)
}
