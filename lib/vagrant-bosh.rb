require "i18n"
require "pathname"
require "vagrant-bosh/plugin"

module VagrantPlugins
  module VagrantBosh
    lib_path = Pathname.new(File.expand_path("../vagrant-bosh", __FILE__))
    autoload :Errors, lib_path.join("errors")

    @source_root = Pathname.new(File.expand_path("../../", __FILE__))

    I18n.load_path << File.expand_path("templates/locales/en.yml", @source_root)
    I18n.reload!
  end
end
