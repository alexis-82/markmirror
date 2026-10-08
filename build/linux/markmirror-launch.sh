#!/bin/sh
# Starts MarkMirror, passing on the file to open (if any).
#
# With some GPU drivers (NVIDIA, virtual machines) WebKitGTK shows an empty
# or black window: disabling the DMA-BUF renderer avoids it. Set
# MARKMIRROR_GPU=1 to keep the default WebKitGTK renderer.
if [ "${MARKMIRROR_GPU:-0}" != "1" ]; then
    export WEBKIT_DISABLE_DMABUF_RENDERER=1
fi

exec /usr/local/bin/markmirror "$@"
