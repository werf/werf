require_relative '../spec_helper'

describe Dapp::Dimg::Filelock do
  include SpecHelper::Dimg

  it 'images locks' do
    config[:_docker][:_from_cache_version] = :lock_spec
    config[:_builder] = :shell
    config[:_shell].keys.each { |stage| config[:_shell][stage] << "date +%s > /#{stage}" }
    dimg_build!

    expect(dimg.tagged_images).to_not be_empty
    dimg.tagged_images.each do |image|
      path = File.expand_path(File.join(dimg.dapp.build_path, 'locks', "#{dimg.dapp.name}.image.#{image.name}".slice(0, 32)))
      expect(File.exist?(path)).to be_truthy
    end
  end
end
