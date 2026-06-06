{ lib, buildGoModule }:
buildGoModule {
  pname = "fleur";
  version = "0.0.1";

  src = ./.;

  vendorHash = "sha256-jEoCS9DZcdp0g8x7OSDucqRucwF0bMdkBJh3zZ2g96c=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "KISS RFC1436 Gopher application framework and CLI ";
    homepage = "https://github.com/theobori/fleur";
    license = lib.licenses.mit;
    mainProgram = "fleur";
  };
}
