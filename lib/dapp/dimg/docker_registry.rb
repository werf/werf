module Dapp
  module Dimg
    module DockerRegistry
      def self.new(dapp, repo)
        Dimg.new(dapp, repo)
      end

      def self.repo_name_format
        rpart = '[a-z0-9]+(([_.]|__|-+)[a-z0-9]+)*'
        hpart = '(?!-)[a-z0-9-]+(?<!-)'
        "(?<hostname>#{hpart}(\\.#{hpart})*(?<port>:[0-9]+)?\/)?(?<repo_suffix>#{rpart}(\/#{rpart})*)"
      end

      def self.repo_name?(name)
        !(/^#{repo_name_format}$/ =~ name).nil?
      end
    end # DockerRegistry
  end # Dimg
end # Dapp
