# terraform http backend

[![Go Version](https://img.shields.io/github/go-mod/go-version/ironpinguin/terraform_http_backend)](https://img.shields.io/github/go-mod/go-version/ironpinguin/terraform_http_backend)
[![Coverage Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/ironpinguin/97d98d096e648370e2848116f7f8289a/raw/terraform_http_backend__main.json)](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/ironpinguin/97d98d096e648370e2848116f7f8289a/raw/terraform_http_backend__main.json)
[![run tests](https://github.com/ironpinguin/terraform_http_backend/actions/workflows/ci.yaml/badge.svg)](https://github.com/ironpinguin/terraform_http_backend/actions/workflows/ci.yaml)

This is a simple go lang implementation of the terraform http backend protocol including locking.
To store the information there is current only the filesystem used.

## Configuration (Environment)

The http server can be configured over environment variables set in the system or in the `.env` file.

Follow Environment are availibe:

| Variable | Description | Default |
|---------------|------------------------------------------------------------------------------------------|---------|
|`TF_STORAGE_DIR`| directory to store the uploaded terraform state file and the lock state | ./store |
|`TF_AUTH_ENABLED`| boolean to enable or disable basic auth security|false|
|`TF_USERNAME`| Username for the basic auth security only used if `TF_AUTH_ENABLED` is `true`|admin|
|`TF_PASSWORD`| Password  for the basic auth security only used if `TF_AUTH_ENABLED` is `true`|admin|
|`TF_PORT`| The Port where this server will listen |8080|
|`TF_IP`| The ip addr for the server to listen. If none is set the server will listen on all interfaces|127.0.0.1|

## Usage

Download latest release config your .env file or set the environment varibles.
Than you can simple start the server with `./terraform_http_backend`

## Upcomming

I the future there will be a docker image and also a example systemd start script