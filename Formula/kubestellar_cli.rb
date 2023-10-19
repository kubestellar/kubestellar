class KubestellarCli < Formula
  desc "KubeStellar user facing commands and kubectl plugins"
  homepage "https://kubestellar.io"
  version "v0.8.0"

  if OS.mac?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.8.0/kubestellaruser_v0.8.0_darwin_amd64.tar.gz"
      sha256 "358c3f051d2badfc5ddf40262aa514dd52f31bed806ff55cf55e101dc4eec79c"
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.8.0/kubestellaruser_v0.8.0_darwin_arm64.tar.gz"
      sha256 "f2ed69415f78f38e3893a5aa908f229a8dc52f71a321f5ff355766311e03a5c9"  
    else
      odie "Unsupported architecture on macOS"
    end
  elsif OS.linux?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.8.0/kubestellaruser_v0.8.0_linux_amd64.tar.gz"
      sha256 "26c7fe7499e8c9e15147651f58c588764a88990649bef006df64ca74cc9ceb2e"
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.8.0/kubestellaruser_v0.8.0_linux_arm64.tar.gz"
      sha256 "9ff7ddfcb350e8f5e2678b9eb2bc2fd783218413abc058e92e7e8562c14fc3a2"
    when :s390x
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.8.0/kubestellaruser_v0.8.0_linux_s390x.tar.gz"
      sha256 "25c00b9855d17a475d9e93d3950dacea8c0a5db315db20c37ea209064189e038"
    when :ppc64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.8.0/kubestellaruser_v0.8.0_linux_ppc64le.tar.gz"
      sha256 "6c34ffb1faf789b20ed2c32d9171c0e43edc444573689eec9b313cf041d249e2"
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
