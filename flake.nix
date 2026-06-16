{
  description = "Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          nativeBuildInputs = [
            # general command-line/development tools
            pkgs.ast-grep
            pkgs.delta
            pkgs.fd
            pkgs.file
            pkgs.gh
            pkgs.ghostscript
            pkgs.git
            pkgs.glow
            pkgs.htop
            pkgs.hyperfine
            pkgs.jq
            pkgs.kitty
            pkgs.lsof
            pkgs.ripgrep
            pkgs.stgit
            pkgs.strace
            pkgs.tree
            pkgs.watchexec

            # go programming specific
            pkgs.gcc
            pkgs.delve
            pkgs.gnumake
            pkgs.go_latest
            pkgs.golangci-lint
            pkgs.golangci-lint-langserver
            pkgs.gomodifytags
            pkgs.gopls
            pkgs.gotests
            pkgs.graphviz
            pkgs.impl

            # Red hat's auth uses Kerberos throughout.
            pkgs.libkrb5
          ];

          shellHook = ''
            echo "🚀 Entering RedHat Dev Shell: Go($(go version))"
          '';
        };
      }
    );
}
