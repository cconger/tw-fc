# Twitch Frame Capture
Simple application for downloading high quality twitch stream thumbnails.  It will, on a specified frequency download and store a high quality thumnail (screen capture) of a twitch stream.  This uses twitch's built in functionality for creating preview thumnails.

I created this tool to help with some machine learning training pipelines by getting a broad swath of input frames.

Each image is downloaded and given a name

`{stream_id}_{timestamp}.jpg`

Twitch updates thumnails relatively infrequetly (on the order of minutes) so we also check to see if the previous image is the same and
deduplicate if it is the same, therefore you might not have an image for every tick you specified.

Usage

`tw-fc -n 500 -p 1m -dir out`

On each tick it will query for the top live streamsby viewer count.  `-n` specifies the number of streams to capture a frame from and `-p 1m` means that it should check every minute. `-dir` specifes the relative path to output files to.

This relies on the deprecated v5 api from twitch.
