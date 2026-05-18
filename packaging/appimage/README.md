# ZenithLens AppImage Layout

This directory contains the static AppImage entrypoint metadata.

The eventual AppDir should contain:

```text
AppDir/
├── AppRun
├── zenithlens.desktop
├── zenithlens.png or zenithlens.svg
└── usr/
    └── bin/
        └── zenithlens
```

`linuxdeploy` or an equivalent AppImage packaging tool can turn that AppDir into
the final `.AppImage` artifact.
