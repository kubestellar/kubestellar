class Kcp < Formula
  desc "Simplifying building massively multi-tenant services. Together."
  homepage "https://kcp.io"
  version "v0.11.0"

  if OS.mac?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kcp_0.11.0_darwin_amd64.tar.gz"
      sha256 "23737dd3201a40cb0326d5315ba6f56c1ac28daed692369567b1251b19ddb0fb"
    when :arm64
      url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kcp_0.11.0_darwin_arm64.tar.gz"
      sha256 "4aad3739af186c975410bb37c113de005b0fb99f12c4848947abfb4d20801320"  
    else
      odie "Unsupported architecture on macOS"
    end
  elsif OS.linux?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kcp_0.11.0_linux_amd64.tar.gz"
      sha256 "8744f863232b1ea202598e286cd061fad81b181798fb1e6612b46df5b6423524"
    when :arm64
      url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kcp_0.11.0_linux_arm64.tar.gz"
      sha256 "3967bf1c7b08e0564690e6927a7fb11baa747726a6f6e3e0ddeae69cbcdaac0a"
    when :s390x
      url ""
      sha256 ""
    when :ppc64
      url "https://github.com/kcp-dev/kcp/releases/download/v0.11.0/kcp_0.11.0_linux_ppc64le.tar.gz"
      sha256 "22941381ee8252795254a0622b5c431a7b36f078525b61a5171236bd00ed0e84"
    else
      odie "Unsupported architecture on Linux"
    end
  else
    odie "Unsupported operating system"
  end

  license "Apache-2.0"

  depends_on "kubectl"
  depends_on "yq"
  depends_on "jq"

  def install
    port_to_check = 6443
    port_in_use = port_open?("#{port_to_check}")
    if port_in_use
      odie "Port 6443 is already in use. Please free up the port before installing."
    end
    prefix.install Dir["*"]
  end

  def port_open?(port)
    `lsof -i :#{port}`
    $?.success?
  end

  def post_install
    puts "\e[1;37mKCP binary has been installed to '#{prefix}' and are symlinked to '#{HOMEBREW_PREFIX}/bin'\e[0m"
    kcp_bin_path = "#{HOMEBREW_PREFIX}/bin/kcp start &> /dev/null &"  # Replace with your binary name
    current_user = `whoami`.strip
    # puts "#{HOMEBREW_PREFIX}"
    # puts "#{current_user}"
    ohai "Do you want to start 'kcp'? (y/n)"
    response = $stdin.gets.chomp.downcase
    if response == "y" || response == "yes"
      system "osascript", "-e", <<-EOS
        do shell script "#{kcp_bin_path}" with administrator privileges
        do shell script "sleep 3"
        do shell script "while [ -e '$(pwd)/.kcp/admin.kubeconfig' ]; do sleep 1; done"
        do shell script "sleep 5"
        do shell script "chown -R #{current_user}: $(pwd)/.kcp" with administrator privileges
        EOS
    end
    puts "\e[1;37mConnecting to the KCP control plane is easy:
        export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
        kubectl ws tree
        \e[0m"
  end

  test do
    system "false"
  end
end

