package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/mwlng/aws-go-clients/clients"
	"github.com/mwlng/aws-go-clients/service"
)

const (
	version  = "1.0.0"
	revision = "01"

	awsCredsFile     = ".aws/credentials"
	awsDefaultRegion = "us-east-1"
	awsMfaConfigFile = "aws_mfa"
)

type awsMfa struct {
	Profiles []map[string]map[string]string `yaml:"profiles"`
}

var (
	cwd             string
	sess            *session.Session
	stsCli          *clients.STSClient
	awsProfile      string
	awsProfileMfa   *awsMfa
	awsRegion       string = awsDefaultRegion
	assumeRole      bool   = false
	assumeRoleArn   string
	roleSessionName string
)

func init() {
	var err error
	if strings.HasPrefix(os.Args[0], "./") {
		cwd, err = filepath.Abs(filepath.Dir(os.Args[0]))
	} else {
		cwd, err = exec.LookPath(os.Args[0])
		if err == nil {
			cwd = filepath.Dir(cwd)
		}
	}
	if err != nil {
		log.Fatal(err)
	}

	mfaConfigFile := fmt.Sprintf("%s/%s.yml", cwd, awsMfaConfigFile)
	awsProfileMfa = readYamlConfig(mfaConfigFile)

	awsProfile = os.Getenv("AWS_PROFILE")
	if len(strings.TrimSpace(awsProfile)) == 0 {
		awsProfile = "default"
	}
	homeDir := getUserHomeDir()
	awsCredsFilePath := fmt.Sprintf("%s/%s", homeDir, awsCredsFile)
	profile := readAwsProfile(awsProfile, awsCredsFilePath)
	if profile == nil {
		log.Fatalf(`
            Can't find aws profile %s in : "%s"
            Please make sure your AWS credentials is setup properly`,
			awsProfile, awsCredsFilePath)
	}
	if profile.HasKey("region") {
		awsRegion = profile.Key("region").String()
	}
	if profile.HasKey("role_session_name") {
		roleSessionName = profile.Key("role_session_name").String()
	} else {
		me, _ := user.Current()
		roleSessionName = me.Name
	}

	var svc = service.Service{}
	if profile.HasKey("role_arn") {
		if profile.HasKey("source_profile") {
			srcAwsProfile := profile.Key("source_profile").String()
			srcProfile := readAwsProfile(srcAwsProfile, awsCredsFilePath)
			if srcProfile == nil {
				log.Fatalf("Can't find source profile %s in %s", srcProfile, awsCredsFilePath)
			}
			svc := service.Service{
				AccessKey: srcProfile.Key("aws_access_key_id").String(),
				SecretKey: srcProfile.Key("aws_secret_access_key").String(),
			}
			sess = svc.NewSession()
			assumeRole = true
			assumeRoleArn = profile.Key("role_arn").String()
		} else {
			log.Fatalf(`
                Assume role profile: "%s" missing source profile
                Please make sure your AWS credentials is setup properly \n`,
				awsProfile)
		}
	} else {
		sess = svc.NewSession()
	}

	stsCli = clients.NewClient("sts", sess).(*clients.STSClient)
}

func main() {
	if len(os.Args) >= 2 {
		service := os.Args[1]
		if service == "help" {
			printUsage()
			os.Exit(0)
		}
	}

	printBanner(version, revision)
	defer fmt.Println("Byte!")

	var mfa string
	var timeout int64 = 3600
	for _, p := range awsProfileMfa.Profiles {
		for k, v := range p {
			if k == awsProfile {
				mfa, _ = v["mfa"]
				if sessTimeout, ok := v["session_timeout"]; ok {
					timeout, _ = strconv.ParseInt(sessTimeout, 10, 64)
				}
				break
			}
		}
	}

	mfaCode := ""
	if len(mfa) > 0 {
		fmt.Printf("Enter your MFA code: ")
		fmt.Scanln(&mfaCode)
	}

	var creds *sts.Credentials
	if assumeRole {
		if len(mfaCode) > 0 {
			creds = stsCli.AssumeRoleWithMfa(&assumeRoleArn, &timeout, &roleSessionName,
				&mfa, &mfaCode)
		} else {
			creds = stsCli.AssumeRoleWithoutMfa(&assumeRoleArn, &timeout, &roleSessionName)
		}
	} else {
		if len(mfaCode) > 0 {
			creds = stsCli.GetSessionCredsWithMfa(&mfa, &mfaCode, &timeout)
		} else {
			creds = stsCli.GetSessionCredsWithoutMfa(&timeout)
		}
	}

	credsMap := structToMap(*creds)
	var credsExportStr = fmt.Sprintf("export AWS_PROFILE=%s ", awsProfile)
	for k, v := range *credsMap {
		if k == "AccessKeyId" {
			credsExportStr += fmt.Sprintf("%s=%s ", "AWS_ACCESS_KEY_ID", v.(string))
			os.Setenv("AWS_ACCESS_KEY_ID", v.(string))
		} else if k == "SecretAccessKey" {
			credsExportStr += fmt.Sprintf("%s=%s ", "AWS_SECRET_ACCESS_KEY", v.(string))
			os.Setenv("AWS_SECRET_ACCESS_KEY", v.(string))
		} else if k == "SessionToken" {
			credsExportStr += fmt.Sprintf("%s=%s ", "AWS_SESSION_TOKEN", v.(string))
			os.Setenv("AWS_SESSION_TOKEN", v.(string))
		} else if k == "Expiration" {
			credsExportStr += fmt.Sprintf("%s=%s ", "AWS_SESSION_EXPIRATION", v.(string))
			os.Setenv("AWS_SESSION_EXPIRATION", v.(string))
		}
	}
	fmt.Printf("The new session token will be expired %s \n", (*credsMap)["Expiration"])
	fmt.Println("")
	fmt.Println("Below AWS session credentials will be exported into your new shell session:")
	fmt.Println(credsExportStr)

	homeDir := getUserHomeDir()
	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   homeDir,
	}

	var err error
	var proc *os.Process
	argsLen := len(os.Args)
	if argsLen > 1 {
		if argsLen == 2 {
			proc, err = os.StartProcess(os.Args[1], []string{}, &pa)
		} else {
			args := os.Args[1:]
			proc, err = os.StartProcess(os.Args[1], args, &pa)
		}
	} else {
		proc, err = os.StartProcess("/bin/bash", []string{}, &pa)
	}
	if err != nil {
		panic(err)
	}

	if _, err = proc.Wait(); err != nil {
		panic(err)
	}
}
