module Dapp
  module Dimg
    module DockerRegistry
      def self.new(repo)
        /^#{repo_name_format}$/ =~ repo
        expected_hostname = Regexp.last_match(:hostname)
        expected_repo_suffix = Regexp.last_match(:repo_suffix)
        expected_hostname_url = expected_hostname ? "http://#{expected_hostname}" : nil

        if hostname_exist?(expected_hostname_url)
          Base.new(repo, expected_hostname_url, expected_repo_suffix)
        else
          Default.new(repo, File.join(*[expected_hostname, expected_repo_suffix].compact))
        end
      end

      def self.repo_name_format
        separator = '[_.]|__|[-]*'
        alpha_numeric = '[[:alnum:]]*'
        component = "#{alpha_numeric}[#{separator}#{alpha_numeric}]*"
        port_number = '[[:digit:]]+'
        hostcomponent = '[[:alnum:]_-]*[[:alnum:]]'
        hostname = "#{hostcomponent}[\\.#{hostcomponent}]*(?<port>:#{port_number})?"
        "(?<hostname>#{hostname}/)?(?<repo_suffix>#{component}[/#{component}]*)"
      end

      def self.repo_name?(name)
        !(/^#{repo_name_format}$/ =~ name).nil?
      end

      def self.hostname_exist?(url)
        return false unless url
        Base.url_available?(url)
      end
    end # DockerRegistry
  end # Dimg
end # Dapp
