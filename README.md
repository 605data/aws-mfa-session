# aws-mfa-session

## Introduction

For security reasons, a lot of companies mandate their AWS users to use the MFA for their AWS API accesses. aws-mfa-session is a command line tool to get your temporary AWS session credentials with your MFA device/token, and export the credentials into your new shell session. Hopefully, this command line tool can help you to mitigate the pain level which is brought in by your company seurity policies.

## Prerequisite
It requires your AWS account, and AWS credentials/profiles file to be setup on your computer.

Use below link for the details

https://boto3.amazonaws.com/v1/documentation/api/latest/guide/configuration.html

## Binary downloads

* macOS \
  [64-bit](https://drive.google.com/uc?export=download&id=16RNi05XNIqE47FfuMUqQ404epZPawIqV)
* Linux \
  [64-bit](https://drive.google.com/uc?export=download&id=1h1QfVWkgJry7lJO_QnaTTxh2oAgsl7Ne)
* Windows \
  [Under implementation]

## Installation
1. Use the above links to download correct version of aws-mfa-session binary for your OS.
2. Uppack downloaded arhcive

   For macOS|Linux
   > $ tar -xvf aws-mfa-session.tgz \
   > $ cd aws-mfa-session \
   > $ chown +x aws-mfa-session

    Directory layout

       aws-mfa-session/
        ├── aws-mfa-session
        └── aws_mfa.yml.sample
 
## How to run it

1. Copy aws_mfa.yml.sample to aws_mfa.yml.
2. Put your aws MFA virtual device's ids into aws_mfa.yml. 
3. Run "./aws-mfa-session help" to print usage and help messages

## Examples

1. If you are going to use your default AWS profile, you just need to run "./aws-mfa-session" in your shell/terminal.

2. If you want to specify a custom AWS profile, you can run like this "AWS_PROFILE=605-management ./aws-mfa-session". 

3. To get your AWS session token expiration time, run below in your aws-mfa-session shell session

> echo $AWS_SESSION_EXPIRATION 


Have fun and enjoy !!!


-----
Note: No PRs will be accepted at this moment.
