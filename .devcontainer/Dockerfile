# syntax=docker/dockerfile:1

###############################################################################
#  Integrated dev container for the "chat" project
#
#  Toolchains bundled in this single image:
#    - Go        : matches go.mod  -> go 1.23.x        (ARG GO_VERSION)
#    - Node.js   : 20.x LTS        -> matches .github/workflows/deploy.yml
#    - Java      : OpenJDK 17      -> required by Android Gradle Plugin 8.2
#    - Android   : cmdline-tools + platform-tools (adb) +
#                  platforms;android-34 + build-tools;34.0.0
#
#  adb does NOT talk to USB directly. It forwards every command to the ADB
#  server running on the HOST PC through the env var ADB_SERVER_SOCKET
#  (configured in devcontainer.json).
#
#  Isolation: no host folder (e.g. ~/.ssh) is bind-mounted into the container,
#  so tools/agents inside the container cannot read the host's private keys.
###############################################################################
FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

# Install everything as root. The running container uses the non-root "vscode"
# user that already exists in this base image (see remoteUser in devcontainer.json).
USER root
ENV DEBIAN_FRONTEND=noninteractive

# ---------------------------------------------------------------------------
# 1) Base tooling + Java 17
# ---------------------------------------------------------------------------
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        wget \
        unzip \
        zip \
        git \
        build-essential \
        openjdk-17-jdk \
    && rm -rf /var/lib/apt/lists/*

ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
ENV PATH="${JAVA_HOME}/bin:${PATH}"

# ---------------------------------------------------------------------------
# 2) Go  (version is overridable: --build-arg GO_VERSION=1.23.4)
# ---------------------------------------------------------------------------
ARG GO_VERSION=1.23.4
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
RUN wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm -f /tmp/go.tar.gz \
    && mkdir -p "${GOPATH}/bin" \
    && chown -R vscode:vscode "${GOPATH}"
ENV PATH="${GOROOT}/bin:${GOPATH}/bin:${PATH}"

# ---------------------------------------------------------------------------
# 3) Node.js 20 LTS + npm
# ---------------------------------------------------------------------------
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------------------
# 4) Android SDK: cmdline-tools / platform-tools (adb) / android-34 / build-tools
# ---------------------------------------------------------------------------
ARG ANDROID_CMDLINE_TOOLS=11076708
ENV ANDROID_HOME=/opt/android-sdk
ENV ANDROID_SDK_ROOT=${ANDROID_HOME}
ENV PATH="${ANDROID_HOME}/cmdline-tools/latest/bin:${ANDROID_HOME}/platform-tools:${ANDROID_HOME}/build-tools/34.0.0:${PATH}"

RUN mkdir -p "${ANDROID_HOME}/cmdline-tools" \
    && wget -q "https://dl.google.com/android/repository/commandlinetools-linux-${ANDROID_CMDLINE_TOOLS}_latest.zip" -O /tmp/cmdline-tools.zip \
    && unzip -q /tmp/cmdline-tools.zip -d /tmp/cmdline-tools-unzip \
    && mv /tmp/cmdline-tools-unzip/cmdline-tools "${ANDROID_HOME}/cmdline-tools/latest" \
    && rm -rf /tmp/cmdline-tools.zip /tmp/cmdline-tools-unzip

# Accept the SDK licenses (ignore the "already accepted" non-zero exit), then
# install only the packages this project actually needs.
RUN yes | sdkmanager --licenses > /dev/null 2>&1 || true
RUN sdkmanager --install "platform-tools" "platforms;android-34" "build-tools;34.0.0"

# The container runs as "vscode", so it must own the SDK tree.
RUN chown -R vscode:vscode "${ANDROID_HOME}"

# ---------------------------------------------------------------------------
# Hand control back to the non-root user for the running container.
# ---------------------------------------------------------------------------
USER vscode
