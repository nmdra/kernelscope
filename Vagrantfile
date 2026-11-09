Vagrant.configure("2") do |config|
  config.vm.box = "kernelscope-rocky10"

  config.vm.box_url =
    "https://dl.rockylinux.org/pub/rocky/10.2/images/x86_64/Rocky-10-Vagrant-Libvirt-10.2-20260525.0.x86_64.vagrant.libvirt.box"

  config.vm.hostname = "kernelscope"

  config.vm.provider "libvirt" do |vm|
    vm.cpus = 2
    vm.memory = 2048
    # The Rocky 10.2 box reports 5 GiB in metadata, but its image is 10 GiB.
    vm.machine_virtual_size = 10
  end

  # Disable the default /vagrant share
  config.vm.synced_folder ".", "/vagrant", disabled: true

  # Sync source code from host to VM
  config.vm.synced_folder ".", "/home/vagrant/kernelscope",
    type: "rsync",
    rsync__exclude: [
      ".git/",
      ".vagrant/",
      "bin/",
      "*.o"
    ]
end

Vagrant.configure("2") do |config|
  # Keep your existing Rocky Linux box configuration
  # and libvirt provider settings here.

  # Disable the default /vagrant share
  config.vm.synced_folder ".", "/vagrant", disabled: true

  # Sync source code from host to VM
  config.vm.synced_folder ".", "/home/vagrant/kernelscope",
    type: "rsync",
    rsync__exclude: [
      ".git/",
      ".vagrant/",
      "bin/",
      "*.o"
    ]
end
