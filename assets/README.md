# Assets

## Icons

### App

```bash
SRC="assets/vigor-app-icon.svg"

# iOS
IOS_DIR="app/ios/Runner/Assets.xcassets/AppIcon.appiconset"
magick -background none "$SRC" -resize 1024x1024 "$IOS_DIR/Icon-App-1024x1024@1x.png"
magick -background none "$SRC" -resize 20x20 "$IOS_DIR/Icon-App-20x20@1x.png"
magick -background none "$SRC" -resize 40x40 "$IOS_DIR/Icon-App-20x20@2x.png"
magick -background none "$SRC" -resize 60x60 "$IOS_DIR/Icon-App-20x20@3x.png"
magick -background none "$SRC" -resize 29x29 "$IOS_DIR/Icon-App-29x29@1x.png"
magick -background none "$SRC" -resize 58x58 "$IOS_DIR/Icon-App-29x29@2x.png"
magick -background none "$SRC" -resize 87x87 "$IOS_DIR/Icon-App-29x29@3x.png"
magick -background none "$SRC" -resize 40x40 "$IOS_DIR/Icon-App-40x40@1x.png"
magick -background none "$SRC" -resize 80x80 "$IOS_DIR/Icon-App-40x40@2x.png"
magick -background none "$SRC" -resize 120x120 "$IOS_DIR/Icon-App-40x40@3x.png"
magick -background none "$SRC" -resize 120x120 "$IOS_DIR/Icon-App-60x60@2x.png"
magick -background none "$SRC" -resize 180x180 "$IOS_DIR/Icon-App-60x60@3x.png"
magick -background none "$SRC" -resize 76x76 "$IOS_DIR/Icon-App-76x76@1x.png"
magick -background none "$SRC" -resize 152x152 "$IOS_DIR/Icon-App-76x76@2x.png"
magick -background none "$SRC" -resize 167x167 "$IOS_DIR/Icon-App-83.5x83.5@2x.png"

# Android
ANDROID_DIR="app/android/app/src/main/res"
magick -background none "$SRC" -resize 48x48 "$ANDROID_DIR/mipmap-mdpi/ic_launcher.png"
magick -background none "$SRC" -resize 72x72 "$ANDROID_DIR/mipmap-hdpi/ic_launcher.png"
magick -background none "$SRC" -resize 96x96 "$ANDROID_DIR/mipmap-xhdpi/ic_launcher.png"
magick -background none "$SRC" -resize 144x144 "$ANDROID_DIR/mipmap-xxhdpi/ic_launcher.png"
magick -background none "$SRC" -resize 192x192 "$ANDROID_DIR/mipmap-xxxhdpi/ic_launcher.png"

# Web
WEB_DIR="app/web"
magick -background none "$SRC" -resize 32x32 "$WEB_DIR/favicon.png"
magick -background none "$SRC" -resize 192x192 "$WEB_DIR/icons/Icon-192.png"
magick -background none "$SRC" -resize 512x512 "$WEB_DIR/icons/Icon-512.png"
magick -background none "$SRC" -resize 192x192 "$WEB_DIR/icons/Icon-maskable-192.png"
magick -background none "$SRC" -resize 512x512 "$WEB_DIR/icons/Icon-maskable-512.png"
```

### Cockpit

```bash
SRC="assets/vigor-cockpit-icon.svg"

magick -background none "$SRC" -resize 32x32 "cockpit/static/favicon.png"
magick -background none "$SRC" -resize 180x180 "cockpit/static/apple-touch-icon.png"
```
