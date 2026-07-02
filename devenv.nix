{ pkgs, ... }: {
  packages = [ pkgs.go ];
  containers.http-destination = {
    name = "http-destination";
    entrypoint = [ "/bin/worker" ];
    copyToRoot = pkgs.buildEnv {
      name = "http-destination-root";
      paths = [
        pkgs.cacert
        (pkgs.runCommand "worker-bin" { } ''
          mkdir -p $out/bin
          cp ${./worker} $out/bin/worker
          chmod +x $out/bin/worker
        '')
      ];
    };
  };

  enterShell = ''
    set +x
    set -a; [ -f .env ] && source .env; set +a
  '';
  processes.worker.exec = "go run .";
}
