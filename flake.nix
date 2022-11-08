{
  description = "Nix flake for alertmanager-ntfy";

  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, flake-utils, nixpkgs }: 
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      rec {
        defaultPackage = with pkgs;
          buildGoModule rec {
            name = "alertmanager-ntfy";
            src = self;

            vendorSha256 = "sha256-YyyCwf8d49ah84QQJB9v7KS+9nY/0GrjsiGpK2w2AC8=";
          };
        devShell = with pkgs; mkShell {
          buildInputs = [
            go
          ];
        };
      }
    );
}
