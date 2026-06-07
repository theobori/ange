{ lib, buildGoModule }:
buildGoModule {
  pname = "fleur";
  version = "0.0.1";

  src = ./.;

  vendorHash = "sha256-w50DtGiFVP3g59927fgOoNNIBzEm0yv1q+GXk4G0PKI=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "Gopher webring based on fleur ";
    homepage = "https://github.com/theobori/ange";
    license = lib.licenses.mit;
    mainProgram = "fleur";
  };
}
