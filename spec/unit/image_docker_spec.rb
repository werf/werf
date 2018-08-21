require_relative '../spec_helper'

describe Dapp::Dimg::Image::Stage do
  context 'positive' do
    [
      %w(i i),
      %w(i-m i-m),
      %w(image:tag.012 image tag.012),
      %w(docker-registry:8000/image:tag image tag docker-registry:8000/)
    ].each do |str, repo_suffix, tag, hostname|
      it str do
        str =~ %r{^#{Dapp::Dimg::Image::Stage.image_name_format}$}
        expect(Regexp.last_match(:hostname)).to eq hostname
        expect(Regexp.last_match(:repo_suffix)).to eq repo_suffix
        expect(Regexp.last_match(:tag)).to eq tag
      end
    end
  end

  context 'negative' do
    %w(image: image:tag:tag image:-tag image:.tag Image:tag).each do |image|
      it image do
        expect(Dapp::Dimg::Image::Stage.image_name?(image)).to be_falsey
      end
    end
  end
end
