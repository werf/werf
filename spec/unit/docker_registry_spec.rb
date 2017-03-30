require_relative '../spec_helper'

describe Dapp::Dimg::DockerRegistry do
  context 'positive' do
    [
      %w(repo repo),
      %w(hostname:1234/repo repo hostname:1234/),
      %w(subdomain.hostname:1234/sub_repo/repo sub_repo/repo subdomain.hostname:1234/)
    ].each do |str, repo_suffix, hostname|
      it str do
        str =~ %r{^#{Dapp::Dimg::DockerRegistry.repo_name_format}$}
        expect(Regexp.last_match(:hostname)).to eq hostname
        expect(Regexp.last_match(:repo_suffix)).to eq repo_suffix
      end
    end
  end

  context 'negative' do
    %w(hostname.ru:6000 hostname:/repo hostname- Hostname).each do |str|
      it str do
        expect(Dapp::Dimg::DockerRegistry.repo_name?(str)).to be_falsey
      end
    end
  end
end
