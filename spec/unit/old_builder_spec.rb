require_relative '../spec_helper'

describe Dapp::Builder do
  it '.dappfiles_paths', test_construct: true do |example|
    apps = %w(a b c c-x c-y)

    apps.each do |app|
      example.metadata[:construct].directory(app) do |dir|
        dir.file('Dappfile')
        dir.file('Anotherfile') if app.start_with? 'c'
      end
    end

    expect(Dapp::Builder.dappfiles_paths(nil, 'a')).to contain_exactly 'a/Dappfile'
    expect(Dapp::Builder.dappfiles_paths('.', 'b')).to contain_exactly './b/Dappfile'
    expect(Dapp::Builder.dappfiles_paths(nil).size).to eq apps.size
    expect(Dapp::Builder.dappfiles_paths(nil, 'c')).to contain_exactly 'c/Dappfile'
    expect(Dapp::Builder.dappfiles_paths(nil, 'c*')).to contain_exactly 'c/Dappfile', 'c-x/Dappfile', 'c-y/Dappfile'
    expect(Dapp::Builder.dappfiles_paths(nil, 'c-*')).to contain_exactly 'c-x/Dappfile', 'c-y/Dappfile'
    expect(Dapp::Builder.dappfiles_paths(nil, 'c-z')).to contain_exactly 'c/Dappfile'
    expect(Dapp::Builder.dappfiles_paths(nil, 'c*-z')).to contain_exactly 'c/Dappfile', 'c-x/Dappfile', 'c-y/Dappfile'
    expect(Dapp::Builder.dappfiles_paths(nil, 'c-*-z')).to contain_exactly 'c-x/Dappfile', 'c-y/Dappfile'

    Dapp::Builder.default_opts[:dappfile_name] = 'Anotherfile'
    expect(Dapp::Builder.dappfiles_paths(nil)).to contain_exactly 'c/Anotherfile', 'c-x/Anotherfile', 'c-y/Anotherfile'
    expect(Dapp::Builder.dappfiles_paths('.', '*x*')).to contain_exactly './c-x/Anotherfile'
  end
end
