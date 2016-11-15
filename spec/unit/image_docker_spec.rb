require_relative '../spec_helper'

describe Dapp::Image::Docker do
  context 'positive' do
    %w(image:tag image).each do |image|
      it "#{image}" do
        expect(image =~ Dapp::Image::Docker.image_regex).to_not be_nil
      end
    end
  end

  context 'negative' do
    %w(image: image:tag:tag).each do |image|
      it "#{image}" do
        expect(image =~ Dapp::Image::Docker.image_regex).to be_nil
      end
    end
  end
end
