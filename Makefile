build-x64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build
	tar -czf cloud_ssh_x64.tar.gz cloud-ssh
	rm cloud-ssh

build-x86:
	GOOS=linux GOARCH=i386 CGO_ENABLED=0 go build
	tar -czf cloud_ssh_x86.tar.gz cloud-ssh
	rm cloud-ssh

build-macos:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build
	tar -czf cloud_ssh_macosx.tar.gz cloud-ssh
	rm cloud-ssh

all:
	build-macos
	build-x86
	build-x64