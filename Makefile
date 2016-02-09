build:
	        GO15VENDOREXPERIMENT=1 go build

build_freebsd:
	        GO15VENDOREXPERIMENT=1 env GOOS=freebsd go build -o filehasher_freebsd

dependencies_init:
	        GO15VENDOREXPERIMENT=1 glide init

dependencies_update:
	        GO15VENDOREXPERIMENT=1 glide up
