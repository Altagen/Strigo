#!/bin/bash

# Script to upload JDKs to Nexus with proper path conventions

DOWNLOAD_DIR="/tmp/strigo-multi-jdk-downloads"
NEXUS_URL="http://localhost:8081"
NEXUS_USER="admin"
NEXUS_PASSWORD="admin"
REPO="raw"

echo "========================================="
echo "Uploading JDKs to Nexus"
echo "========================================="
echo ""

# Helper function to extract version from tar.gz
get_version() {
    local file=$1
    # Extract just the directory name from the tar
    tar -tzf "$file" | head -1 | cut -d/ -f1
}

# Upload function
upload_jdk() {
    local file=$1
    local path=$2
    local filename=$3

    echo "Uploading: $filename to $path"
    curl -u "$NEXUS_USER:$NEXUS_PASSWORD" -X POST \
        "$NEXUS_URL/service/rest/v1/components?repository=$REPO" \
        -F "raw.directory=/$path" \
        -F "raw.asset1=@$DOWNLOAD_DIR/$file" \
        -F "raw.asset1.filename=$filename" 2>&1 | grep -v "^  "
    echo "✅ $filename uploaded"
    echo ""
}

cd "$DOWNLOAD_DIR"

# Temurin JDK 8 - version 8.0.432_6
echo "[1/10] Uploading Temurin JDK 8..."
upload_jdk "temurin-jdk8.tar.gz" "jdk/adoptium/temurin/8/8.0.432_6" "OpenJDK8U-jdk_x64_linux_hotspot_8u432b06.tar.gz"

# Temurin JDK 17 - version 17.0.13_11
echo "[2/10] Uploading Temurin JDK 17..."
upload_jdk "temurin-jdk17.tar.gz" "jdk/adoptium/temurin/17/17.0.13_11" "OpenJDK17U-jdk_x64_linux_hotspot_17.0.13_11.tar.gz"

# Temurin JDK 21 - version 21.0.5_11
echo "[3/10] Uploading Temurin JDK 21..."
upload_jdk "temurin-jdk21.tar.gz" "jdk/adoptium/temurin/21/21.0.5_11" "OpenJDK21U-jdk_x64_linux_hotspot_21.0.5_11.tar.gz"

# Corretto JDK 11 - version 11.0.25.9.1
echo "[4/10] Uploading Corretto JDK 11..."
upload_jdk "corretto-jdk11.tar.gz" "jdk/amazon/corretto/11/11.0.25.9.1" "amazon-corretto-11.0.25.9.1-linux-x64.tar.gz"

# Corretto JDK 17 - version 17.0.13.11.1
echo "[5/10] Uploading Corretto JDK 17..."
upload_jdk "corretto-jdk17.tar.gz" "jdk/amazon/corretto/17/17.0.13.11.1" "amazon-corretto-17.0.13.11.1-linux-x64.tar.gz"

# Corretto JDK 21 - version 21.0.5.11.1
echo "[6/10] Uploading Corretto JDK 21..."
upload_jdk "corretto-jdk21.tar.gz" "jdk/amazon/corretto/21/21.0.5.11.1" "amazon-corretto-21.0.5.11.1-linux-x64.tar.gz"

# Azul Zulu JDK 8 - version 8.0.432
echo "[7/10] Uploading Azul Zulu JDK 8..."
upload_jdk "zulu-jdk8.tar.gz" "jdk/azul/zulu/8/8.0.432" "zulu8.82.0.21-ca-jdk8.0.432-linux_x64.tar.gz"

# Azul Zulu JDK 11 - version 11.0.25
echo "[8/10] Uploading Azul Zulu JDK 11..."
upload_jdk "zulu-jdk11.tar.gz" "jdk/azul/zulu/11/11.0.25" "zulu11.76.21-ca-jdk11.0.25-linux_x64.tar.gz"

# Azul Zulu JDK 21 - version 21.0.5
echo "[9/10] Uploading Azul Zulu JDK 21..."
upload_jdk "zulu-jdk21.tar.gz" "jdk/azul/zulu/21/21.0.5" "zulu21.38.21-ca-jdk21.0.5-linux_x64.tar.gz"

# Mandrel 23 - version 23.1.5.0
echo "[10/10] Uploading Mandrel 23..."
upload_jdk "mandrel-23.tar.gz" "jdk/graalvm/mandrel/23/23.1.5.0" "mandrel-java23-linux-amd64-23.1.5.0-Final.tar.gz"

echo ""
echo "========================================="
echo "✅ All JDKs uploaded to Nexus!"
echo "========================================="
