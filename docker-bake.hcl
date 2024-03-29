variable "DEFAULT_TAG" {
  default = "discord-bot:local"
}

// GitHub reference as defined in GitHub Actions (eg. refs/head/master)
variable "GITHUB_REF" {
  default = ""
}

target "git-ref" {
  args = {
    GIT_REF = GITHUB_REF
  }
}

target "docker-metadata-action" {
  tags = ["${DEFAULT_TAG}"]
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
    "windows/amd64",
    "darwin/arm64",
    "darwin/amd64"
  ]
}

target "image" {
  inherits = ["git-ref", "docker-metadata-action"]
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

target "lint" {
  dockerfile = "./hack/lint.Dockerfile"
  target = "lint"
  output = ["type=cacheonly"]
}

target "test" {
  dockerfile = "./hack/test.Dockerfile"
  target = "test-coverage"
  output = ["./coverage"]
}