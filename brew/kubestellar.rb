class Kubestellar < Formula
  desc "KubeStellar is a flexible solution for challenges associated with multi-cluster configuration management for edge, multi-cloud, and hybrid cloud"
  homepage "https://kubestellar.io"
  version "v0.6.0"

  if OS.mac?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.6.0/kubestellar_v0.6.0_darwin_amd64.tar.gz"
      sha256 "310f6724555da8e243d624c6519989acb1b95583a5d0657883f497b64d63d375"
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.6.0/kubestellar_v0.6.0_darwin_arm64.tar.gz"
      sha256 "590f5867a6bba6d8dbfc93b5c4c19c3acbbc592df8544ef48dcd737cceab443b"  
    else
      odie "Unsupported architecture on macOS"
    end
  elsif OS.linux?
    case Hardware::CPU.arch
    when :x86_64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.6.0/kubestellar_v0.6.0_linux_amd64.tar.gz"
      sha256 "14aaeb105c77b2e8466c48553d2f1b97394fd5b27caaf7358adcc9e7108de247"
    when :arm64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.6.0/kubestellar_v0.6.0_linux_arm64.tar.gz"
      sha256 "bb80d33d1a62fe5b3aaf0cecee2f12ec92e2e78d5a0f5678eef08996ba29ce71"
    when :s390x
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.6.0/kubestellar_v0.6.0_linux_s390x.tar.gz"
      sha256 "303937fa034bb307e6db8f6202ac3f6a79dd9ffa22d8f7269569bd8d723a804e"
    when :ppc64
      url "https://github.com/kubestellar/kubestellar/releases/download/v0.6.0/kubestellar_v0.6.0_linux_ppc64le.tar.gz"
      sha256 "0b9ec309912a0d74a562e8dbe6b9e0fa9ed0384cb08ad132cebb9293a7b11ca0"
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
  depends_on "helm"
  # depends_on "kubestellar_provider_kcp"
  # depends_on "kubestellar_provider_kcp_kubectl"

  def install
    
    prefix.install Dir["*"]
  end

  def post_install
    ENV["KUBECONFIG"] = ".kcp/admin.kubeconfig"
    show_tree = `kubectl ws tree`
    puts "#{show_tree}"
    if !$?.success?
      puts "'kubectl ws tree' failed. Please remove this formula and attempt to install it again"
    end
    switch_to_root_compute = `kubectl ws use root`
    puts "#{switch_to_root_compute}"
    if !$?.success?
      puts "'kubectl ws root:compute' failed. Please remove this formula and attempt to install it again"
    end
    puts "\e[1;37mKubeStellar kubectl extensions have been installed to '#{prefix}/bin' and are symlinked to '#{HOMEBREW_PREFIX}/bin'\e[0m"
    puts "\e[1;37mAll other files have been installed to '#{prefix}'\e[0m\n\n"
  end

  test do
    # `test do` will create, run in and delete a temporary directory.
    #
    # This test will fail and we won't accept that! For Homebrew/homebrew-core
    # this will need to be a test that verifies the functionality of the
    # software. Run the test with `brew test kubestellar`. Options passed
    # to `brew install` such as `--HEAD` also need to be provided to `brew test`.
    #
    # The installed folder is not in the path, so use the entire path to any
    # executables being tested: `system "#{bin}/program", "do", "something"`.
    system "false"
  end
end
