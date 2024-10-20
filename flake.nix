{
  description = "Find upcoming NC DMV appointments";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs, ... }: let
    # No chromium on darwin: https://github.com/NixOS/nixpkgs/issues/247855
    supportedSystems = [ "x86_64-linux" "aarch64-linux" ];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    pname = "ncdmv";
    owner = "aksiksi";
    version = "0.1.0";
  in {
    packages = forAllSystems (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        default = pkgs.buildGoModule {
          inherit pname;
          inherit version;
          src = ./.;
          vendorHash = "sha256-mW+0FAVnBP5ISpQgBcAIUmvayMeu0dLfo6tcpHYoMfs=";
          buildInputs = [
            pkgs.sqlite
          ];
        };
        docker = pkgs.dockerTools.streamLayeredImage {
          name = "ghcr.io/aksiksi/ncdmv";
          tag = "latest";
          # Use commit date to ensure image creation date is reproducible.
          created = builtins.substring 0 8 self.lastModifiedDate;
          contents = [
            # Required for Discord webhook over HTTPS.
            pkgs.cacert
            pkgs.chromium
            self.outputs.packages.${system}.default
          ];
          extraCommands = ''
            # Required by Chromium
            mkdir tmp
          '';
          config = {
            Entrypoint = [ "ncdmv" ];
            Volumes = {
              # DB storage
              "/config" = null;
            };
          };
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
            pkgs.chromium
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

