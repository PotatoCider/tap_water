## `tap_water`: A fast TUN/TAP userspace driver for PL25A1-based cables

## Usage

```bash
# Build it
go build

# Check Usage
./tap_water -help

# Run it
sudo ./tap_water

# Cross-compile for Windows
make windows
```
Note that `root` is required for managing interfaces in Linux.

Administrator privileges are not needed on Windows, unless network interface configuration flags are passed (i.e. `-ip`, `-gw`) 

## Build Requirements

You would need to have `libusb` installed to build this repository.

### Windows Cross-Compilation
In order to cross compile, you would need to have `mingw-w64-gcc` from the Arch repositories

Because the libusb library is statically linked for Windows, you would need to get the libusb windows binaries from https://libusb.info/

Paste the contents of libusb-MinGW-x64 folder into /usr/x86_64-w64-mingw32 (MINGW_DIR in Makefile)

## Installation

### Linux
Works out of the box with Linux.

### Windows
You would need to install the [OpenVPN TAP drivers](https://openvpn.net/community-downloads/). The OpenVPN client is not needed.

### MacOS
You would need to install the [Tunnelblink](https://tunnelblick.net/downloads.html) tun/tap extensions.

```bash
brew install --cask tunnelblick
```

### Notes

Currently, this repo is only tested on Plugable Easy Transfer Cable. It should also work with other PL25A1-based cables. This repository can also be repurposed to allow other simple USB bulk buffer transfers. 
