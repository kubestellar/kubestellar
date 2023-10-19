class KcpCli < Formula
    desc "Simplifying building massively multi-tenant services. Together."
    homepage "https://kcp.io"
    version "v0.11.0"
  
    if OS.mac?
      case Hardware::CPU.arch
      when :x86_64
        url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kubectl-kcp-plugin_0.11.0_darwin_amd64.tar.gz"
        sha256 "c16a426f1c92e6e968b24a8c302e5ba945b81d2473469ecc60fe3647a1651fc3"
      when :arm64
        url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kubectl-kcp-plugin_0.11.0_darwin_arm64.tar.gz"
        sha256 "cd8d3b01ec7b1fedd769f6935974a4f4743f8381e57302d014d2e441f8e8360a"  
      else
        odie "Unsupported architecture on macOS"
      end
    elsif OS.linux?
      case Hardware::CPU.arch
      when :x86_64
        url "http://github.com/kcp-dev/kcp/releases/download/v0.11.0/kubectl-kcp-plugin_0.11.0_linux_amd64.tar.gz"
        sha256 "adb2a035015af424a3e2ff9a848a2812354de87a532259947dd77fe49eba2dfe"
      when :arm64
        url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kubectl-kcp-plugin_0.11.0_linux_arm64.tar.gz"
        sha256 "c34da195c053613673f4db5b27ab5081e269196aa9e2e27bc5572abeebb6ff2b"
      when :s390x
        url ""
        sha256 ""
      when :ppc64
        url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kubectl-kcp-plugin_0.11.0_linux_ppc64le.tar.gz"
        sha256 "e54457f1284e54e6eff9b00aa9d3ca06dc45e7475a0a2de090f3e8322b6f2037"
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
      bin.install Dir["*"]
    end
  
    def post_install
      puts "\e[1;37mKCP kubectl plugins have been installed to '#{prefix}' and are symlinked to '#{HOMEBREW_PREFIX}/bin'\e[0m"
    end
  
    test do
      # `test do` will create, run in and delete a temporary directory.
      #
      # This test will fail and we won't accept that! For Homebrew/homebrew-core
      # this will need to be a test that verifies the functionality of the
      # software. Run the test with `brew test kcp`. Options passed
      # to `brew install` such as `--HEAD` also need to be provided to `brew test`.
      #
      # The installed folder is not in the path, so use the entire path to any
      # executables being tested: `system "#{bin}/program", "do", "something"`.
      system "false"
    end
  end