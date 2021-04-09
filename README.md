# openfaas-functions
Repository for OpenFaaS functions

## Commands Example
Create a new function:
```bash
faas-cli new faas-cli new <function_name> --lang <language> --prefix <docker_image_prefix>
```
Build a function:
```bash
faas-cli build -f <stack.yml>
```
Publish a function:
```bash
faas-cli publish -f <stack.yml>
```
Deploy a function:
```bash
faas-cli deploy -f <stack.yml>
```
Build, publish and deploy all in one command:
```bash
faas-cli up -f <stack.yml>
```
