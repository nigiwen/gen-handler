#!/bin/bash

# gen-handler multi-platform packaging script

set -euo pipefail

VERSION="${1:-dev}"
APP_NAME="gen-handler"
BUILD_DIR="dist"

echo "🚀 开始编译 $APP_NAME v$VERSION"
echo ""

# 清理旧的构建文件
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# 定义要编译的平台
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# 编译函数
build() {
    local GOOS=$1
    local GOARCH=$2
    local OUTPUT_NAME=$APP_NAME
    
    # Windows 平台需要 .exe 后缀
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    local OUTPUT_PATH="$BUILD_DIR/${GOOS}_${GOARCH}/${OUTPUT_NAME}"
    
    echo "📦 编译 $GOOS/$GOARCH..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.version=$VERSION" \
        -o "$OUTPUT_PATH" \
        .
    
    # 创建压缩包
    local ARCHIVE_NAME="${APP_NAME}_${VERSION}_${GOOS}_${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        cd "$BUILD_DIR/${GOOS}_${GOARCH}"
        # 尝试使用 zip，如果没有则使用 tar
        if command -v zip >/dev/null 2>&1; then
            zip -q "../../${BUILD_DIR}/${ARCHIVE_NAME}.zip" "$OUTPUT_NAME"
        else
            tar -czf "../../${BUILD_DIR}/${ARCHIVE_NAME}.tar.gz" "$OUTPUT_NAME"
        fi
        cd ../..
    else
        cd "$BUILD_DIR/${GOOS}_${GOARCH}"
        tar -czf "../../${BUILD_DIR}/${ARCHIVE_NAME}.tar.gz" "$OUTPUT_NAME"
        cd ../..
    fi
    
    echo "✅ $GOOS/$GOARCH 编译完成: $OUTPUT_PATH"
}

# 编译所有平台
for PLATFORM in "${PLATFORMS[@]}"; do
    PLATFORM_SPLIT=(${PLATFORM//\// })
    GOOS=${PLATFORM_SPLIT[0]}
    GOARCH=${PLATFORM_SPLIT[1]}
    build $GOOS $GOARCH
done

echo ""
echo "🎉 所有平台编译完成！"
echo ""
echo "📦 构建产物："
ls -lh $BUILD_DIR/*.{zip,tar.gz} 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'
echo ""
echo "📁 构建目录: $BUILD_DIR/"
echo ""
echo "💡 发布到 GitHub Release 时，上传以下文件："
echo "   - ${APP_NAME}_${VERSION}_linux_amd64.tar.gz"
echo "   - ${APP_NAME}_${VERSION}_linux_arm64.tar.gz"
echo "   - ${APP_NAME}_${VERSION}_darwin_amd64.tar.gz"
echo "   - ${APP_NAME}_${VERSION}_darwin_arm64.tar.gz"
echo "   - ${APP_NAME}_${VERSION}_windows_amd64.zip"
