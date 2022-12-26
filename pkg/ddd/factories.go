package ddd

type CreateCommandHandler func() (CommandHandler, error)
type CreateEventHandler func() (EventHandler, error)
