require_relative '../spec_helper'

describe Dapp::Dimg::Image::Docker do
  context 'positive' do
    [
      %w(image image),
      %w(image:tag.012 image tag.012),
      %w(docker_registry:8000/image:tag image tag docker_registry:8000/)
    ].each do |str, repo_suffix, tag, hostname|
      it str do
        str =~ %r{^#{Dapp::Dimg::Image::Docker.image_name_format}$}
        expect(hostname).to eq Regexp.last_match(:hostname)
        expect(repo_suffix).to eq Regexp.last_match(:repo_suffix)
        expect(tag).to eq Regexp.last_match(:tag)
      end
    end
  end

  context 'negative' do
    %w(image: image:tag:tag image:-tag).each do |image|
      it image do
        expect(Dapp::Dimg::Image::Docker.image_name?(image)).to be_falsey
      end
    end
  end
end
