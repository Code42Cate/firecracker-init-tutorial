#! /bin/bash

echo "Building custom init process..."
go build -o my-custom-init init/main.go

echo "Creating and formatting filesystem image..."
truncate -s 400M rootfs.ext4
mkfs.ext4 rootfs.ext4

echo "Mounting filesystem image..."
mkdir -p /mnt
sudo mount rootfs.ext4 /mnt

echo "Downloading busybox..."
curl -LO "https://busybox.net/downloads/binaries/1.35.0-x86_64-linux-musl/busybox"

echo "Setting up busybox in chroot..."
sudo mkdir -p /mnt/bin
sudo cp busybox /mnt/bin/busybox
sudo chmod +x /mnt/bin/busybox

echo "Creating necessary directories in chroot..."
sudo mkdir -p /mnt/etc/services
sudo mkdir -p /mnt/var/log

echo "Linking ls service..."
sudo ln -s /bin/ls /mnt/etc/services/ls

echo "Installing busybox applets..."
sudo chroot /mnt /bin/busybox --install -s /bin

echo "Copying custom init process into chroot..."
sudo cp -r my-custom-init /mnt/my-custom-init

echo "Unmounting filesystem..."
sudo umount /mnt

echo "Filesystem setup complete."