# sgo - Setup Go Programming Language Tool

This tool was started to make it simple to download and "install" Go. There are
prepared binaries packed up in .zip and .tar.gz files out there that are
portable and can just be extracted to make Go compiler and tools pretty much
ready to be used.

New updates are released every so often, and this is where this tool comes in,
trying to ease the process to upgrade. Repositories might not be as quick to
have the latest version of Go available as quickly as it's released.

Right now this tool just downloads and extracts, but eventually this might be
made even smarter to setup and prepare environment variables and whatever else
that would be great to do. This way, not only can you easily upgrade, you will
also be able to change to other versions easily.

This has only been tested in Linux currently, but should work fine in Windows
and Mac too.

![Sgo](https://github.com/Chillance/sgo/blob/master/sgo.gif)

### Install
This is one way of doing it, that should work:

1. Clone the repo.
2. Set GOPATH.
3. Run "go get"
4. Run "go run sgo.go"

### Usage

Start by just invoking ./sgo in the terminal. It will start to download the
single page download Go web site, and when done show the parsed data in
different views. You might have to resize your terminal to see everything,
as it's not smart enough to adapt perfectly to the size of the terminal yet.

Navigate up and down with the arrow keys. Tab changes between the Go versions
view and Go files table list view. "q"/"Q" or CTRL+C exists the program.

Note also that the download will resume if it doesn't finish and you download
the same file again. This will also be shown as a big spike in the sparkline
graph.
