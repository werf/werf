module Dapp
  module Helper
    module Url
      def git_url_to_name(url)
        url_without_scheme = url.split("://", 2).last
        # This may be broken, because "@" should delimit creds, not a ":"
        url_without_creds = url_without_scheme.split(":", 2).last
        url_without_creds.gsub(%r{.*?([^\/ ]+\/[^\/ ]+)\.git}, '\\1')
      end

      def get_host_from_git_url(url)
        url_without_scheme = url.split("://", 2).last
        url_without_creds = url_without_scheme.split("@", 2).last

        # Split out part after ":" in this kind of url: github.com:flant/dapp.git
        url_part = url_without_creds.split(":", 2).first

        # Split out part after first "/": github.com/flant/dapp.git
        url_part.split("/", 2).first
      end
    end # Url
  end # Helper
end # Dapp
