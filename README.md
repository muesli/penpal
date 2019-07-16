# penpal
A Linux daemon to sync Wacom Bamboo devices

# Dependencies
Make sure you have [tuhi](https://github.com/tuhiproject/tuhi/) installed and
the tuhi daemon running.

You will also need librsvg and its development headers to compile `penpal`.

Pair your Bamboo device via Bluetooth and link it to tuhi, as described in their
README.

# Usage

Start `penpal` without any additional arguments and it will keep syncing new
drawings on your Bamboo device as SVG and PNG files in your current working
directory.

Alternatively you can start it with `-animation ID output.gif` to create a
rendered animation of a drawing.
