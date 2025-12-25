#!/bin/bash

# Script to download various JDK versions for testing

DOWNLOAD_DIR="/tmp/strigo-multi-jdk-downloads"
mkdir -p "$DOWNLOAD_DIR"
cd "$DOWNLOAD_DIR"

echo "========================================="
echo "Downloading Multiple JDK Versions"
echo "========================================="
echo ""

# Temurin JDK 8 (Java 8 - jre/lib/security/cacerts path)
echo "[1/10] Downloading Temurin JDK 8..."
curl -L "https://api.adoptium.net/v3/binary/latest/8/ga/linux/x64/jdk/hotspot/normal/eclipse?project=jdk" \
  -o temurin-jdk8.tar.gz &

# Temurin JDK 17 (Java 17 - lib/security/cacerts path)
echo "[2/10] Downloading Temurin JDK 17..."
curl -L "https://api.adoptium.net/v3/binary/latest/17/ga/linux/x64/jdk/hotspot/normal/eclipse?project=jdk" \
  -o temurin-jdk17.tar.gz &

# Temurin JDK 21 (Java 21 - lib/security/cacerts path)
echo "[3/10] Downloading Temurin JDK 21..."
curl -L "https://api.adoptium.net/v3/binary/latest/21/ga/linux/x64/jdk/hotspot/normal/eclipse?project=jdk" \
  -o temurin-jdk21.tar.gz &

# Corretto JDK 11
echo "[4/10] Downloading Corretto JDK 11..."
curl -L "https://corretto.aws/downloads/latest/amazon-corretto-11-x64-linux-jdk.tar.gz" \
  -o corretto-jdk11.tar.gz &

# Corretto JDK 17
echo "[5/10] Downloading Corretto JDK 17..."
curl -L "https://corretto.aws/downloads/latest/amazon-corretto-17-x64-linux-jdk.tar.gz" \
  -o corretto-jdk17.tar.gz &

# Corretto JDK 21
echo "[6/10] Downloading Corretto JDK 21..."
curl -L "https://corretto.aws/downloads/latest/amazon-corretto-21-x64-linux-jdk.tar.gz" \
  -o corretto-jdk21.tar.gz &

# Azul Zulu JDK 8
echo "[7/10] Downloading Azul Zulu JDK 8..."
curl -L "https://cdn.azul.com/zulu/bin/zulu8.82.0.21-ca-jdk8.0.432-linux_x64.tar.gz" \
  -o zulu-jdk8.tar.gz &

# Azul Zulu JDK 11
echo "[8/10] Downloading Azul Zulu JDK 11..."
curl -L "https://cdn.azul.com/zulu/bin/zulu11.76.21-ca-jdk11.0.25-linux_x64.tar.gz" \
  -o zulu-jdk11.tar.gz &

# Azul Zulu JDK 21
echo "[9/10] Downloading Azul Zulu JDK 21..."
curl -L "https://cdn.azul.com/zulu/bin/zulu21.38.21-ca-jdk21.0.5-linux_x64.tar.gz" \
  -o zulu-jdk21.tar.gz &

# Mandrel 23 (GraalVM-based)
echo "[10/10] Downloading Mandrel 23..."
curl -L "https://github.com/graalvm/mandrel/releases/download/mandrel-23.1.5.0-Final/mandrel-java23-linux-amd64-23.1.5.0-Final.tar.gz" \
  -o mandrel-23.tar.gz &

echo ""
echo "Waiting for downloads to complete..."
wait

echo ""
echo "========================================="
echo "Download Summary"
echo "========================================="
ls -lh *.tar.gz

echo ""
echo "✅ All JDKs downloaded successfully!"
echo "📂 Location: $DOWNLOAD_DIR"
