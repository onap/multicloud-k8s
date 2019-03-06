# -*- mode: ruby -*-
# vi: set ft=ruby :

vars = {
  "demo_artifacts_version"     => "1.3.0",
  'vfw_private_ip_0'           => '192.168.10.100',
  'vfw_private_ip_1'           => '192.168.20.100',
  'vfw_private_ip_2'           => '10.10.100.2',
  'vpg_private_ip_0'           => '192.168.10.200',
  'vpg_private_ip_1'           => '10.0.100.3',
  'vsn_private_ip_0'           => '192.168.20.250',
  'vsn_private_ip_1'           => '10.10.100.4',
  'dcae_collector_ip'          => '10.0.4.1',
  'dcae_collector_port'        => '8081',
  'protected_net_gw'           => '192.168.20.100',
  'protected_net_cidr'         => '192.168.20.0/24',
  'protected_private_net_cidr' => '192.168.10.0/24',
  'onap_private_net_cidr'      => '10.10.0.0/16'
}

if ENV['no_proxy'] != nil or ENV['NO_PROXY']
  $no_proxy = ENV['NO_PROXY'] || ENV['no_proxy'] || "127.0.0.1,localhost"
  $subnet = "192.168.121"
  # NOTE: This range is based on vagrant-libivirt network definition
  (1..27).each do |i|
    $no_proxy += ",#{$subnet}.#{i}"
  end
end

Vagrant.configure("2") do |config|
  config.vm.box = "elastic/ubuntu-16.04-x86_64"

  if ENV['http_proxy'] != nil and ENV['https_proxy'] != nil
    if not Vagrant.has_plugin?('vagrant-proxyconf')
      system 'vagrant plugin install vagrant-proxyconf'
      raise 'vagrant-proxyconf was installed but it requires to execute again'
    end
    config.proxy.http     = ENV['http_proxy'] || ENV['HTTP_PROXY'] || ""
    config.proxy.https    = ENV['https_proxy'] || ENV['HTTPS_PROXY'] || ""
    config.proxy.no_proxy = $no_proxy
  end

  config.vm.provider 'libvirt' do |v|
    v.cpu_mode = 'host-passthrough' # DPDK requires Supplemental Streaming SIMD Extensions 3 (SSSE3)
  end

  config.vm.define :packetgen do |packetgen|
    packetgen.vm.hostname = "packetgen"
    packetgen.vm.provision 'shell', path: 'packetgen', env: vars
    packetgen.vm.network :private_network, :ip => vars['vpg_private_ip_0'], :type => :static, :netmask => "255.255.255.0" # unprotected_private_net_cidr
    packetgen.vm.network :private_network, :ip => vars['vpg_private_ip_1'], :type => :static, :netmask => "255.255.0.0" # onap_private_net_cidr
  end	
  config.vm.define :firewall do |firewall|
    firewall.vm.hostname = "firewall"
    firewall.vm.provision 'shell', path: 'firewall', env: vars
    firewall.vm.network :private_network, :ip => vars['vfw_private_ip_0'], :type => :static, :netmask => "255.255.255.0" # unprotected_private_net_cidr
    firewall.vm.network :private_network, :ip => vars['vfw_private_ip_1'], :type => :static, :netmask => "255.255.255.0" # protected_private_net_cidr
    firewall.vm.network :private_network, :ip => vars['vfw_private_ip_2'], :type => :static, :netmask => "255.255.0.0" # onap_private_net_cidr
  end
  config.vm.define :sink do |sink|
    sink.vm.hostname = "sink"
    sink.vm.provision 'shell', path: 'sink', env: vars
    sink.vm.network :private_network, :ip => vars['vsn_private_ip_0'], :type => :static, :netmask => "255.255.255.0" # protected_private_net_cidr
    sink.vm.network :private_network, :ip => vars['vsn_private_ip_1'], :type => :static, :netmask => "255.255.0.0" # onap_private_net_cidr
  end
end
