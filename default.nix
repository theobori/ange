{ lib, buildGoModule }:
buildGoModule {
  pname = "fleur";
  version = "0.0.1";

  src = ./.;

  vendorHash = "sha256-n7N0poky1s53Z7AHPIahHJWbKZPZHJaYIfuXaqkE/Bg=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "Gopher webring based on fleur";
    homepage = "https://github.com/theobori/ange";
    license = lib.licenses.mit;
    mainProgram = "fleur";
  };
}
