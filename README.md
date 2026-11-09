# KernelScope

An eBPF learning project for exploring Linux kernel observability and tracing.

The current setup provides a Rocky Linux 10.2 development VM through Vagrant
and libvirt.

## Requirements

- Linux host with KVM/libvirt configured
- Vagrant with the `vagrant-libvirt` plugin
- rsync

## Quick start

```bash
vagrant plugin install vagrant-libvirt
vagrant up --provider=libvirt
vagrant ssh
```

The VM uses 2 CPUs and 2 GiB of RAM. Project files sync to
`/home/vagrant/kernelscope` in the guest. Run `vagrant rsync` after local changes,
or `vagrant rsync-auto` to watch for changes.

Stop the VM with `vagrant halt`.

## Troubleshooting

If Vagrant waits for an IP address on a host using UFW, see the
[libvirt DHCP troubleshooting guide](docs/vagrant/trobleshoot/vagrant-libvirt-dhcp-ufw-rca.md).

Local VM state and SSH keys in `.vagrant/` are not committed.
