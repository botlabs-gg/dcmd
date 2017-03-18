# dcmd

dcmd is a extensible discord command system based on interfaces.

It's very much work in progress at the moment, if you start using it now you have to be okay with things changing and the fact that you will find bugs.

## Features:

For now look in the example folder. Still planning things out.

## TODO:

 - [ ] Full test coverage (See below for info on progress)
 - [ ] Only build middleware chains once?
 - [ ] Standard Help generator

## Test Coverage:

 - Argument parsers
      + [x] int
      + [x] float
      + [ ] string
      + [ ] user
      + [x] Full line argdef parsing
      + [ ] Full line switch parsing
 - System
      + [x] FindPrefix
      + [ ] HandleResponse
 - Container
      + [ ] Middleware chaining
      + [ ] Add/Remove middleware
      + [ ] Command searching
      + [ ] 
 - Other
      + [ ] Help
