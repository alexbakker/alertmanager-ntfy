{
  description = "Nix flake for alertmanager-ntfy";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, flake-utils, nixpkgs }:
  flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = nixpkgs.legacyPackages.${system};
    in rec {
      packages = flake-utils.lib.flattenTree rec {
        default = alertmanager-ntfy;
        alertmanager-ntfy = with pkgs; buildGoModule rec {
          name = "alertmanager-ntfy";
          src = self;

          vendorSha256 = "sha256-wQXS+QCkR6AIli6Dzv2fYlQiweHUVJWjqLQulIlddRQ=";
        };
      };
      devShells.default = with pkgs; mkShell {
        buildInputs = [
          go
        ];
      };
      nixosModules.default = ({ pkgs, ... }: {
        imports = [ ./module.nix ];
        nixpkgs.overlays = [
          (_self: _super: {
            alertmanager-ntfy = self.packages.${pkgs.system}.alertmanager-ntfy;
          })
        ];
      });
    }
  );
}
