# dcmd

dcmd is a extensible discord command system based on interfaces.

It's very much work in progress atm, check out the example folder for examples
.
## Features:

For now look in the example folder. Still planning things out.

## TODO:

 - [ ] Full test coverage (See below for info on progress)
 - [ ] Only build middleware chains once?
 - [ ] Standard Help generator
 - [ ] Customizable error handling

## Test Coverage:

 - Argument parsers
     + [x] int
     + [x] float
     + [ ] string
     + [ ] user
     + [ ] 
 - System
     + [x] FindPrefix
     + [ ] HandleResponse
     + [ ] 
 - Container
     + [ ] Middleware chaining
     + [ ] Add/Remove middleware
     + [ ] Command searching