apt install -y go make

wget https://github.com/firecracker-microvm/firecracker/releases/download/v1.11.0/firecracker-v1.11.0-x86_64.tgz
tar -xzf firecracker-v1.11.0-x86_64.tgz
cp release-v1.11.0-x86_64/firecracker-v1.11.0-x86_64 /usr/bin/firecracker

rm -rf release-v1.11.0-x86_64 firecracker-v1.11.0-x86_64.tgz

ARCH="$(uname -m)"

latest=$(wget "http://spec.ccfc.min.s3.amazonaws.com/?prefix=firecracker-ci/v1.11/$ARCH/vmlinux-5.10&list-type=2" -O - 2>/dev/null | grep -oP "(?<=<Key>)(firecracker-ci/v1.11/$ARCH/vmlinux-5\.10\.[0-9]{1,3})(?=</Key>)")

# Download a linux kernel binary
wget "https://s3.amazonaws.com/spec.ccfc.min/${latest}"

go mod tidy
