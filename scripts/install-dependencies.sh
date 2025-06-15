#!/bin/bash

sudo apt update
sudo apt install -y libvirt-daemon-system libvirt-clients virtinst qemu-kvm genisoimage

sudo usermod -aG libvirt $(whoami)

virsh list --all