class Peaktop < Formula
  desc "Apple Silicon system monitor for the terminal"
  homepage "https://github.com/brodie/peaktop"
  url "https://github.com/brodie/peaktop/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"peaktop", "./cmd/peaktop/"
  end

  test do
    system bin/"peaktop", "-h"
  end
end
