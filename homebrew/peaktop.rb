class Peaktop < Formula
  desc "Apple Silicon system monitor for the terminal"
  homepage "https://github.com/1lent/peaktop"
  url "https://github.com/1lent/peaktop/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_SHA256_FROM_RELEASE"
  license "MIT"
  version "0.1.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"peaktop", "./cmd/peaktop/"
  end

  test do
    system bin/"peaktop", "-i", "100"
    sleep 1
  end
end
