# dcmd

dcmd is a extensible discord command system based on interfaces.

It's very much work in progress at the moment, if you start using it now you have to be okay with things changing and the fact that you will find bugs.

## Features:

For now look in the example folder. Still planning things out.

## TODO:

 - [ ] Full test coverage (See below for info on progress)
 - [ ] Only build middleware chains once?
 - [x] Standard Help generator
 - Flags
      + [ ] FlagHideFromHelp
      + [ ] FlagRunInDM
      + [ ] FlagIgnoreMentions
      + [ ] FlagIgnorePrefix

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

### Time of day example

```go

func main() {
      // Create a new command system
      system := dcmd.NewStandardSystem("[")

      // Add the time of day command to the root container of the system
      system.Root.AddCommand(&CmdTimeOfDay{Format: time.RFC822}, "Time", "t")

      // Create the discordgo session
      session, err := discordgo.New(os.Getenv("DG_TOKEN"))
      if err != nil {
            log.Fatal("Failed setting up session:", err)
      }

      // Add the command system handler to discordgo
      session.AddHandler(system.HandleMessageCreate)

      err = session.Open()
      if err != nil {
            log.Fatal("Failed opening gateway connection:", err)
      }
      log.Println("Running, Ctrl-c to stop.")
      select {}
}

type CmdTimeOfDay struct {
      Format string
}

// Descriptions should return a short description (used in the overall help overiview) and one long descriptions for targetted help
func (t *CmdTimeOfDay) Descriptions(data *dcmd.Data) (string, string) {
      return "Responds with the current time in utc", ""
}

// Run implements the dcmd.Cmd interface and gets called when the command is invoked
func (t *CmdTimeOfDay) Run(data *dcmd.Data) (interface{}, error) {
      return time.Now().UTC().Format(t.Format), nil
}

// Compilie time assertions, will be not compiled unless CmdTimeOfDay implements these interfaces
var _ dcmd.Cmd = (*CmdTimeOfDay)(nil)
var _ dcmd.CmdWithDescriptions = (*CmdTimeOfDay)(nil)

```
