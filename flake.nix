{
  description = "Find upcoming NC DMV appointments";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs, ... }: let
    supportedSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    pname = "ncdmv";
    owner = "aksiksi";
    version = "0.1.0";
  in {
    packages = forAllSystems (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        selfPackages = self.outputs.packages.${system};
        entrypoint = pkgs.writeScript "docker-entrypoint.sh" ''
          #!${pkgs.stdenv.shell}
          exec ${selfPackages.default}/bin/ncdmv \
            -t "$NCDMV_APPT_TYPE" \
            -l "$NCDMV_LOCATIONS" \
            -d "$NCDMV_DATABASE_PATH" \
            -w "$NCDMV_DISCORD_WEBHOOK" \
            --timeout "$NCDMV_TIMEOUT" \
            --interval "$NCDMV_INTERVAL" \
            --notify-unavailable=$NCDMV_NOTIFY_UNAVAILABLE \
            --disable-gpu=$NCDMV_DISABLE_GPU \
            --debug=$NCDMV_DEBUG \
            --debug-chrome=$NCDMV_DEBUG_CHROME
        '';
        imageArch =
          if pkgs.lib.strings.hasPrefix "x86_64" system
          then "amd64"
          else "arm64";
      in {
        default = pkgs.buildGoModule {
          inherit pname;
          inherit version;
          src = ./.;
          vendorHash = "sha256-IhB0urJVKYbIIFSoiuR92/KlTe+gxOcIuLEZrZYyxO0=";
          buildInputs = [
            pkgs.sqlite
          ];
        };
        # https://hub.docker.com/r/chromedp/headless-shell
        chrome-headless = pkgs.dockerTools.pullImage {
          imageName = "docker.io/chromedp/headless-shell";
          imageDigest = "sha256:bbe8b1153719af55adcc5d197e490ed7cb146468f09ce939cd53b7dd280c951b";
          finalImageTag = "latest";
          sha256 = "sha256-mS3kdcGc4v3KPLfN3WCvTP39lALhSCVhfaWLQLgtoNE=";
          os = "linux";
          arch = imageArch;
        };
        docker = pkgs.dockerTools.streamLayeredImage {
          name = "ghcr.io/aksiksi/ncdmv";
          tag = "latest";
          fromImage = selfPackages.chrome-headless;
          # Use commit date to ensure image creation date is reproducible.
          created = builtins.substring 0 8 self.lastModifiedDate;
          architecture = imageArch;
          config = {
            Entrypoint = [ entrypoint ];
            Volumes = {
              # DB storage
              "/config" = {};
            };
            Env = [
              "NCDMV_APPT_TYPE="
              "NCDMV_LOCATIONS="
              "NCDMV_DATABASE_PATH=/config/ncdmv.db"
              "NCDMV_DISCORD_WEBHOOK="
              "NCDMV_TIMEOUT=5m"
              "NCDMV_INTERVAL=5m"
              "NCDMV_DISABLE_GPU=false"
              "NCDMV_NOTIFY_UNAVAILABLE=true"
              "NCDMV_DEBUG=false"
              "NCDMV_DEBUG_CHROME=false"
            ];
          };
          # Default is 100, so this ensures this image gets its own layer(s)
          # after being merged with the base image.
          maxLayers = 120;
        };
      }
    );

    # Development shell
    # nix develop
    devShells = forAllSystems (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        default = pkgs.mkShell {
          buildInputs = [
            pkgs.sqlite
          ];
          packages = [
            pkgs.go
            pkgs.gopls
          ];
        };
      }
    );
  };
}

