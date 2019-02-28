# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "elastic/ubuntu-16.04-x86_64"
  config.vm.hostname = "demo"
  config.vm.provision 'shell', path: 'postinstall.sh'
  config.vm.network :private_network, :ip => "192.168.10.5", :type => :static # unprotected_private_net_cidr
  config.vm.network :private_network, :ip => "192.168.20.5", :type => :static # protected_private_net_cidr
  config.vm.network :private_network, :ip => "10.10.12.5", :type => :static, :netmask => "16" # onap_private_net_cidr

  if ENV['http_proxy'] != nil and ENV['https_proxy'] != nil
    if not Vagrant.has_plugin?('vagrant-proxyconf')
      system 'vagrant plugin install vagrant-proxyconf'
      raise 'vagrant-proxyconf was installed but it requires to execute again'
    end
    config.proxy.http     = ENV['http_proxy'] || ENV['HTTP_PROXY'] || ""
    config.proxy.https    = ENV['https_proxy'] || ENV['HTTPS_PROXY'] || ""
    config.proxy.no_proxy = ENV['NO_PROXY'] || ENV['no_proxy'] || "127.0.0.1,localhost"
    config.proxy.enabled = { docker: false }
  end

  config.vm.provider 'virtualbox' do |v|
    v.customize ["modifyvm", :id, "--memory", 8192]
    v.customize ["modifyvm", :id, "--cpus", 2]
  end
  config.vm.provider 'libvirt' do |v|
    v.memory = 8192
    v.cpus = 2
    v.nested = true
    v.cpu_mode = 'host-passthrough'
  end
end
