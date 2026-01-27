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

          vendorHash = "sha256-1M0H/d6aBA8sKhVp8U9VL7mb09+q1sMdoCI2eg9/F9U=";
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
