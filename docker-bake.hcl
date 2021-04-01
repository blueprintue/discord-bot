// GitHub reference as defined in GitHub Actions (eg. refs/head/master))
variable "GITHUB_REF" {
  default = ""
}

target "git-ref" {
  args = {
    GIT_REF = GITHUB_REF
  }
}

// Special target: https://github.com/crazy-max/ghaction-docker-meta#bake-definition
target "ghaction-docker-meta" {
  tags = ["discord-bot"]
}

group "default" {
  targets = ["image-local"]
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
  inherits = ["git-ref", "ghaction-docker-meta"]
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/amd64",
    "linux/arm/v7",
    "linux/arm64"
  ]
}