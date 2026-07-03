{ pkgs, ... }: {
  packages = [ pkgs.go ];
  containers.http-destination = {
    name = "http-destination";
    registry = "ghcr.io/junctionio/";
    # devenv's containers.<name>.copyToRoot content lands under /env (the
    # container's homeDir/workingDir), not container root - so this must
    # be /env/bin/worker, not /bin/worker. Confirmed by inspecting a
    # locally-built image with `devenv container copy --registry
    # docker-daemon:` and `docker run --entrypoint /bin/sh ... ls /env/bin`.
    entrypoint = [ "/env/bin/worker" ];
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
