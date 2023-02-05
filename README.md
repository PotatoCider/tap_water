# `tap_water`: A fast TUN/TAP userspace driver for PL25A1-based cables

## Usage

```bash
# Build it
go build

# Check Usage
./tap_water -help

# Run it
sudo ./tap_water

```
Note that `root` is required for managing interfaces in Linux.

Administrator privileges are not needed on Windows, unless network interface configuration flags are passed (i.e. `-ip`, `-gw`) 

## Installation

### Linux
Works out of the box with Linux.

### Windows
You would need to install the OpenVPN TAP drivers from https://openvpn.net/community-downloads/. The OpenVPN client is not needed.

### Cross-Compilation
In order to cross compile, you would need to have `mingw-w64-gcc` from the Arch repositories

Because the libusb library is statically linked for Windows, you would need to get the libusb windows binaries from https://libusb.info/

Paste the contents of libusb-MinGW-x64 folder into /usr/x86_64-w64-mingw32 (MINGW_DIR in Makefile)


### Notes

Currently, this repo is only tested on Plugable Easy Transfer Cable. It should also work with other PL25A1-based cables. This repository can also be repurposed to allow other simple USB bulk buffer transfers. 