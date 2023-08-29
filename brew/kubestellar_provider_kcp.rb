class KubestellarProviderKcp < Formula
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

  depends_on "kubestellar_provider_kcp_kubectl"

  def install
    port_to_check = 6443
    port_in_use = port_open?("#{port_to_check}")
    
    prefix.install Dir["*"]
  end

  def port_open?(port)
    port_info = `lsof -i -n -P | grep LISTEN | grep :#{port}`.strip
    # `lsof -i :#{port}`
    if port_info.empty?
      # puts "Port #{port} is not in use."
    else
      puts "\nPlease remove the process running on #{port} and re-run this formula"
      puts "Port #{port} is in use by the following command:"
      puts "#{port_info}\n\n"
      odie
    end
  end

  def post_install
    puts "\e[1;37mKCP binary has been installed to '#{prefix}' and are symlinked to '#{HOMEBREW_PREFIX}/bin'\e[0m"
    current_user = `whoami`.strip
    if OS.mac?
      kcp_bin_path = "sudo -u #{current_user} #{HOMEBREW_PREFIX}/bin/kcp start &> /tmp/kcp.log &"  # Replace with your binary name
      # ohai "Do you want to start 'kcp'? (y/n)"
      # response = $stdin.gets.chomp.downcase
      # if response == "y" || response == "yes"
        system "osascript", "-e", <<-EOS
          do shell script "#{kcp_bin_path}" with administrator privileges
            do shell script "sleep 10"            
          EOS
      # end
      export_kubeconfig = `export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig`
      if $?.success?
        puts "kubeconfig exported successfully."
      else
        puts "kubeconfig export failed."
      end

      max_attempts = 20
      attempts = 0
      success = false
      
      while attempts < max_attempts && !success
        kubectl_ws_tree = `export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig ; kubectl ws tree &> /dev/null`
  
        if $?.success?
          puts "'kubectl ws tree' succeeded. KCP is now installed and running properly."
          success = true
        else
          if attempts == max_attempts
            odie "'kubectl ws tree' failed. Please remove this formula and attempt to install it again"
          end
          attempts += 1
          sleep   5
        end
      end

    elsif OS.linux?
      kcp_bin_path = "su -c #{HOMEBREW_PREFIX}/bin/kcp start #{current_user} &> /tmp/kcp.log &"  # Replace with your binary name
    end
    
    puts "\n\e[1;37mConnecting to the KCP control plane is easy:
        export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
        kubectl ws tree
        \e[0m"
  end

  test do
    system "false"
  end
end

