# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "generic/ubuntu1804"
  config.vm.hostname = "ubuntu18"
  config.vm.synced_folder ".", "/vagrant"
  config.vm.provider :libvirt do |libvirt|
    libvirt.graphics_ip = '0.0.0.0'
    # add random suffix to allow running multiple jobs
    libvirt.random_hostname = 'yes'
    libvirt.cpus = 16
    libvirt.memory = 32768
    libvirt.machine_virtual_size = 400
  end
end
