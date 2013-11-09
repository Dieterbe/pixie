An opiniated photo app for minimalists.

* you have one or more directories that contain the pictures you pull from your devices, just the way they are.
  (no renames, file modifications, or deletes that don't have corresponding deletes on the device)
  this allows for trivial syncing/backups.
* any modifications done to pictures are saved in files in a separate directory, but with the same name.
* you assign tags to pictures based on their contents or desired goal (i.e. whether it should go in a certain album or not).
  Pixies makes this as qucik and easy as possible.
* Photo albums are deterministically generated from the aforementioned directories
  and the tag metadata.  They can be erased and regenerated at any point

It is conceivable that with the advent of new technology some of these things may change, however a simple file layout
that works on all posix file systems is very valuable (i.e. just directories, files and symlinks).
