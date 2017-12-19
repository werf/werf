module Dapp
  module Dimg
    module DockerRegistry
      def self.new(repo)
        /^#{repo_name_format}$/ =~ repo
        expected_hostname = Regexp.last_match(:hostname)
        expected_repo_suffix = Regexp.last_match(:repo_suffix)

        if expected_hostname
          %w(https http).each do |protocol|
            expected_hostname_url = [protocol, expected_hostname].join('://')
            return Dimg.new(repo, expected_hostname_url, expected_repo_suffix) if hostname_exist?(expected_hostname_url)
          end
          raise Error::Registry, code: :registry_not_available, data: { registry: repo }
        else
          Default.new(repo, expected_repo_suffix)
        end
      end

      def self.repo_name_format
        rpart = '[a-z0-9]+(([_.]|__|-+)[a-z0-9]+)*'
        hpart = '(?!-)[a-z0-9-]+(?<!-)'
        "(?<hostname>#{hpart}(\\.#{hpart})*(?<port>:[0-9]+)?\/)?(?<repo_suffix>#{rpart}(\/#{rpart})*)"
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
