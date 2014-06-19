# -*- encoding: utf-8 -*-
$:.push File.expand_path("../lib", __FILE__)

Gem::Specification.new do |s|
  s.name    = "vagrant-bosh"
  s.version = "0.0.1"

  s.homepage    = "https://github.com/cppforlife/vagrant-bosh"
  s.summary     = %q{Vagrant BOSH provisioner plugin.}
  s.description = %q{BOSH provisioner allows to provision guest VM by specifying regular BOSH deployment manifest.}

  s.authors  = ["Dmitriy Kalinin"]
  s.email    = ["cppforlife@gmail.com"]
  s.licenses = ["MIT"]

  s.files         = `git ls-files`.split("\n")
  s.test_files    = `git ls-files -- {test,spec,features}/*`.split("\n")
  s.executables   = `git ls-files -- bin/*`.split("\n").map{ |f| File.basename(f) }
  s.require_paths = ["lib", "templates"]
end
