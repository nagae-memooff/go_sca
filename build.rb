#!/usr/bin/env ruby
#encoding=utf-8
progname = "demo"

case ARGV.first
when "deploy"
  IPS = [202, 203, 232]

  FILES = %W{
  id_factory_local
  }

  IPS.each do |_ip|
    ip = if _ip.class == Fixnum
           "root@192.168.100.#{_ip}"
         else
           _ip
         end

    system "ssh #{ip} 'su - ewhine -c \"monit unmonitor all\"'"
    system "ssh #{ip} 'su - ewhine -c \"/etc/init.d/mx_id_factory_local stop\"'"

    FILES.each do |file|
      system "scp #{file} #{ip}:/home/ewhine/deploy/id_factory_local/#{file}"
    end

    system "ssh #{ip} 'su - ewhine -c \"/etc/init.d/mx_id_factory_local start\"'"
  end

when "first_install"
  system "mkdir -p /home/ewhine/deploy/#{progname}"

  system "mv mx_#{progname} /etc/init.d/"
  system "mv #{progname} #{progname}.conf start-stop-daemon version /home/ewhine/deploy/#{progname}/"

  system "chkconfig mx_#{progname} on"
  system "chown ewhine:ewhine -R /home/ewhine/deploy/#{progname}"

  puts "done. please set monit && start #{progname} service if needed."
when "install"
  system "/etc/init.d/mx_#{progname} stop"
  system "cd /home/ewhine/ewhine_pkg/ && mv #{progname} start-stop-daemon version /home/ewhine/deploy/#{progname}/ && mv #{progname}.conf /home/ewhine/deploy/#{progname}/#{progname}.demo.conf"
  system "/etc/init.d/mx_#{progname} start"

  puts "done."
end
