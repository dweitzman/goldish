# goldish

A library to help with golden testing when you want many tests encoded in a
single file in a human-friendly encoding.

The reason for a custom file format was that most off-the-shelf file formats
don't seem very human-friendly when it comes to multiline strings. Sometimes
you want to be able to copy-paste big chunks of arbitrary text into a test
without worrying about line wrapping or escaping.
