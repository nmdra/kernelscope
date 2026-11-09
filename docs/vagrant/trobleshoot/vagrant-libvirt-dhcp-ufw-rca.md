# RCA: Vagrant waits for a libvirt IP address because UFW blocks DHCP

## Scope

This report covers the host firewall issue only. It does not cover the separate Rocky box disk-size issue.

## Symptom

`vagrant up --provider=libvirt` stopped progressing at:

```text
==> default: Waiting for domain to get an IP address...
```

The Rocky Linux guest had booted to its login prompt, but Vagrant could not discover an IPv4 address or connect over SSH.

## Environment

- Host firewall: UFW, active
- Vagrant provider: libvirt
- Libvirt network: `vagrant-libvirt`
- Host bridge: `virbr1`
- Guest subnet: `192.168.121.0/24`
- Host bridge address: `192.168.121.1`
- Host outbound interface: `wlan0`
- Guest interface: `ens5`

Interface names and subnet values are specific to this host. Check them before applying these commands elsewhere.

## Evidence

Before the fix:

- The guest reached the Rocky Linux login prompt.
- The guest interface `ens5` was up but had no IPv4 address.
- NetworkManager repeatedly attempted DHCP, then reported `ip-config-unavailable` and `no lease`.
- The libvirt network was active and had a DHCP range configured.
- The libvirt `dnsmasq` process was running and bound to `virbr1`.
- `virsh net-dhcp-leases` returned no leases.
- UFW was active. Its configured default input and forwarding policies were `DROP`.
- UFW had no explicit rules for guest DHCP, DNS, or outbound forwarding on this bridge.

After the firewall rules were added, Vagrant reported:

```text
==> default: Waiting for machine to boot. This may take a few minutes...
    default: SSH address: 192.168.121.90:22
==> default: Machine booted and ready!
```

Vagrant also replaced the insecure SSH key, set the hostname, and completed both rsync operations.

## Root cause

The host firewall configuration did not allow the traffic needed by the libvirt guest network. In particular, DHCP requests from the guest needed to reach the host's DHCP service on UDP port 67 through `virbr1`.

Without a DHCP lease, the guest had no usable IPv4 address. Vagrant therefore remained in its IP-address discovery step.

The successful startup after adding the rules supports this diagnosis. No packet capture or DHCP-specific firewall drop log was collected. All three rules were applied together, so their effects were not tested independently. DHCP permission directly addresses the observed wait; DNS and forwarding permissions support later guest network access.

## Fix

Run these commands on the **host**, not inside the guest:

```bash
# Allow guest DHCP requests to the host's DHCP server.
sudo ufw allow in on virbr1 proto udp to any port 67

# Allow guest DNS queries to the host, over UDP and TCP.
sudo ufw allow in on virbr1 from 192.168.121.0/24 to any port 53

# Allow guest traffic to leave through the host's outbound interface.
sudo ufw route allow in on virbr1 out on wlan0 from 192.168.121.0/24

vagrant up --provider=libvirt
```

These rules keep UFW enabled. They do not change the global input or forwarding policy. Libvirt supplies the NAT configuration; the forwarding rule permits traffic through UFW and does not itself create NAT.

UFW stores these rules persistently. No separate `ufw reload` is needed after successful `ufw allow` commands.

## Verification

On the host:

```bash
sudo ufw status verbose
virsh -c qemu:///system net-dhcp-leases vagrant-libvirt
vagrant status
vagrant ssh -c 'ip -4 addr; ip route'
```

Expected results:

- UFW lists the bridge-specific rules.
- Libvirt lists a guest DHCP lease.
- Vagrant reports the VM as running.
- SSH works and the guest has an IPv4 address in `192.168.121.0/24`.

To check DNS and external connectivity separately:

```bash
vagrant ssh -c 'getent hosts dl.rockylinux.org'
vagrant ssh -c 'curl -I --connect-timeout 10 https://dl.rockylinux.org/'
```

These last checks were not recorded during the incident. Successful boot and rsync confirm DHCP and host-to-guest SSH, not external connectivity.

## Diagnostic commands for recurrence

```bash
# Check the guest state and host network configuration.
virsh -c qemu:///system list --all
virsh -c qemu:///system net-list --all
virsh -c qemu:///system net-dumpxml vagrant-libvirt
virsh -c qemu:///system net-dhcp-leases vagrant-libvirt
ip -br addr
ip route show default
sudo ufw status verbose

# Check the DHCP service and available firewall logs.
journalctl -b -t dnsmasq-dhcp --no-pager
sudo journalctl -k -b --no-pager --grep='UFW.*virbr1'
```

If there is still no lease, use the guest console to check boot and DHCP status:

```bash
virsh -c qemu:///system console kernelscope_default
```

Inside the guest:

```bash
ip -br addr
nmcli device status
sudo journalctl -b -u NetworkManager --no-pager -n 50
```

Press `Ctrl+]` to leave the libvirt console. A running VM alone does not prove that the guest finished booting.

## Rollback

To remove only the rules added for this fix:

```bash
sudo ufw delete allow in on virbr1 proto udp to any port 67
sudo ufw delete allow in on virbr1 from 192.168.121.0/24 to any port 53
sudo ufw delete route allow in on virbr1 out on wlan0 from 192.168.121.0/24
```

Do not remove rules that other VMs need. Removing these permissions can cause DHCP or guest connectivity failures again.

## Prevention

- Keep the required UFW rules with the host setup instructions.
- Recheck the forwarding rule when the host switches outbound interfaces, such as from Wi-Fi to Ethernet.
- Recheck bridge and subnet values if the libvirt network is recreated.
- Do not disable UFW globally to solve a guest DHCP problem.
- Do not recreate a VM solely because it has no DHCP lease. First check guest boot status, DHCP service status, and host firewall rules.
