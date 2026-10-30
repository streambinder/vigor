# Exercise player page

Single-exercise Three.js page for the app exercise animation player.
Flutter loads `player.html` in a WebView and hands it the animation
document and the model bytes as base64 chunks (see `window.startPlayer`
in the page). The model itself is not bundled: it comes from the
backend API, like the animation document.

## Vendored libraries

- `three.min.js`: Three.js r128, UMD build, MIT licence. Same pair
  the QA harness used to verify the animation documents.
- `GLTFLoader.js`: the matching UMD GLTFLoader for r128, MIT
  licence. Loaded as a plain script; it registers
  `THREE.GLTFLoader`.

Do not upgrade these without re-verifying the player against the
gallery motion: the pose contract (rest-relative quaternions, root
track at the pelvis, hand stripping) was verified with this pair.
