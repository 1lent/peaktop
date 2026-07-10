class Peaktop < Formula
  desc "Apple Silicon system monitor for the terminal"
  homepage "https://github.com/1lent/peaktop"
  url "https://github.com/1lent/peaktop/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "864ef3f666f668dd175e3f04c0ca490ae807e56b6f712093115fb55ee04249e8"
  license "MIT"
  version "0.1.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"peaktop", "./cmd/peaktop/"
  end

  test do
    assert_match "Usage", shell_output("#{bin}/peaktop -h 2>&1", 1)
  end
end
