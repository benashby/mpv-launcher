{
  description = "mpv-launcher";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      packages.${system}.default = pkgs.buildGoModule {
        pname = "mpv-launcher";
        version = "0.1.0";
        src = ./.;
        vendorHash = null;
        nativeBuildInputs = [ pkgs.pkg-config ];
        buildInputs = [ pkgs.ffmpeg ];
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          delve
          go-tools
          pkg-config
          ffmpeg
        ];
      };
    };
}
