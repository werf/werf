module Dapp
  # DockerRegistry
  module DockerRegistry
    def self.new(repo)
      repo_regex =~ repo
      expected_hostname = Regexp.last_match(1)
      expected_repo_suffix = Regexp.last_match(2)
      expected_hostname_url = expected_hostname ? "http://#{expected_hostname}" : nil

      if hostname_exist?(expected_hostname_url)
        Base.new(repo, expected_hostname_url, expected_repo_suffix)
      else
        Default.new(repo, File.join(*[expected_hostname, expected_repo_suffix].compact))
      end
    end

    def self.repo_regex
      separator = /[_.]|__|[-]*/
      alpha_numeric = /[[:alnum:]]*/
      component = /#{alpha_numeric}[#{separator}#{alpha_numeric}]*/
      port_number = /[0-9]+/
      hostcomponent = /[[:alnum:]-]*[[:alnum:]]/
      hostname = /#{hostcomponent}[\.#{hostcomponent}]*[:#{port_number}]?/
      %r{^(#{hostname}/)?(#{component}[/#{component}]*)$}
    end

    def self.hostname_exist?(url)
      return false unless url
      Mod::Request.url_available?(url)
    end
  end # DockerRegistry
end # Dapp
