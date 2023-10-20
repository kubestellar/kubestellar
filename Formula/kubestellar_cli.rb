class KubestellarCli < Formula
  desc "KubeStellar user facing commands and kubectl plugins"
  homepage "https://kubestellar.io"
  version "v0.9.0"

  if OS.mac?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.9.0/kubestellaruser_v0.9.0_darwin_amd64.tar.gz"
      sha256 "081deadc66cf55d8672a8991106c238fb9f52c33a0b7f128363dfef8386f50b1"
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.9.0/kubestellaruser_v0.9.0_darwin_arm64.tar.gz"
      sha256 "baca7cac1225225329df960376beae170bcddc395855318750c214dd50c736c1"  
    else
      odie "Unsupported architecture on macOS"
    end
  elsif OS.linux?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.9.0/kubestellaruser_v0.9.0_linux_amd64.tar.gz"
      sha256 "184148db34c488fdac038e62ecb0614e07e6092359f3da1c24876dc5641d5ae7"
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.9.0/kubestellaruser_v0.9.0_linux_arm64.tar.gz"
      sha256 "7422a6c8336c5341c6dab2c364fa308ff5b95b56bbd9e7ffa4d8dfd7536ec552"
    when :s390x
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.9.0/kubestellaruser_v0.9.0_linux_s390x.tar.gz"
      sha256 "d97e9cc4b5ffdd211702bdc5e0eebac3e24f3cfca41f8736d7a9c55e48b58cb6"
    when :ppc64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.9.0/kubestellaruser_v0.9.0_linux_ppc64le.tar.gz"
      sha256 "46c3c07a15db9a366b25476b21ec79b14c9e28db0629976e93cdb45f2545068c"
    else
      odie "Unsupported architecture on Linux"
    end
  else
    odie "Unsupported operating system"
  end

  license "Apache-2.0"

  if system("which kubectl &> /dev/null")
    depends_on "kubectl"
  end

  def install
    prefix.install Dir["*"]
  end

  def post_install
    puts "\e[1;37mKubeStellar user commands and kubectl plugins have been installed to '#{prefix}' and are symlinked to '#{HOMEBREW_PREFIX}/bin'\e[0m"
  end

  test do
    # `test do` will create, run in and delete a temporary directory.
    #
    # This test will fail and we won't accept that! For Homebrew/homebrew-core
    # this will need to be a test that verifies the functionality of the
    # software. Run the test with `brew test kubestellar cli`. Options passed
    # to `brew install` such as `--HEAD` also need to be provided to `brew test`.
    #
    # The installed folder is not in the path, so use the entire path to any
    # executables being tested: `system "#{bin}/program", "do", "something"`.
    system "false"
  end
end
