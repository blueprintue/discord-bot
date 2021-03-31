// GitHub reference as defined in GitHub Actions (eg. refs/head/master))
variable "GITHUB_REF" {
  default = ""
}

target "git-ref" {
  args = {
    GIT_REF = GITHUB_REF
  }
}

target "artifact" {
  inherits = ["git-ref"]
  target = "artifact"
  output = ["./dist"]
}

target "artifact-all" {
  inherits = ["artifact"]
  platforms = [
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "windows/amd64"
  ]
}

target "image" {
  inherits = ["git-ref"]
}

target "image-local" {
  inherits = ["image"]
  tags = ["blueprint-discord"]
  output = ["type=docker"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64",
    "windows/amd64"
  ]
}