# Homebrew formula for Smart 360 Feedback.
#
# This formula installs the single-binary distribution (Go server with the
# Vue SPA embedded). It does NOT install MongoDB — install and start it
# separately (the easiest path is `brew install mongodb-community` from the
# mongodb/brew tap, or run MongoDB in Docker).
#
# After installing, run:
#   smart360-setup        # writes ~/.config/smart360/.env from the template
#   $EDITOR ~/.config/smart360/.env
#   brew services start smart360
#
# Distribution: this formula lives in https://github.com/mondial7/homebrew-tap
# so end users install with `brew tap mondial7/tap && brew install smart360`.
# The copy in this repo is the source of truth — release automation should
# sync it into the tap repo (see NEXT_STEPS: "Release workflow" issue).
#
# Update procedure when cutting a new release:
#   1. `make release VERSION=vX.Y.Z` to generate dist/*.tar.gz + SHA256SUMS
#   2. Create a GitHub release and upload the tarballs
#   3. Update `version` below + each `sha256` value from SHA256SUMS.txt
#   4. Copy this file into mondial7/homebrew-tap → Formula/smart360.rb
#   5. Test locally: `brew install --build-from-source ./Formula/smart360.rb`

class Smart360 < Formula
  desc "AI-powered anonymous 360-feedback platform (single-binary distribution)"
  homepage "https://github.com/mondial7/smart-360"
  version "0.0.0-REPLACE-ON-RELEASE"
  license "MIT"

  # Replace these URLs and sha256 values from `dist/smart360-vX.Y.Z-SHA256SUMS.txt`
  # after running `make release VERSION=vX.Y.Z`.
  on_macos do
    on_arm do
      url "https://github.com/mondial7/smart-360/releases/download/v#{version}/smart360-v#{version}-darwin-arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
    on_intel do
      url "https://github.com/mondial7/smart-360/releases/download/v#{version}/smart360-v#{version}-darwin-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/mondial7/smart-360/releases/download/v#{version}/smart360-v#{version}-linux-arm64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
    on_intel do
      url "https://github.com/mondial7/smart-360/releases/download/v#{version}/smart360-v#{version}-linux-amd64.tar.gz"
      sha256 "0000000000000000000000000000000000000000000000000000000000000000"
    end
  end

  # MongoDB is intentionally NOT declared as a Homebrew dependency: the
  # official MongoDB Community formula lives in `mongodb/brew` and not every
  # user wants Homebrew to manage it. We document the requirement in the
  # caveats instead.

  def install
    bin.install "smart360"
    pkgshare.install ".env.example"
    (bin/"smart360-setup").write setup_script
    chmod 0755, bin/"smart360-setup"
  end

  # Generates a tiny helper script that copies .env.example to the user's
  # config dir on first run.
  def setup_script
    <<~SH
      #!/usr/bin/env bash
      set -euo pipefail
      target_dir="$HOME/.config/smart360"
      target="$target_dir/.env"
      mkdir -p "$target_dir"
      if [[ -f "$target" ]]; then
        echo "Config already exists at $target"
        exit 0
      fi
      cp "#{pkgshare}/.env.example" "$target"
      chmod 600 "$target"
      echo "Wrote $target"
      echo "Edit it with your favourite editor, then run: brew services start smart360"
    SH
  end

  service do
    run [opt_bin/"smart360"]
    environment_variables PATH: std_service_path_env
    # Load env from the user's config file so `brew services` picks up secrets.
    working_dir "#{Dir.home}/.config/smart360"
    keep_alive true
    log_path var/"log/smart360/smart360.log"
    error_log_path var/"log/smart360/smart360.error.log"
  end

  def caveats
    <<~EOS
      Smart 360 needs a running MongoDB instance and a populated .env file.

      1. Install & start MongoDB (one of):
           brew tap mongodb/brew && brew install mongodb-community
           brew services start mongodb-community
         …or run it via Docker:
           docker run -d --name smart360-mongo -p 27017:27017 mongo:8.0

      2. Create your config from the template:
           smart360-setup
           $EDITOR ~/.config/smart360/.env

      3. Start the service:
           brew services start smart360

         The app is then available at:
           http://localhost:8080

      Required env vars in ~/.config/smart360/.env:
        MONGODB_URI, JWT_SECRET,
        GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL,
        GEMINI_API_KEY

      See the project README for a full reference.
    EOS
  end

  test do
    assert_match "smart360", shell_output("#{bin}/smart360 --version")
  end
end
